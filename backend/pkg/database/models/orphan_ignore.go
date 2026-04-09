package models

import "github.com/google/uuid"

// OrphanIgnore records an asset filename that a family has chosen to ignore in health checks.
type OrphanIgnore struct {
	ID       uint      `gorm:"primaryKey;autoIncrement"`
	FamilyID uuid.UUID `gorm:"type:uuid;index:idx_orphan_ignore_family_file,unique"`
	Filename string    `gorm:"index:idx_orphan_ignore_family_file,unique"`
}
