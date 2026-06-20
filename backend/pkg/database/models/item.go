package models

import (
	coremodels "github.com/ya-breeze/kin-core/models"
)

type Item struct {
	coremodels.TenantModel
	// Composite unique index (family_id, date) is created manually in migration.
	Date  string `gorm:"not null"`
	Title string
	Body  string
	Tags  StringList `gorm:"type:json"`
	// PendingTags holds AI tag suggestions awaiting user acceptance. Kept disjoint
	// from Tags: a name confirmed in Tags is never also present in PendingTags.
	PendingTags StringList `gorm:"type:json"`
	// TagsSourceHash is the hash of the title, body, and sorted referenced asset
	// filenames at the time tags were last computed. Used to detect staleness for
	// edit-triggered retagging and the backfill health check.
	TagsSourceHash string
}
