package database

import (
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
