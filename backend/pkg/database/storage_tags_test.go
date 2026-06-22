package database

import (
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database/models"
)

func TestGetDistinctTagsAndAISettings(t *testing.T) {
	logger := slog.Default()
	tempDir, err := os.MkdirTemp("", "tags_test")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s := NewStorage(logger, &config.Config{DataPath: tempDir})
	if err := s.Open(); err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()

	fam, err := s.CreateFamily("fam")
	if err != nil {
		t.Fatalf("create family: %v", err)
	}

	// Empty before any tagged entries.
	if got, err := s.GetDistinctTags(fam.ID); err != nil || len(got) != 0 {
		t.Fatalf("expected empty, got %v err %v", got, err)
	}

	// Two entries sharing one tag, plus a unique one each.
	for _, it := range []*models.Item{
		{Date: "2024-01-01", Title: "a", Tags: models.StringList{"travel", "family"}},
		{Date: "2024-01-02", Title: "b", Tags: models.StringList{"family", "work"}},
	} {
		if err := s.PutItem(fam.ID, it); err != nil {
			t.Fatalf("put item: %v", err)
		}
	}

	tags, err := s.GetDistinctTags(fam.ID)
	if err != nil {
		t.Fatalf("GetDistinctTags: %v", err)
	}
	want := []string{"family", "travel", "work"} // deduped + sorted
	if len(tags) != len(want) {
		t.Fatalf("got %v want %v", tags, want)
	}
	for i := range want {
		if tags[i] != want[i] {
			t.Fatalf("got %v want %v", tags, want)
		}
	}

	// A second family does not see the first family's tags.
	other, _ := s.CreateFamily("other")
	if got, _ := s.GetDistinctTags(other.ID); len(got) != 0 {
		t.Fatalf("expected family scoping, got %v", got)
	}

	// AI settings toggle round-trips.
	if got, _ := s.GetFamily(fam.ID); got.AITaggingEnabled {
		t.Fatal("expected AITaggingEnabled to default false")
	}
	if err := s.SetFamilyAITaggingEnabled(fam.ID, true); err != nil {
		t.Fatalf("SetFamilyAITaggingEnabled: %v", err)
	}
	if got, _ := s.GetFamily(fam.ID); !got.AITaggingEnabled {
		t.Fatal("expected AITaggingEnabled true after update")
	}
}

// newTagStorage spins up a fresh on-disk storage with one family for tag tests.
func newTagStorage(t *testing.T) (Storage, *models.Family) {
	t.Helper()
	s := NewStorage(slog.Default(), &config.Config{DataPath: t.TempDir()})
	if err := s.Open(); err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	fam, err := s.CreateFamily("fam")
	if err != nil {
		t.Fatalf("create family: %v", err)
	}
	return s, fam
}

func putItems(t *testing.T, s Storage, familyID uuid.UUID, items ...*models.Item) {
	t.Helper()
	for _, it := range items {
		if err := s.PutItem(familyID, it); err != nil {
			t.Fatalf("put item %s: %v", it.Date, err)
		}
	}
}

func tagsOf(t *testing.T, s Storage, familyID uuid.UUID, date string) []string {
	t.Helper()
	item, err := s.GetItem(familyID, date)
	if err != nil {
		t.Fatalf("get item %s: %v", date, err)
	}
	return []string(item.Tags)
}

func TestGetTagStats(t *testing.T) {
	s, fam := newTagStorage(t)

	// Empty before any tags.
	if got, err := s.GetTagStats(fam.ID); err != nil || len(got) != 0 {
		t.Fatalf("expected empty stats, got %v err %v", got, err)
	}

	putItems(t, s, fam.ID,
		&models.Item{Date: "2024-01-01", Title: "a", Tags: models.StringList{"family", "travel"}},
		&models.Item{Date: "2024-01-02", Title: "b", Tags: models.StringList{"family", "work"}},
		&models.Item{Date: "2024-01-03", Title: "c", Tags: models.StringList{"family"}},
	)

	stats, err := s.GetTagStats(fam.ID)
	if err != nil {
		t.Fatalf("GetTagStats: %v", err)
	}
	// Sorted by count desc then name asc: family(3), travel(1), work(1).
	want := []TagStat{{"family", 3}, {"travel", 1}, {"work", 1}}
	if len(stats) != len(want) {
		t.Fatalf("got %v want %v", stats, want)
	}
	for i := range want {
		if stats[i] != want[i] {
			t.Fatalf("at %d got %v want %v", i, stats[i], want[i])
		}
	}

	// Family scoping: a second family sees nothing.
	other, _ := s.CreateFamily("other")
	if got, _ := s.GetTagStats(other.ID); len(got) != 0 {
		t.Fatalf("expected scoping, got %v", got)
	}
}

