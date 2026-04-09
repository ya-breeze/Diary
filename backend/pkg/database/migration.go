package database

import (
	"gorm.io/gorm"

	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/kin-core/authdb"
)

func autoMigrateModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Family{},
		&models.User{},
		&models.Item{},
		&models.ItemChange{},
		&models.OrphanIgnore{},
		&authdb.RefreshToken{},
		&authdb.BlacklistedToken{},
	)
}
