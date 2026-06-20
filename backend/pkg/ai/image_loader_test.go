package ai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadImageAssetsEmpty(t *testing.T) {
	result := LoadImageAssets("no images here", t.TempDir(), "family1")
	if len(result) != 0 {
		t.Fatalf("expected no assets, got %d", len(result))
	}
}

func TestLoadImageAssetsMissingFile(t *testing.T) {
	// Referenced file does not exist — should silently skip it.
	result := LoadImageAssets("![alt](photo.jpg)", t.TempDir(), "family1")
	if len(result) != 0 {
		t.Fatalf("expected no assets for missing file, got %d", len(result))
	}
}

func TestLoadImageAssetsJPEG(t *testing.T) {
	dir := t.TempDir()
	familyID := "family-test"
	assetsDir := filepath.Join(dir, "diary-assets", familyID)
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Minimal valid JPEG (SOI + EOI markers).
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10,
		0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01,
		0x00, 0x01, 0x00, 0x00, 0xFF, 0xD9}
	if err := os.WriteFile(filepath.Join(assetsDir, "photo.jpg"), jpegData, 0o644); err != nil {
		t.Fatal(err)
	}

	body := "![alt](photo.jpg)"
	result := LoadImageAssets(body, dir, familyID)
	if len(result) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(result))
	}
	if result[0].MIMEType != "image/jpeg" {
		t.Errorf("expected image/jpeg, got %q", result[0].MIMEType)
	}
}

func TestLoadImageAssetsPathTraversal(t *testing.T) {
	// "../secret" should be rejected by the path-traversal guard.
	result := LoadImageAssets("![alt](../secret.jpg)", t.TempDir(), "family1")
	if len(result) != 0 {
		t.Fatalf("expected no assets for path traversal attempt, got %d", len(result))
	}
}

func TestLoadImageAssetsNonImage(t *testing.T) {
	dir := t.TempDir()
	familyID := "family-test"
	assetsDir := filepath.Join(dir, "diary-assets", familyID)
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// A plain-text file should be skipped (MIME not in supported set).
	if err := os.WriteFile(filepath.Join(assetsDir, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := LoadImageAssets("![alt](note.txt)", dir, familyID)
	if len(result) != 0 {
		t.Fatalf("expected non-image file to be skipped, got %d", len(result))
	}
}