func TestRenameTag(t *testing.T) {
	s, fam := newTagStorage(t)

	putItems(t, s, fam.ID,
		&models.Item{Date: "2024-01-01", Title: "a", Tags: models.StringList{"vacaiton", "work"}},
		&models.Item{Date: "2024-01-02", Title: "b", Tags: models.StringList{"vacaiton"}},
		// Collision: already carries the target name alongside the typo.
		&models.Item{Date: "2024-01-03", Title: "c", Tags: models.StringList{"vacaiton", "vacation"}},
		&models.Item{Date: "2024-01-04", Title: "d", Tags: models.StringList{"unrelated"}},
	)

	if err := s.RenameTag(fam.ID, "vacaiton", "vacation"); err != nil {
		t.Fatalf("RenameTag: %v", err)
	}

	if got := tagsOf(t, s, fam.ID, "2024-01-01"); !equalTags(got, []string{"vacation", "work"}) {
		t.Fatalf("01-01 got %v", got)
	}
	if got := tagsOf(t, s, fam.ID, "2024-01-02"); !equalTags(got, []string{"vacation"}) {
		t.Fatalf("01-02 got %v", got)
	}
	// Merge-on-collision: a single de-duplicated tag.
	if got := tagsOf(t, s, fam.ID, "2024-01-03"); !equalTags(got, []string{"vacation"}) {
		t.Fatalf("01-03 (merge) got %v", got)
	}
	// Untouched entry is unchanged.
	if got := tagsOf(t, s, fam.ID, "2024-01-04"); !equalTags(got, []string{"unrelated"}) {
		t.Fatalf("01-04 got %v", got)
	}

	// Renaming a non-existent tag changes nothing (no error).
	if err := s.RenameTag(fam.ID, "ghost", "spirit"); err != nil {
		t.Fatalf("RenameTag non-existent: %v", err)
	}
	if got := tagsOf(t, s, fam.ID, "2024-01-04"); !equalTags(got, []string{"unrelated"}) {
		t.Fatalf("01-04 after ghost rename got %v", got)
	}

	// Family scoping: another family with the same tag is untouched.
	other, _ := s.CreateFamily("other")
	putItems(t, s, other.ID, &models.Item{Date: "2024-01-01", Title: "x", Tags: models.StringList{"vacation"}})
	if err := s.RenameTag(fam.ID, "vacation", "holiday"); err != nil {
		t.Fatalf("RenameTag: %v", err)
	}
	if got := tagsOf(t, s, other.ID, "2024-01-01"); !equalTags(got, []string{"vacation"}) {
		t.Fatalf("other family changed: %v", got)
	}
}

func TestDeleteTag(t *testing.T) {
	s, fam := newTagStorage(t)

	putItems(t, s, fam.ID,
		&models.Item{Date: "2024-01-01", Title: "a", Tags: models.StringList{"misc", "work"}},
		&models.Item{Date: "2024-01-02", Title: "b", Tags: models.StringList{"misc"}},
		&models.Item{Date: "2024-01-03", Title: "c", Tags: models.StringList{"work"}},
	)

	if err := s.DeleteTag(fam.ID, "misc"); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}
	if got := tagsOf(t, s, fam.ID, "2024-01-01"); !equalTags(got, []string{"work"}) {
		t.Fatalf("01-01 got %v", got)
	}
	// Last tag removed leaves an empty list.
	if got := tagsOf(t, s, fam.ID, "2024-01-02"); len(got) != 0 {
		t.Fatalf("01-02 expected empty, got %v", got)
	}
	if got := tagsOf(t, s, fam.ID, "2024-01-03"); !equalTags(got, []string{"work"}) {
		t.Fatalf("01-03 got %v", got)
	}

	// Deleting a non-existent tag is a no-op.
	if err := s.DeleteTag(fam.ID, "ghost"); err != nil {
		t.Fatalf("DeleteTag non-existent: %v", err)
	}
}

// TestGetItemsTagFilterToleratesMalformedTags reproduces the browse-by-tag 500:
// a legacy row whose tags column holds a non-JSON value (e.g. an empty string)
// must not break a tag-only filter. The JSON_EXTRACT-based query raised
// "malformed JSON" on such a row and failed the whole scan.
func TestGetItemsTagFilterToleratesMalformedTags(t *testing.T) {
	s, fam := newTagStorage(t)

	putItems(t, s, fam.ID,
		&models.Item{Date: "2024-05-01", Title: "tagged", Tags: models.StringList{"girls"}},
		&models.Item{Date: "2024-05-02", Title: "legacy", Tags: models.StringList{"girls"}},
	)
	// Force a non-JSON (empty-string) tags column on the second row to mimic the
	// legacy data that triggered the production "malformed JSON" error.
	if err := s.GetDB().Exec(
		"UPDATE items SET tags = '' WHERE family_id = ? AND date = ?", fam.ID, "2024-05-02",
	).Error; err != nil {
		t.Fatalf("force empty tags: %v", err)
	}

	// A tag-only filter (no date, no search) must not error and must find the
	// well-formed tagged row.
	items, _, err := s.GetItems(fam.ID, SearchParams{Tags: []string{"girls"}})
	if err != nil {
		t.Fatalf("GetItems with malformed row present: %v", err)
	}
	if len(items) != 1 || items[0].Date != "2024-05-01" {
		t.Fatalf("expected the single tagged row, got %d items", len(items))
	}
}

func equalTags(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
