package checker

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

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
		userID := user.ID.String()
		userDir := filepath.Join(assetsBase, userID)

		entries, err := os.ReadDir(userDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading asset dir for user %s: %w", userID, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".jpg" {
				continue
			}

			filePath := filepath.Join(userDir, entry.Name())
			ext, err := detectVideoExtension(filePath)
			if err != nil {
				logger.Warn("Could not read magic bytes", "file", filePath, "error", err)
				continue
			}
			if ext == "" {
				continue // not a video
			}

			newName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) + ext
			newPath := filepath.Join(userDir, newName)
			oldName := entry.Name()

			issues = append(issues, Issue{
				Check:   "mime",
				UserID:  userID,
				Path:    filePath,
				Message: fmt.Sprintf("video file saved as .jpg, should be %s", ext),
				fixable: true,
				fix:     makeMimeFix(db, cfg, logger, userID, filePath, newPath, oldName, newName),
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
	cfg *config.Config,
	logger *slog.Logger,
	userID, oldPath, newPath, oldName, newName string,
) func() error {
	return func() error {
		// Idempotent: if target already exists, skip rename
		if _, err := os.Stat(newPath); err == nil {
			logger.Info("Target already exists, skipping rename", "target", newPath)
		} else {
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("renaming %s -> %s: %w", oldPath, newPath, err)
			}
			logger.Info("Renamed asset file", "old", oldName, "new", newName)
		}

		// Update all diary entries for this user that reference the old filename
		items, _, err := db.GetItems(userID, database.SearchParams{SearchText: oldName})
		if err != nil {
			return fmt.Errorf("querying items for user %s: %w", userID, err)
		}
		for _, item := range items {
			if !strings.Contains(item.Body, oldName) {
				continue
			}
			item.Body = strings.ReplaceAll(item.Body, oldName, newName)
			if err := db.PutItem(userID, item); err != nil {
				return fmt.Errorf("updating item %s/%s: %w", userID, item.Date, err)
			}
			logger.Info("Updated item body", "date", item.Date, "old", oldName, "new", newName)
		}

		// Also check DataPath for assets stored directly (outside DB) - not needed here
		_ = cfg
		return nil
	}
}
