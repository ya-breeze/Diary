package models

// OrphanIgnore records an asset filename that a user has chosen to ignore in health checks.
type OrphanIgnore struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	UserID   string `gorm:"index:idx_orphan_ignore_user_file,unique"`
	Filename string `gorm:"index:idx_orphan_ignore_user_file,unique"`
}
