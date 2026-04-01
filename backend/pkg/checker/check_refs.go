package checker

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

// RefsCheck finds diary entries that reference asset files which no longer exist on disk.
type RefsCheck struct{}

func (RefsCheck) Name() string { return "refs" }

func (RefsCheck) Run(db database.Storage, cfg *config.Config, _ *slog.Logger) ([]Issue, error) {
	users, err := db.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}

	var issues []Issue
	assetsBase := filepath.Join(cfg.DataPath, config.AssetsDirName)

	for _, user := range users {
		userID := user.ID.String()
		userDir := filepath.Join(assetsBase, userID)

		items, _, err := db.GetItems(userID, database.SearchParams{})
		if err != nil {
			return nil, fmt.Errorf("getting items for user %s: %w", userID, err)
		}

		for _, item := range items {
			for _, name := range utils.GetAssetsFromMarkdown(item.Body) {
				filePath := filepath.Join(userDir, name)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					issues = append(issues, Issue{
						Check:   "refs",
						UserID:  userID,
						Path:    item.Date + "/" + name,
						Message: fmt.Sprintf("entry %q references missing file %q", item.Date, name),
					})
				}
			}
		}
	}

	return issues, nil
}
