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

// OrphansCheck finds asset files on disk that are not referenced by any diary entry.
type OrphansCheck struct{}

func (OrphansCheck) Name() string { return "orphans" }

func (OrphansCheck) Run(db database.Storage, cfg *config.Config, _ *slog.Logger) ([]Issue, error) {
	users, err := db.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}

	var issues []Issue
	assetsBase := filepath.Join(cfg.DataPath, config.AssetsDirName)

	for _, user := range users {
		userID := user.ID.String()
		userDir := filepath.Join(assetsBase, userID)

		// Build set of all referenced filenames from diary entries
		referenced := make(map[string]bool)
		items, _, err := db.GetItems(userID, database.SearchParams{})
		if err != nil {
			return nil, fmt.Errorf("getting items for user %s: %w", userID, err)
		}
		for _, item := range items {
			for _, name := range utils.GetAssetsFromMarkdown(item.Body) {
				referenced[name] = true
			}
		}

		// Build set of ignored filenames
		ignoredList, err := db.GetIgnoredOrphans(userID)
		if err != nil {
			return nil, fmt.Errorf("getting ignored orphans for user %s: %w", userID, err)
		}
		ignored := make(map[string]bool, len(ignoredList))
		for _, f := range ignoredList {
			ignored[f] = true
		}

		// Walk the user's asset directory
		entries, err := os.ReadDir(userDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading asset dir for user %s: %w", userID, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !referenced[entry.Name()] && !ignored[entry.Name()] {
				issues = append(issues, Issue{
					Check:   "orphans",
					UserID:  userID,
					Path:    filepath.Join(userDir, entry.Name()),
					Message: "asset file not referenced by any diary entry",
				})
			}
		}
	}

	return issues, nil
}
