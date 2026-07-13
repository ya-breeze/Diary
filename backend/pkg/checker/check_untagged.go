package checker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/ya-breeze/diary.be/pkg/ai"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

// maxBackfillPerRun bounds how many AI suggestion calls the untagged check makes
// per family in a single run, so a family with a long untagged history doesn't
// trigger a huge batch every sweep. Remaining days are picked up on later runs.
const maxBackfillPerRun = 25

// UntaggedCheck finds untagged or stale days for families that have enabled the
// AI tagging backfill, and surfaces them through the health-issues flow. It is a
// no-op unless a suggester is available and the family has opted in.
type UntaggedCheck struct {
	Suggester ai.Suggester
}

func (UntaggedCheck) Name() string { return "untagged" }

func (c UntaggedCheck) Run(db database.Storage, cfg *config.Config, logger *slog.Logger) ([]Issue, error) {
	if c.Suggester == nil || !c.Suggester.Enabled() {
		return nil, nil // AI unavailable → no backfill
	}

	users, err := db.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}

	var issues []Issue
	seen := map[uuid.UUID]bool{}
	for _, user := range users {
		familyID := user.FamilyID
		if seen[familyID] {
			continue
		}
		seen[familyID] = true

		family, err := db.GetFamily(familyID)
		if err != nil {
			return nil, fmt.Errorf("getting family %s: %w", familyID, err)
		}
		if !family.AITaggingEnabled || !family.AITaggingBackfill {
			continue
		}

		famIssues, err := c.runForFamily(db, cfg, logger, family)
		if err != nil {
			return nil, err
		}
		issues = append(issues, famIssues...)
	}
	return issues, nil
}

func (c UntaggedCheck) runForFamily(
	db database.Storage, cfg *config.Config, logger *slog.Logger, family *models.Family,
) ([]Issue, error) {
	familyID := family.ID
	knownTags, err := db.GetDistinctTags(familyID)
	if err != nil {
		return nil, fmt.Errorf("getting tags for family %s: %w", familyID, err)
	}
	items, _, err := db.GetItems(familyID, database.SearchParams{})
	if err != nil {
		return nil, fmt.Errorf("getting items for family %s: %w", familyID, err)
	}

	issues := make([]Issue, 0)
	generated := 0
	hitCap := false
	analysisFailed := false
	for _, item := range items {
		untagged := len(item.Tags) == 0
		stale := item.TagsSourceHash != utils.ComputeTagsSourceHash(item.Title, item.Body)

		// Skip days we've already processed for this content that have nothing
		// pending — whether they got tagged or the user dismissed every suggestion.
		// A never-processed day has an empty hash, so it reads as stale.
		if !stale && len(item.PendingTags) == 0 {
			continue
		}

		// Already-staged suggestions (content unchanged): surface without a new call.
		if len(item.PendingTags) > 0 && !stale {
			issues = append(issues, c.reviewIssue(familyID, item.Date, len(item.PendingTags)))
			continue
		}

		// Fresh analysis is the one-time backfill's job only: once the family's
		// backfill has completed, un-analyzed days are left alone (completion
		// stops new model calls, not the review of already-staged pending).
		if family.AITaggingBackfillDone {
			continue
		}

		if generated >= maxBackfillPerRun {
			hitCap = true
			break // bound AI cost; remaining days handled next run
		}
		generated++

		issue, ok, analyzed := c.processItem(db, cfg, logger, family, item, knownTags, untagged)
		if !analyzed {
			analysisFailed = true
		}
		if ok {
			issues = append(issues, issue)
		}
	}

	// The backfill is exhausted when a full scan needed no further model calls:
	// it did not stop at the per-run cap and no analysis failed (a failed day is
	// still un-analyzed and must be retried next run). Flip the flag so no
	// automatic analysis runs for this family again (until backfill is
	// re-toggled off->on, which resets it).
	if !family.AITaggingBackfillDone && !hitCap && !analysisFailed {
		if err := db.SetFamilyBackfillDone(familyID, true); err != nil {
			logger.Error("Untagged check: failed to mark backfill done", "familyID", familyID, "error", err)
		} else {
			logger.Info("Untagged check: backfill complete", "familyID", familyID)
		}
	}
	return issues, nil
}

