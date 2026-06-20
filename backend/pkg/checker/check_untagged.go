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

		if generated >= maxBackfillPerRun {
			break // bound AI cost; remaining days handled next run
		}
		generated++

		issue, ok := c.processItem(db, cfg, logger, family, item, knownTags, untagged)
		if ok {
			issues = append(issues, issue)
		}
	}
	return issues, nil
}

// processItem generates suggestions for one candidate day. Under auto mode it
// applies confident tags immediately (resolving the day, no issue). Otherwise it
// stages the suggestions as pending and returns a non-fixable "review" issue.
// ok=false when there's nothing to surface (resolved, empty, or error).
func (c UntaggedCheck) processItem(
	db database.Storage, cfg *config.Config, logger *slog.Logger,
	family *models.Family, item *models.Item, knownTags []string, untagged bool,
) (Issue, bool) {
	familyID := family.ID
	suggestions, err := c.Suggester.SuggestTags(context.Background(), item.Title, item.Body, knownTags)
	if err != nil {
		logger.Error("Untagged check: suggestion failed", "familyID", familyID, "date", item.Date, "error", err)
		return Issue{}, false
	}
	names, confident := ai.Partition(suggestions, item.Tags, cfg.AITaggingThreshold)
	if len(names) == 0 {
		return Issue{}, false
	}

	// Auto mode: apply confident tags to an untagged day right away — no manual
	// "fix" step. The day is resolved, so it produces no issue.
	if family.AITaggingAuto && untagged && len(confident) > 0 {
		if err := db.AddConfirmedTags(familyID, item.Date, confident); err != nil {
			logger.Error("Untagged check: auto-apply failed", "familyID", familyID, "date", item.Date, "error", err)
			return Issue{}, false
		}
		logger.Info("Untagged check: auto-applied confident tags", "familyID", familyID, "date", item.Date, "tags", confident)
		return Issue{}, false
	}

	// Non-auto, or uncertain under auto: stage suggestions for per-entry review.
	if err := db.SetPendingTags(familyID, item.Date, names); err != nil {
		logger.Error("Untagged check: failed to stage pending tags", "familyID", familyID, "date", item.Date, "error", err)
		return Issue{}, false
	}
	return c.reviewIssue(familyID, item.Date, len(names)), true
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
