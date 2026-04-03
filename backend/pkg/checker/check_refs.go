package checker

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

// RefsCheck finds diary entries that reference asset files which no longer exist on disk.
type RefsCheck struct{}

func (RefsCheck) Name() string { return "refs" }

func (RefsCheck) Run(db database.Storage, cfg *config.Config, logger *slog.Logger) ([]Issue, error) {
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
						Fixable: true,
						fix:     makeRefsFix(db, logger, userID, item.Date, name),
					})
				}
			}
		}
	}

	return issues, nil
}

// makeRefsFix returns a closure that removes a broken image reference from a diary entry.
func makeRefsFix(db database.Storage, logger *slog.Logger, userID, date, filename string) func() error {
	return func() error {
		item, err := db.GetItem(userID, date)
		if err != nil {
			return fmt.Errorf("getting item %s/%s: %w", userID, date, err)
		}
		newBody := removeMarkdownImageRef(item.Body, filename)
		if newBody == item.Body {
			return nil // already clean
		}
		item.Body = newBody
		if err := db.PutItem(userID, item); err != nil {
			return fmt.Errorf("saving item %s/%s: %w", userID, date, err)
		}
		logger.Info("Removed broken image reference", "date", date, "file", filename)
		return nil
	}
}

// removeMarkdownImageRef strips all occurrences of ![any alt](filename) from md.
func removeMarkdownImageRef(md, filename string) string {
	pattern := `!\[[^\]]*\]\(` + regexp.QuoteMeta(filename) + `\)`
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(md, "")
}
