package models

import (
	coremodels "github.com/ya-breeze/kin-core/models"
)

type Item struct {
	coremodels.TenantModel
	// Composite unique index (family_id, date) is created manually in migration.
	Date  string     `gorm:"not null"`
	Title string
	Body  string
	Tags  StringList `gorm:"type:json"`
}
