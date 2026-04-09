package models

import (
	coremodels "github.com/ya-breeze/kin-core/models"
)

type Item struct {
	coremodels.TenantModel
	Date  string     `gorm:"uniqueIndex:idx_family_date;not null"`
	Title string
	Body  string
	Tags  StringList `gorm:"type:json"`
}
