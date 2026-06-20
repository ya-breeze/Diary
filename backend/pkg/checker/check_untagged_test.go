package checker

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/ya-breeze/diary.be/pkg/ai"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
)

// fakeSuggester returns canned suggestions for the untagged check tests.
type fakeSuggester struct {
	enabled     bool
	suggestions []ai.TagSuggestion
}

func (f fakeSuggester) Enabled() bool { return f.enabled }
func (f fakeSuggester) SuggestTags(
	_ context.Context, _, _ string, _ []string,
) ([]ai.TagSuggestion, error) {
	return f.suggestions, nil
}

func setupUntagged(t *testing.T) (database.Storage, *config.Config, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "untagged_test")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	cfg := &config.Config{DataPath: tempDir, AITaggingThreshold: 0.8}
	s := database.NewStorage(slog.Default(), cfg)
	if err := s.Open(); err != nil {
		t.Fatalf("open: %v", err)
	}
	return s, cfg, func() { s.Close(); os.RemoveAll(tempDir) }
}

// makeLegacy clears a day's tags_source_hash, simulating an entry created before
// AI tagging existed (an empty hash reads as "stale" → a backfill candidate).
func makeLegacy(t *testing.T, s database.Storage, date string) {
	t.Helper()
	if err := s.GetDB().Model(&models.Item{}).
		Where("date = ?", date).Update("tags_source_hash", "").Error; err != nil {
		t.Fatalf("makeLegacy: %v", err)
	}
}

func runUntagged(t *testing.T, s database.Storage, cfg *config.Config, sug ai.Suggester) []Issue {
	t.Helper()
	issues, err := UntaggedCheck{Suggester: sug}.Run(s, cfg, slog.Default())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return issues
}

func TestUntaggedDisabledSuggester(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false)
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "x"})

	if issues := runUntagged(t, s, cfg, fakeSuggester{enabled: false}); len(issues) != 0 {
		t.Fatalf("disabled suggester should produce no issues, got %d", len(issues))
	}
}

func TestUntaggedBackfillOff(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, false, false) // backfill OFF
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "beach day"})

	sug := fakeSuggester{enabled: true, suggestions: []ai.TagSuggestion{{Name: "beach", Confidence: 0.9}}}
	if issues := runUntagged(t, s, cfg, sug); len(issues) != 0 {
		t.Fatalf("backfill off should produce no issues, got %d", len(issues))
	}
}

func TestUntaggedNonAutoStagesPending(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false) // backfill on, auto OFF
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "beach day"})
	makeLegacy(t, s, "2024-01-01")

	sug := fakeSuggester{enabled: true, suggestions: []ai.TagSuggestion{{Name: "beach", Confidence: 0.95}}}
	issues := runUntagged(t, s, cfg, sug)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Fixable {
		t.Fatal("non-auto issue should not be fixable")
	}
	// Suggestions staged as pending; confirmed tags untouched.
	item, _ := s.GetItem(fam.ID, "2024-01-01")
	if len(item.Tags) != 0 {
		t.Fatalf("confirmed tags should be empty, got %v", item.Tags)
	}
	if len(item.PendingTags) != 1 || item.PendingTags[0] != "beach" {
		t.Fatalf("expected pending [beach], got %v", item.PendingTags)
	}
}

func TestUntaggedAutoConfidentFixApplies(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, true) // auto ON
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "beach day"})
	makeLegacy(t, s, "2024-01-01")

	sug := fakeSuggester{enabled: true, suggestions: []ai.TagSuggestion{{Name: "beach", Confidence: 0.95}}}
	issues := runUntagged(t, s, cfg, sug)
	// Auto mode applies confident tags during the run and resolves the day —
	// no issue is surfaced.
	if len(issues) != 0 {
		t.Fatalf("expected no issue (auto-applied), got %+v", issues)
	}
	item, _ := s.GetItem(fam.ID, "2024-01-01")
	if len(item.Tags) != 1 || item.Tags[0] != "beach" {
		t.Fatalf("expected confirmed [beach] auto-applied, got %v", item.Tags)
	}
}

