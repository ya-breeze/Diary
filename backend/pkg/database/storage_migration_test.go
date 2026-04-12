package database

import (
	"encoding/base64"
	"log/slog"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	kinauth "github.com/ya-breeze/kin-core/auth"

	"github.com/ya-breeze/diary.be/pkg/config"
)

// TestMigrationPasswordCompat verifies that users migrated from the old schema
// (which stored base64(bcrypt) passwords) can still log in after migration.
// The old Diary auth decoded base64 before calling bcrypt.CompareHashAndPassword.
// kin-core's VerifyPassword expects raw bcrypt strings, so the migration must
// strip the base64 encoding layer.
func TestMigrationPasswordCompat(t *testing.T) {
	logger := slog.Default()

	db, err := openSqlite(logger, ":memory:", false)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Create the OLD schema — users table with 'login' column triggers migration detection.
	// Items table must also exist (migration reads it).
	oldTables := []string{
		`CREATE TABLE users (id TEXT PRIMARY KEY, login TEXT UNIQUE, hashed_password TEXT,
			start_date DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE items (id TEXT PRIMARY KEY, user_id TEXT, date TEXT,
			title TEXT, body TEXT, tags TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE item_changes (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id TEXT,
			date TEXT, operation_type TEXT, timestamp DATETIME,
			item_user_id TEXT, item_date TEXT, item_title TEXT, item_body TEXT, item_tags TEXT,
			metadata TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE orphan_ignores (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id TEXT,
			filename TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
	}
	for _, ddl := range oldTables {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("failed to create old table: %v\nDDL: %s", err, ddl)
		}
	}

	// Simulate old password storage: bcrypt the password, then base64-encode it
	plainPassword := "diary-user-password"
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to bcrypt password: %v", err)
	}
	oldStoredHash := base64.StdEncoding.EncodeToString(bcryptHash) // old format: base64(bcrypt)

	now := time.Now()
	if err := db.Exec(`INSERT INTO users (id, login, hashed_password, start_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"11111111-1111-1111-1111-111111111111", "diary@example.com", oldStoredHash, now, now, now,
	).Error; err != nil {
		t.Fatalf("failed to insert old user: %v", err)
	}

	// Run kin-core migration
	cfg := &config.Config{DataPath: t.TempDir()}
	if err := runMigrationIfNeeded(logger, db, cfg); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// After migration, find the user and verify the password works with kin-core
	var passwordHash string
	if err := db.Raw("SELECT password_hash FROM users WHERE username = ?", "diary@example.com").
		Scan(&passwordHash).Error; err != nil || passwordHash == "" {
		t.Fatalf("migrated user not found or empty hash: %v", err)
	}

	// kin-core VerifyPassword must work with the migrated hash
	if !kinauth.VerifyPassword(plainPassword, passwordHash) {
		t.Errorf("VerifyPassword failed after migration — existing users cannot log in")
	}

	// Wrong password must still fail
	if kinauth.VerifyPassword("wrong-password", passwordHash) {
		t.Errorf("VerifyPassword accepted wrong password")
	}
}
