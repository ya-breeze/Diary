package checker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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
		if !untagged && !stale {
			continue // already tagged and up to date
		}

		// Already-staged suggestions (content unchanged): surface without a new call.
		if len(item.PendingTags) > 0 && !stale {
			issues = append(issues, c.pendingIssue(familyID, item.Date, len(item.PendingTags)))
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

// processItem generates suggestions for one candidate day and returns the issue
// to surface (ok=false when there's nothing to suggest or an error occurred).
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
	names, confident := splitByConfidence(suggestions, item.Tags, cfg.AITaggingThreshold)
	if len(names) == 0 {
		return Issue{}, false
	}

	if family.AITaggingAuto && untagged && len(confident) > 0 {
		return Issue{
			Check:    "untagged",
			FamilyID: familyID.String(),
			Path:     item.Date,
			Message:  fmt.Sprintf("%d confident tag(s) can be auto-applied", len(confident)),
			Fixable:  true,
			fix:      makeUntaggedFix(db, logger, familyID, item.Date, confident),
		}, true
	}

	// Non-auto, or uncertain under auto: stage suggestions for manual review.
	if err := db.SetPendingTags(familyID, item.Date, names); err != nil {
		logger.Error("Untagged check: failed to stage pending tags", "familyID", familyID, "date", item.Date, "error", err)
		return Issue{}, false
	}
	msg := fmt.Sprintf("%d suggested tag(s) — review on the entry", len(names))
	if family.AITaggingAuto {
		msg = fmt.Sprintf("%d low-confidence suggestion(s) — solve manually", len(names))
	}
	return Issue{
		Check:    "untagged",
		FamilyID: familyID.String(),
		Path:     item.Date,
		Message:  msg,
		Fixable:  false,
	}, true
}

func (UntaggedCheck) pendingIssue(familyID uuid.UUID, date string, n int) Issue {
	return Issue{
		Check:    "untagged",
		FamilyID: familyID.String(),
		Path:     date,
		Message:  fmt.Sprintf("%d suggested tag(s) awaiting review", n),
		Fixable:  false,
	}
}

// splitByConfidence returns all suggested names (excluding tags already confirmed
// on the entry) and the subset whose confidence meets the threshold.
func splitByConfidence(
	suggestions []ai.TagSuggestion, confirmed []string, threshold float64,
) ([]string, []string) {
	confirmedSet := make(map[string]struct{}, len(confirmed))
	for _, t := range confirmed {
		confirmedSet[strings.ToLower(t)] = struct{}{}
	}
	var names, confident []string
	for _, s := range suggestions {
		if _, ok := confirmedSet[strings.ToLower(s.Name)]; ok {
			continue
		}
		names = append(names, s.Name)
		if s.Confidence >= threshold {
			confident = append(confident, s.Name)
		}
	}
	return names, confident
}

// makeUntaggedFix returns a closure that adds the confident tags to a day's
// confirmed tags (additive — never removes existing tags) and clears those names
// from its pending list.
func makeUntaggedFix(
	db database.Storage, logger *slog.Logger, familyID uuid.UUID, date string, confident []string,
) func() error {
	return func() error {
		item, err := db.GetItem(familyID, date)
		if err != nil {
			return fmt.Errorf("getting item %s/%s: %w", familyID, date, err)
		}
		existing := make(map[string]struct{}, len(item.Tags))
		for _, t := range item.Tags {
			existing[strings.ToLower(t)] = struct{}{}
		}
		merged := append(models.StringList{}, item.Tags...)
		for _, name := range confident {
			if _, ok := existing[strings.ToLower(name)]; !ok {
				merged = append(merged, name)
			}
		}
		item.Tags = merged
		if err := db.PutItem(familyID, item); err != nil {
			return fmt.Errorf("saving item %s/%s: %w", familyID, date, err)
		}
		logger.Info("Auto-applied confident tags", "date", date, "tags", confident)
		return nil
	}
}