func TestUntaggedAutoUncertainStagesPending(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, true) // auto ON
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "beach day"})
	makeLegacy(t, s, "2024-01-01")

	// Below threshold (0.8) → uncertain → manual review, not auto-applied.
	sug := fakeSuggester{enabled: true, suggestions: []ai.TagSuggestion{{Name: "beach", Confidence: 0.5}}}
	issues := runUntagged(t, s, cfg, sug)
	if len(issues) != 1 || issues[0].Fixable {
		t.Fatalf("expected 1 non-fixable issue, got %+v", issues)
	}
	item, _ := s.GetItem(fam.ID, "2024-01-01")
	if len(item.Tags) != 0 {
		t.Fatalf("uncertain suggestion must not be auto-applied, got %v", item.Tags)
	}
	if len(item.PendingTags) != 1 {
		t.Fatalf("expected pending staged, got %v", item.PendingTags)
	}
}

func TestUntaggedSkipsTaggedDays(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, true)
	// Already tagged and saved (hash up to date) → not a candidate.
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "beach day", Tags: models.StringList{"beach"}})

	sug := fakeSuggester{enabled: true, suggestions: []ai.TagSuggestion{{Name: "ocean", Confidence: 0.95}}}
	if issues := runUntagged(t, s, cfg, sug); len(issues) != 0 {
		t.Fatalf("tagged up-to-date day should be skipped, got %d issues", len(issues))
	}
}

// countingSuggester records how many times the model was queried.
type countingSuggester struct {
	calls int
	out   []ai.TagSuggestion
}

func (c *countingSuggester) Enabled() bool { return true }
func (c *countingSuggester) SuggestTags(
	_ context.Context, _, _ string, _ []string,
) ([]ai.TagSuggestion, error) {
	c.calls++
	return c.out, nil
}

func TestUntaggedDoesNotRequeryStagedDays(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false) // non-auto: stages pending
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "beach day"})
	makeLegacy(t, s, "2024-01-01")

	sug := &countingSuggester{out: []ai.TagSuggestion{{Name: "beach", Confidence: 0.95}}}

	// First run stages pending and stamps the hash (1 model call).
	if issues := runUntagged(t, s, cfg, sug); len(issues) != 1 {
		t.Fatalf("first run: expected 1 issue, got %d", len(issues))
	}
	if sug.calls != 1 {
		t.Fatalf("first run: expected 1 model call, got %d", sug.calls)
	}

	// Second run, content unchanged: surfaces staged suggestions without re-querying.
	if issues := runUntagged(t, s, cfg, sug); len(issues) != 1 {
		t.Fatalf("second run: expected 1 issue, got %d", len(issues))
	}
	if sug.calls != 1 {
		t.Fatalf("second run: model should not be called again, got %d calls", sug.calls)
	}
}

func TestUntaggedDismissClearsFromReview(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false) // non-auto
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "beach day"})
	makeLegacy(t, s, "2024-01-01")

	sug := &countingSuggester{out: []ai.TagSuggestion{{Name: "beach", Confidence: 0.95}}}

	// First run stages "beach" as pending and reports a review issue.
	if issues := runUntagged(t, s, cfg, sug); len(issues) != 1 {
		t.Fatalf("first run: expected 1 issue, got %d", len(issues))
	}

	// The user dismisses the only suggestion → pending becomes empty.
	if err := s.SetPendingTags(fam.ID, "2024-01-01", nil); err != nil {
		t.Fatalf("dismiss: %v", err)
	}

	// Next run must NOT re-surface or re-query the dismissed day.
	if issues := runUntagged(t, s, cfg, sug); len(issues) != 0 {
		t.Fatalf("after dismiss: expected no issue, got %+v", issues)
	}
	if sug.calls != 1 {
		t.Fatalf("after dismiss: model should not be called again, got %d calls", sug.calls)
	}
}
