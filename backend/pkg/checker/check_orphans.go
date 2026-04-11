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
		familyID := user.FamilyID
		familyDir := filepath.Join(assetsBase, familyID.String())

		// Build set of all referenced filenames from diary entries
		referenced := make(map[string]bool)
		items, _, err := db.GetItems(familyID, database.SearchParams{})
		if err != nil {
			return nil, fmt.Errorf("getting items for family %s: %w", familyID, err)
		}
		for _, item := range items {
			for _, name := range utils.GetAssetsFromMarkdown(item.Body) {
				referenced[name] = true
			}
		}

		// Build set of ignored filenames
		ignoredList, err := db.GetIgnoredOrphans(familyID)
		if err != nil {
			return nil, fmt.Errorf("getting ignored orphans for family %s: %w", familyID, err)
		}
		ignored := make(map[string]bool, len(ignoredList))
		for _, f := range ignoredList {
			ignored[f] = true
		}

		// Walk the family's asset directory
		entries, err := os.ReadDir(familyDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading asset dir for family %s: %w", familyID, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !referenced[entry.Name()] && !ignored[entry.Name()] {
				issues = append(issues, Issue{
					Check:    "orphans",
					FamilyID: familyID.String(),
					Path:     filepath.Join(familyDir, entry.Name()),
					Message:  "asset file not referenced by any diary entry",
				})
			}
		}
	}

	return issues, nil
}
