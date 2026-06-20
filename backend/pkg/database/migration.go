package database

import (
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/kin-core/authdb"
)

func autoMigrateModels(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Family{},
		&models.User{},
		&models.Item{},
		&models.ItemChange{},
		&models.OrphanIgnore{},
		&authdb.RefreshToken{},
		&authdb.BlacklistedToken{},
	); err != nil {
		return err
	}
	// Composite unique index on items(family_id, date) — can't be defined via GORM field
	// tags because FamilyID lives in embedded TenantModel (kin-core).
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_items_family_date ON items(family_id, date)").Error
}

// scrubBlankTags removes empty and whitespace-only entries from the tags and
// pending_tags of every item. Runs at startup to clean up data written before
// the save-path filter was in place.
func scrubBlankTags(log *slog.Logger, db *gorm.DB) error {
	var items []models.Item
	if err := db.Select("id, tags, pending_tags").Find(&items).Error; err != nil {
		return err
	}
	fixed := 0
	for _, item := range items {
		cleanTags := filterStringList(item.Tags)
		cleanPending := filterStringList(item.PendingTags)
		if len(cleanTags) == len(item.Tags) && len(cleanPending) == len(item.PendingTags) {
			continue
		}
		if err := db.Model(&item).Updates(map[string]any{
			"tags":         models.StringList(cleanTags),
			"pending_tags": models.StringList(cleanPending),
		}).Error; err != nil {
			return err
		}
		fixed++
	}
	if fixed > 0 {
		log.Info("Scrubbed blank tags from items", "count", fixed)
	}
	return nil
}

func filterStringList(in models.StringList) []string {
	out := make([]string, 0, len(in))
	for _, t := range in {
		if s := strings.TrimSpace(t); s != "" {
			out = append(out, s)
		}
	}
	return out
}
