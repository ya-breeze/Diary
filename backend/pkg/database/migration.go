package database

import (
	"gorm.io/gorm"

	"github.com/ya-breeze/diary.be/pkg/database/models"
)

func autoMigrateModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Item{},
		&models.ItemChange{},
	)
}
