package database

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/ya-breeze/diary.be/pkg/config"
	"gorm.io/gorm"
)

// userMapping holds the migration mapping for one user.
type userMapping struct {
	oldID    string
	familyID uuid.UUID
}

// oldUser represents the pre-kin-core user row for migration purposes.
type oldUser struct {
	ID             string    `gorm:"column:id"`
	Login          string    `gorm:"column:login"`
	HashedPassword string    `gorm:"column:hashed_password"`
	StartDate      time.Time `gorm:"column:start_date"`
}

func (oldUser) TableName() string { return "users" }

// oldItem represents the pre-kin-core item row for migration purposes.
type oldItem struct {
	UserID string `gorm:"column:user_id"`
	Date   string `gorm:"column:date"`
	Title  string `gorm:"column:title"`
	Body   string `gorm:"column:body"`
	Tags   string `gorm:"column:tags"`
}

func (oldItem) TableName() string { return "items" }

// oldItemChange represents the pre-kin-core item_change row.
type oldItemChange struct {
	ID            uint      `gorm:"column:id"`
	UserID        string    `gorm:"column:user_id"`
	Date          string    `gorm:"column:date"`
	OperationType string    `gorm:"column:operation_type"`
	Timestamp     time.Time `gorm:"column:timestamp"`
	ItemUserID    string    `gorm:"column:item_user_id"`
	ItemDate      string    `gorm:"column:item_date"`
	ItemTitle     string    `gorm:"column:item_title"`
	ItemBody      string    `gorm:"column:item_body"`
	ItemTags      string    `gorm:"column:item_tags"`
	Metadata      string    `gorm:"column:metadata"`
}

func (oldItemChange) TableName() string { return "item_changes" }

// oldOrphanIgnore represents the pre-kin-core orphan_ignore row.
type oldOrphanIgnore struct {
	ID       uint   `gorm:"column:id"`
	UserID   string `gorm:"column:user_id"`
	Filename string `gorm:"column:filename"`
}

func (oldOrphanIgnore) TableName() string { return "orphan_ignores" }

