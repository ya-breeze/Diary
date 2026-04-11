package database

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/kin-core/authdb"
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

// runMigrationIfNeeded detects old schema (login column in users) and migrates to family-based schema.
func runMigrationIfNeeded(log *slog.Logger, db *gorm.DB, cfg *config.Config) error {
	// Check if users table has 'login' column (old schema used Login, new schema uses Username)
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='login'").Scan(&count)
	if count == 0 {
		return nil // already migrated or fresh DB
	}

	log.Info("Detected old schema, running kin-core migration")

	// Step 1: Read all old data before renaming anything
	var oldUsers []oldUser
	if err := db.Find(&oldUsers).Error; err != nil {
		return fmt.Errorf("failed to read old users: %w", err)
	}

	// Build userID → userMapping (each old user gets a new family)
	mappings := make(map[string]userMapping, len(oldUsers))
	for _, u := range oldUsers {
		mappings[u.ID] = userMapping{
			oldID:    u.ID,
			familyID: uuid.New(),
		}
	}

	var oldItems []oldItem
	if err := db.Raw("SELECT user_id, date, title, body, tags FROM items").Scan(&oldItems).Error; err != nil {
		return fmt.Errorf("failed to read old items: %w", err)
	}

	var oldChanges []oldItemChange
	if err := db.Raw("SELECT id, user_id, date, operation_type, timestamp, item_user_id, item_date, item_title, item_body, item_tags, metadata FROM item_changes").Scan(&oldChanges).Error; err != nil {
		return fmt.Errorf("failed to read old item_changes: %w", err)
	}

	var oldOrphans []oldOrphanIgnore
	if err := db.Raw("SELECT id, user_id, filename FROM orphan_ignores").Scan(&oldOrphans).Error; err != nil {
		return fmt.Errorf("failed to read old orphan_ignores: %w", err)
	}

	log.Info("Read old data", "users", len(oldUsers), "items", len(oldItems), "changes", len(oldChanges), "orphans", len(oldOrphans))

	// Step 2: Drop old tables entirely (data is already in memory).
	// This avoids index name collisions when GORM creates new tables.
	for _, tbl := range []string{"orphan_ignores", "item_changes", "items", "users", "families"} {
		// DROP TABLE IF EXISTS (some tables may not exist in all prod versions)
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tbl)).Error; err != nil {
			return fmt.Errorf("drop table %s: %w", tbl, err)
		}
	}

	// Step 3: Let GORM create new tables with the correct schema (matches model exactly)
	if err := db.AutoMigrate(
		&models.Family{},
		&models.User{},
		&models.Item{},
		&models.ItemChange{},
		&models.OrphanIgnore{},
		&authdb.RefreshToken{},
		&authdb.BlacklistedToken{},
	); err != nil {
		return fmt.Errorf("auto-migrate new schema: %w", err)
	}
	// Add composite unique index on items(family_id, date) — not possible via field tags
	// because FamilyID is in embedded TenantModel (kin-core).
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_items_family_date ON items(family_id, date)").Error; err != nil {
		return fmt.Errorf("create idx_items_family_date: %w", err)
	}
	log.Info("New schema created")

	// Step 4: Insert migrated data into new tables
	now := time.Now()

	// Families: one per user
	for _, u := range oldUsers {
		m := mappings[u.ID]
		family := models.Family{}
		family.ID = m.familyID
		family.Name = u.Login
		family.CreatedAt = now
		family.UpdatedAt = now
		if err := db.Create(&family).Error; err != nil {
			return fmt.Errorf("insert family for user %s: %w", u.Login, err)
		}
	}
	log.Info("Families inserted", "count", len(oldUsers))

	// Users: remap login→username, hashed_password→password_hash, add family_id
	for _, u := range oldUsers {
		m := mappings[u.ID]
		user := models.User{}
		user.ID = uuid.MustParse(u.ID)
		user.Username = u.Login
		user.PasswordHash = u.HashedPassword
		user.FamilyID = m.familyID
		user.StartDate = u.StartDate
		user.CreatedAt = now
		user.UpdatedAt = now
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("insert migrated user %s: %w", u.Login, err)
		}
	}
	log.Info("Users migrated", "count", len(oldUsers))

	// Items: remap user_id → family_id, generate new UUID id
	for _, item := range oldItems {
		m, ok := mappings[item.UserID]
		if !ok {
			log.Warn("Skipping item with unknown user_id", "user_id", item.UserID, "date", item.Date)
			continue
		}
		newItem := models.Item{}
		newItem.ID = uuid.New()
		newItem.FamilyID = m.familyID
		newItem.Date = item.Date
		newItem.Title = item.Title
		newItem.Body = item.Body
		newItem.Tags = models.StringList([]string{})
		if item.Tags != "" {
			// Tags stored as JSON array: ["tag1","tag2"]
			newItem.Tags = models.StringList([]string{}) // will be set below via raw
		}
		newItem.CreatedAt = now
		newItem.UpdatedAt = now
		if err := db.Exec(
			"INSERT INTO items (id, family_id, date, title, body, tags, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			newItem.ID.String(), m.familyID.String(), item.Date, item.Title, item.Body, item.Tags, now, now,
		).Error; err != nil {
			return fmt.Errorf("insert migrated item %s: %w", item.Date, err)
		}
	}
	log.Info("Items migrated", "count", len(oldItems))

	// Item changes
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
		if err := db.Exec(
			`INSERT INTO item_changes (family_id, date, operation_type, timestamp, item_family_id, item_date, item_title, item_body, item_tags, metadata)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			m.familyID.String(), ch.Date, ch.OperationType, ch.Timestamp,
			itemFamilyID, ch.ItemDate, ch.ItemTitle, ch.ItemBody, ch.ItemTags, ch.Metadata,
		).Error; err != nil {
			return fmt.Errorf("insert migrated item_change: %w", err)
		}
	}
	log.Info("Item changes migrated", "count", len(oldChanges))

	// Orphan ignores
	for _, o := range oldOrphans {
		m, ok := mappings[o.UserID]
		if !ok {
			log.Warn("Skipping orphan_ignore with unknown user_id", "user_id", o.UserID)
			continue
		}
		if err := db.Exec(
			"INSERT OR IGNORE INTO orphan_ignores (family_id, filename) VALUES (?, ?)",
			m.familyID.String(), o.Filename,
		).Error; err != nil {
			return fmt.Errorf("insert migrated orphan_ignore: %w", err)
		}
	}
	log.Info("Orphan ignores migrated", "count", len(oldOrphans))

	// Step 5: rename asset directories from user_id to family_id
	migrateAssetDirs(log, cfg, mappings)

	log.Info("Migration complete")
	return nil
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
