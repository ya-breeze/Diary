package database

import (
	"log/slog"
	"os"
	"testing"

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