// runMigrationIfNeeded detects old schema (user_id column in items) and migrates to family-based schema.
func runMigrationIfNeeded(log *slog.Logger, db *gorm.DB, cfg *config.Config) error {
	// Check if old items table has user_id column (old schema)
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('items') WHERE name='user_id'").Scan(&count)
	if count == 0 {
		return nil // already migrated or fresh DB
	}

	log.Info("Detected old schema, running kin-core migration")

	// Read all old users
	var oldUsers []oldUser
	if err := db.Find(&oldUsers).Error; err != nil {
		return fmt.Errorf("failed to read old users: %w", err)
	}

	// Build userID → userMapping
	mappings := make(map[string]userMapping, len(oldUsers))
	for _, u := range oldUsers {
		mappings[u.ID] = userMapping{
			oldID:    u.ID,
			familyID: uuid.New(),
		}
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Step 1: create families table and populate
		if err := tx.Exec(`CREATE TABLE IF NOT EXISTS families (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			name TEXT NOT NULL
		)`).Error; err != nil {
			return fmt.Errorf("create families table: %w", err)
		}

		for _, u := range oldUsers {
			m := mappings[u.ID]
			if err := tx.Exec(
				"INSERT OR IGNORE INTO families (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
				m.familyID.String(), u.Login, time.Now(), time.Now(),
			).Error; err != nil {
				return fmt.Errorf("insert family for user %s: %w", u.Login, err)
			}
		}
		log.Info("Families created", "count", len(oldUsers))

		// Step 2: recreate users table with kin-core schema
		if err := tx.Exec("ALTER TABLE users RENAME TO users_old").Error; err != nil {
			return fmt.Errorf("rename users table: %w", err)
		}
		if err := tx.Exec(`CREATE TABLE users (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			family_id TEXT NOT NULL,
			start_date DATETIME
		)`).Error; err != nil {
			return fmt.Errorf("create new users table: %w", err)
		}
		for _, u := range oldUsers {
			m := mappings[u.ID]
			if err := tx.Exec(
				"INSERT INTO users (id, username, password_hash, family_id, start_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				u.ID, u.Login, u.HashedPassword, m.familyID.String(), u.StartDate, time.Now(), time.Now(),
			).Error; err != nil {
				return fmt.Errorf("insert migrated user %s: %w", u.Login, err)
			}
		}
		log.Info("Users migrated", "count", len(oldUsers))

		// Step 3: migrate items
		var oldItems []oldItem
		if err := tx.Raw("SELECT user_id, date, title, body, tags FROM items").Scan(&oldItems).Error; err != nil {
			return fmt.Errorf("read old items: %w", err)
		}
		if err := tx.Exec("ALTER TABLE items RENAME TO items_backup").Error; err != nil {
			return fmt.Errorf("rename items table: %w", err)
		}
		if err := tx.Exec(`CREATE TABLE items (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			family_id TEXT NOT NULL,
			date TEXT NOT NULL,
			title TEXT,
			body TEXT,
			tags TEXT
		)`).Error; err != nil {
			return fmt.Errorf("create new items table: %w", err)
		}
		if err := tx.Exec("CREATE UNIQUE INDEX idx_family_date ON items(family_id, date)").Error; err != nil {
			return fmt.Errorf("create items unique index: %w", err)
		}
		for _, item := range oldItems {
			m, ok := mappings[item.UserID]
			if !ok {
				log.Warn("Skipping item with unknown user_id", "user_id", item.UserID, "date", item.Date)
				continue
			}
			if err := tx.Exec(
				"INSERT INTO items (id, family_id, date, title, body, tags, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				uuid.New().String(), m.familyID.String(), item.Date, item.Title, item.Body, item.Tags, time.Now(), time.Now(),
			).Error; err != nil {
				return fmt.Errorf("insert migrated item %s: %w", item.Date, err)
			}
		}
		log.Info("Items migrated", "count", len(oldItems))

		// Step 4: migrate item_changes
		var oldChanges []oldItemChange
		if err := tx.Raw("SELECT id, user_id, date, operation_type, timestamp, item_user_id, item_date, item_title, item_body, item_tags, metadata FROM item_changes").Scan(&oldChanges).Error; err != nil {
			return fmt.Errorf("read old item_changes: %w", err)
		}
		if err := tx.Exec("ALTER TABLE item_changes RENAME TO item_changes_backup").Error; err != nil {
			return fmt.Errorf("rename item_changes table: %w", err)
		}
		if err := tx.Exec(`CREATE TABLE item_changes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			family_id TEXT NOT NULL,
			date TEXT NOT NULL,
			operation_type TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			item_id TEXT,
			item_family_id TEXT,
			item_created_at DATETIME,
			item_updated_at DATETIME,
			item_deleted_at DATETIME,
			item_date TEXT,
			item_title TEXT,
			item_body TEXT,
			item_tags TEXT,
			metadata TEXT
		)`).Error; err != nil {
			return fmt.Errorf("create new item_changes table: %w", err)
		}
		for _, ch := range oldChanges {
			m, ok := mappings[ch.UserID]
			if !ok {
				log.Warn("Skipping item_change with unknown user_id", "user_id", ch.UserID)
				continue
			}
			itemFamilyID := ""
			if im, ok2 := mappings[ch.ItemUserID]; ok2 {
				itemFamilyID = im.familyID.String()
			}
			if err := tx.Exec(
				`INSERT INTO item_changes (family_id, date, operation_type, timestamp, item_family_id, item_date, item_title, item_body, item_tags, metadata)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				m.familyID.String(), ch.Date, ch.OperationType, ch.Timestamp,
				itemFamilyID, ch.ItemDate, ch.ItemTitle, ch.ItemBody, ch.ItemTags, ch.Metadata,
			).Error; err != nil {
				return fmt.Errorf("insert migrated item_change: %w", err)
			}
		}
		log.Info("Item changes migrated", "count", len(oldChanges))

		// Step 5: migrate orphan_ignores
		var oldOrphans []oldOrphanIgnore
		if err := tx.Raw("SELECT id, user_id, filename FROM orphan_ignores").Scan(&oldOrphans).Error; err != nil {
			return fmt.Errorf("read old orphan_ignores: %w", err)
		}
		if err := tx.Exec("ALTER TABLE orphan_ignores RENAME TO orphan_ignores_backup").Error; err != nil {
			return fmt.Errorf("rename orphan_ignores table: %w", err)
		}
		if err := tx.Exec(`CREATE TABLE orphan_ignores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			family_id TEXT,
			filename TEXT,
			UNIQUE(family_id, filename)
		)`).Error; err != nil {
			return fmt.Errorf("create new orphan_ignores table: %w", err)
		}
		for _, o := range oldOrphans {
			m, ok := mappings[o.UserID]
			if !ok {
				log.Warn("Skipping orphan_ignore with unknown user_id", "user_id", o.UserID)
				continue
			}
			if err := tx.Exec(
				"INSERT OR IGNORE INTO orphan_ignores (family_id, filename) VALUES (?, ?)",
				m.familyID.String(), o.Filename,
			).Error; err != nil {
				return fmt.Errorf("insert migrated orphan_ignore: %w", err)
			}
		}
		log.Info("Orphan ignores migrated", "count", len(oldOrphans))

		// Step 6: rename asset directories from user_id to family_id
		migrateAssetDirs(log, cfg, mappings)

		return nil
	})
}

// migrateAssetDirs renames diary-assets/<user_id> → diary-assets/<family_id>
func migrateAssetDirs(log *slog.Logger, cfg *config.Config, mappings map[string]userMapping) {
	assetsBase := filepath.Join(cfg.DataPath, config.AssetsDirName)
	for _, m := range mappings {
		oldPath := filepath.Join(assetsBase, m.oldID)
		newPath := filepath.Join(assetsBase, m.familyID.String())
		if _, err := os.Stat(oldPath); os.IsNotExist(err) {
			continue // no assets for this user
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			log.Warn("Failed to rename asset dir", "from", oldPath, "to", newPath, "error", err)
		} else {
			log.Info("Renamed asset dir", "from", oldPath, "to", newPath)
		}
	}
}

