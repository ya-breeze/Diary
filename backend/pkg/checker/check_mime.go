package checker

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
)

// MimeCheck finds asset files whose extension doesn't match their actual content.
// Currently detects videos saved with .jpg extension (caused by old single-upload bug).
type MimeCheck struct{}

func (MimeCheck) Name() string { return "mime" }

func (MimeCheck) Run(db database.Storage, cfg *config.Config, logger *slog.Logger) ([]Issue, error) {
	users, err := db.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}

	var issues []Issue
	assetsBase := filepath.Join(cfg.DataPath, config.AssetsDirName)

	for _, user := range users {
		familyID := user.FamilyID
		familyDir := filepath.Join(assetsBase, familyID.String())

		entries, err := os.ReadDir(familyDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading asset dir for family %s: %w", familyID, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".jpg" {
				continue
			}

			filePath := filepath.Join(familyDir, entry.Name())
			ext, err := detectVideoExtension(filePath)
			if err != nil {
				logger.Warn("Could not read magic bytes", "file", filePath, "error", err)
				continue
			}
			if ext == "" {
				continue // not a video
			}

			newName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) + ext
			newPath := filepath.Join(familyDir, newName)
			oldName := entry.Name()

			issues = append(issues, Issue{
				Check:    "mime",
				FamilyID: familyID.String(),
				Path:     filePath,
				Message:  fmt.Sprintf("video file saved as .jpg, should be %s", ext),
				Fixable:  true,
				fix:      makeMimeFix(db, logger, familyID, filePath, newPath, oldName, newName),
			})
		}
	}

	return issues, nil
}

// detectVideoExtension reads magic bytes and returns the correct extension if the
// file is a video, or "" if it appears to be a real image.
func detectVideoExtension(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 12)
	n, err := f.Read(buf)
	if err != nil || n < 8 {
		return "", nil
	}

	switch {
	// MP4 / MOV: bytes 4–7 are "ftyp"
	case n >= 8 && string(buf[4:8]) == "ftyp":
		return ".mp4", nil
	// AVI: bytes 0–3 are "RIFF"
	case string(buf[0:4]) == "RIFF" && n >= 12 && string(buf[8:12]) == "AVI ":
		return ".avi", nil
	// MKV / WebM: bytes 0–3 are 0x1A 0x45 0xDF 0xA3
	case buf[0] == 0x1A && buf[1] == 0x45 && buf[2] == 0xDF && buf[3] == 0xA3:
		return ".mkv", nil
	}

	return "", nil
}

func makeMimeFix(
	db database.Storage,
	logger *slog.Logger,
	familyID uuid.UUID,
	oldPath, newPath, oldName, newName string,
) func() error {
	return func() error {
		// Update all diary entries for this family that reference the old filename first,
		// so that a partial failure never leaves the file renamed but DB still pointing to the old name.
		items, _, err := db.GetItems(familyID, database.SearchParams{SearchText: oldName})
		if err != nil {
			return fmt.Errorf("querying items for family %s: %w", familyID, err)
		}
		for _, item := range items {
			if !strings.Contains(item.Body, oldName) {
				continue
			}
			item.Body = strings.ReplaceAll(item.Body, oldName, newName)
			if err := db.PutItem(familyID, item); err != nil {
				return fmt.Errorf("updating item %s/%s: %w", familyID, item.Date, err)
			}
			logger.Info("Updated item body", "date", item.Date, "old", oldName, "new", newName)
		}

		// Rename the file only after all DB entries have been updated successfully.
		// Idempotent: if target already exists, skip rename.
		if _, err := os.Stat(newPath); err == nil {
			logger.Info("Target already exists, skipping rename", "target", newPath)
		} else {
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("renaming %s -> %s: %w", oldPath, newPath, err)
			}
			logger.Info("Renamed asset file", "old", oldName, "new", newName)
		}

		return nil
	}
}