// processItem generates suggestions for one candidate day. Under auto mode it
// applies confident tags immediately (resolving the day, no issue). Otherwise it
// stages the suggestions as pending and returns a non-fixable "review" issue.
// ok=false when there's nothing to surface (resolved, empty, or error).
// analyzed reports whether the day was fully processed and marked (hash
// stamped); it is false on any failure so the day is retried next run and the
// backfill is not marked complete over it.
func (c UntaggedCheck) processItem(
	db database.Storage, cfg *config.Config, logger *slog.Logger,
	family *models.Family, item *models.Item, knownTags []string, untagged bool,
) (Issue, bool, bool) {
	familyID := family.ID
	var images []ai.ImageAsset
	if family.AITaggingUseImages {
		images = ai.LoadImageAssets(item.Body, cfg.DataPath, familyID.String())
	}
	if family.AITaggingUseVideo {
		images = append(images, ai.LoadVideoKeyframes(item.Body, cfg.DataPath, familyID.String(), logger, ai.MaxImages-len(images))...)
	}
	suggestions, err := c.Suggester.SuggestTags(context.Background(), item.Title, item.Body, images, knownTags)
	if err != nil {
		logger.Error("Untagged check: suggestion failed", "familyID", familyID, "date", item.Date, "error", err)
		return Issue{}, false, false
	}
	names, confident := ai.Partition(suggestions, item.Tags, cfg.AITaggingThreshold)
	if len(names) == 0 {
		// Nothing to suggest is still a completed analysis: stamp the hash (via an
		// empty pending write) so the one-time backfill never revisits this day.
		if err := db.SetPendingTags(familyID, item.Date, nil); err != nil {
			logger.Error("Untagged check: failed to mark analyzed", "familyID", familyID, "date", item.Date, "error", err)
			return Issue{}, false, false
		}
		return Issue{}, false, true
	}

	// Auto mode: apply confident tags to an untagged day right away — no manual
	// "fix" step. Low-confidence suggestions are still staged for review.
	if family.AITaggingAuto && untagged && len(confident) > 0 {
		if err := db.AddConfirmedTags(familyID, item.Date, confident); err != nil {
			logger.Error("Untagged check: auto-apply failed", "familyID", familyID, "date", item.Date, "error", err)
			return Issue{}, false, false
		}
		logger.Info("Untagged check: auto-applied confident tags", "familyID", familyID, "date", item.Date, "tags", confident)
		uncertain := subtractStrings(names, confident)
		if len(uncertain) > 0 {
			if err := db.SetPendingTags(familyID, item.Date, uncertain); err != nil {
				logger.Error("Untagged check: failed to stage pending tags", "familyID", familyID, "date", item.Date, "error", err)
				return Issue{}, false, false
			}
			return c.reviewIssue(familyID, item.Date, len(uncertain)), true, true
		}
		return Issue{}, false, true
	}

	// Non-auto, or auto with no confident suggestions: stage all for review.
	if err := db.SetPendingTags(familyID, item.Date, names); err != nil {
		logger.Error("Untagged check: failed to stage pending tags", "familyID", familyID, "date", item.Date, "error", err)
		return Issue{}, false, false
	}
	return c.reviewIssue(familyID, item.Date, len(names)), true, true
}

func subtractStrings(all, exclude []string) []string {
	if len(exclude) == 0 {
		return all
	}
	ex := make(map[string]struct{}, len(exclude))
	for _, s := range exclude {
		ex[s] = struct{}{}
	}
	var out []string
	for _, s := range all {
		if _, ok := ex[s]; !ok {
			out = append(out, s)
		}
	}
	return out
}

// reviewIssue is a non-fixable issue pointing the user at a day to review.
func (UntaggedCheck) reviewIssue(familyID uuid.UUID, date string, n int) Issue {
	return Issue{
		Check:    "untagged",
		FamilyID: familyID.String(),
		Path:     date,
		Message:  fmt.Sprintf("%d suggested tag(s) to review", n),
		Fixable:  false,
	}
}
