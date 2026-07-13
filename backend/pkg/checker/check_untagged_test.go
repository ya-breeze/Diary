package checker

import (
	"context"
	"fmt"
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
	_ context.Context, _, _ string, _ []ai.ImageAsset, _ []string,
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
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false)
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "x"})

	if issues := runUntagged(t, s, cfg, fakeSuggester{enabled: false}); len(issues) != 0 {
		t.Fatalf("disabled suggester should produce no issues, got %d", len(issues))
	}
}

func TestUntaggedEmptyResultMarkedAnalyzedNotRequeried(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false)
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "quiet day"})
	makeLegacy(t, s, "2024-01-01")

	// The model has nothing to suggest: the day must still be marked analyzed
	// (hash stamped) so the next run does not re-query it.
	sug := &countingSuggester{out: nil}
	if issues := runUntagged(t, s, cfg, sug); len(issues) != 0 {
		t.Fatalf("first run: expected no issues, got %d", len(issues))
	}
	if sug.calls != 1 {
		t.Fatalf("first run: expected 1 model call, got %d", sug.calls)
	}
	if issues := runUntagged(t, s, cfg, sug); len(issues) != 0 {
		t.Fatalf("second run: expected no issues, got %d", len(issues))
	}
	if sug.calls != 1 {
		t.Fatalf("second run: day was re-queried (calls=%d)", sug.calls)
	}
}

func TestUntaggedBackfillDoneFlipsWhenExhausted(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false)
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "beach day"})
	makeLegacy(t, s, "2024-01-01")

	sug := &countingSuggester{out: nil}
	_ = runUntagged(t, s, cfg, sug)

	got, err := s.GetFamily(fam.ID)
	if err != nil {
		t.Fatalf("GetFamily: %v", err)
	}
	if !got.AITaggingBackfillDone {
		t.Fatal("expected AITaggingBackfillDone=true after the corpus is exhausted")
	}
}

func TestUntaggedBackfillDoneNotFlippedAtCap(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false)
	// One more candidate than the per-run cap: the run must stop at the cap and
	// NOT mark the backfill done (work remains for the next run).
	for i := 0; i <= maxBackfillPerRun; i++ {
		date := fmt.Sprintf("2024-02-%02d", i+1)
		_ = s.PutItem(fam.ID, &models.Item{Date: date, Title: "day " + date})
		makeLegacy(t, s, date)
	}

	sug := &countingSuggester{out: nil}
	_ = runUntagged(t, s, cfg, sug)
	if sug.calls != maxBackfillPerRun {
		t.Fatalf("expected exactly %d model calls (the cap), got %d", maxBackfillPerRun, sug.calls)
	}
	got, _ := s.GetFamily(fam.ID)
	if got.AITaggingBackfillDone {
		t.Fatal("backfill must not be marked done when the run stopped at the cap")
	}

	// The next run finishes the remainder and flips done.
	_ = runUntagged(t, s, cfg, sug)
	got, _ = s.GetFamily(fam.ID)
	if !got.AITaggingBackfillDone {
		t.Fatal("expected done after the second run exhausted the corpus")
	}
}

func TestUntaggedDoneFamilyNoCallsButPendingStillSurfaced(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false)
	// One day with staged pending (analyzed), one never-analyzed day.
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "reviewed day"})
	_ = s.SetPendingTags(fam.ID, "2024-01-01", []string{"beach"}) // stamps hash
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-02", Title: "never analyzed"})
	makeLegacy(t, s, "2024-01-02")
	// Mark the family's backfill complete.
	if err := s.SetFamilyBackfillDone(fam.ID, true); err != nil {
		t.Fatalf("SetFamilyBackfillDone: %v", err)
	}

	sug := &countingSuggester{out: []ai.TagSuggestion{{Name: "x", Confidence: 0.9}}}
	issues := runUntagged(t, s, cfg, sug)
	if sug.calls != 0 {
		t.Fatalf("done family must make no model calls, got %d", sug.calls)
	}
	if len(issues) != 1 {
		t.Fatalf("expected the staged-pending day to still be surfaced (1 issue), got %d", len(issues))
	}
}

func TestBackfillToggleResetsDone(t *testing.T) {
	s, _, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false)
	if err := s.SetFamilyBackfillDone(fam.ID, true); err != nil {
		t.Fatalf("SetFamilyBackfillDone: %v", err)
	}

	// Toggle backfill off, then on: done must reset to false.
	_ = s.SetFamilyAISettings(fam.ID, true, false, false, false, false)
	got, _ := s.GetFamily(fam.ID)
	if !got.AITaggingBackfillDone {
		t.Fatal("turning backfill OFF must not reset done")
	}
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false)
	got, _ = s.GetFamily(fam.ID)
	if got.AITaggingBackfillDone {
		t.Fatal("turning backfill back ON must reset done to false")
	}
}

func TestUntaggedEmptySuggestionsTolerated(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false)
	_ = s.PutItem(fam.ID, &models.Item{Date: "2024-01-01", Title: "beach day"})
	makeLegacy(t, s, "2024-01-01")

	// A model that produced no usable content yields (nil, nil): the check must
	// treat the day as having no suggestions — no issue, no error (runUntagged
	// fails the test on error).
	sug := fakeSuggester{enabled: true, suggestions: nil}
	if issues := runUntagged(t, s, cfg, sug); len(issues) != 0 {
		t.Fatalf("empty suggestions should produce no issues, got %d", len(issues))
	}
}

func TestUntaggedBackfillOff(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, false, false, false, false) // backfill OFF
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
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false) // backfill on, auto OFF
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
	_ = s.SetFamilyAISettings(fam.ID, true, true, true, false, false) // auto ON
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
	_ = s.SetFamilyAISettings(fam.ID, true, true, true, false, false) // auto ON
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
	_ = s.SetFamilyAISettings(fam.ID, true, true, true, false, false)
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
	_ context.Context, _, _ string, _ []ai.ImageAsset, _ []string,
) ([]ai.TagSuggestion, error) {
	c.calls++
	return c.out, nil
}

func TestUntaggedDoesNotRequeryStagedDays(t *testing.T) {
	s, cfg, done := setupUntagged(t)
	defer done()
	fam, _ := s.CreateFamily("f")
	_, _ = s.CreateUser("u", "p", fam.ID)
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false) // non-auto: stages pending
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
	_ = s.SetFamilyAISettings(fam.ID, true, true, false, false, false) // non-auto
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
