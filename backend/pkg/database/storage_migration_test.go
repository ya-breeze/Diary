package database

import (
	"encoding/base64"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	kinauth "github.com/ya-breeze/kin-core/auth"
	"golang.org/x/crypto/bcrypt"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database/models"
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

// TestScrubBlankTags verifies that scrubBlankTags removes empty/whitespace entries
// while leaving valid tags untouched.
func TestScrubBlankTags(t *testing.T) {
	logger := slog.Default()
	db, err := openSqlite(logger, ":memory:", false)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := autoMigrateModels(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	familyID := uuid.New()

	cases := []struct {
		date        string
		tags        string // raw JSON, to simulate legacy data with blank tags
		pending     string
		wantTags    models.StringList
		wantPending models.StringList
	}{
		{
			date:        "2024-01-01",
			tags:        `[""]`,
			pending:     `["  "]`,
			wantTags:    models.StringList{},
			wantPending: models.StringList{},
		},
		{
			date:        "2024-01-02",
			tags:        `["family","","alisa"]`,
			pending:     `[]`,
			wantTags:    models.StringList{"family", "alisa"},
			wantPending: models.StringList{},
		},
		{
			date:        "2024-01-03",
			tags:        `["work"]`,
			pending:     `["beach"]`,
			wantTags:    models.StringList{"work"},
			wantPending: models.StringList{"beach"},
		},
	}

	// Insert via raw SQL to bypass the application-layer write filter so we can
	// seed the blank-tag patterns that existed in legacy data.
	now := time.Now().UTC().Format(time.RFC3339)
	for _, c := range cases {
		if err := db.Exec(
			`INSERT INTO items (id, family_id, date, title, tags, pending_tags, created_at, updated_at)
			 VALUES (?, ?, ?, 't', ?, ?, ?, ?)`,
			uuid.New().String(), familyID, c.date, c.tags, c.pending, now, now,
		).Error; err != nil {
			t.Fatalf("seed %s: %v", c.date, err)
		}
	}

	if err := scrubBlankTags(logger, db); err != nil {
		t.Fatalf("scrubBlankTags: %v", err)
	}

	for _, c := range cases {
		var got models.Item
		if err := db.Select("tags, pending_tags").Where("date = ? AND family_id = ?", c.date, familyID).First(&got).Error; err != nil {
			t.Fatalf("fetch %s: %v", c.date, err)
		}
		if len(got.Tags) != len(c.wantTags) {
			t.Errorf("%s tags: got %v, want %v", c.date, got.Tags, c.wantTags)
			continue
		}
		for i := range c.wantTags {
			if got.Tags[i] != c.wantTags[i] {
				t.Errorf("%s tags[%d]: got %q, want %q", c.date, i, got.Tags[i], c.wantTags[i])
			}
		}
		if len(got.PendingTags) != len(c.wantPending) {
			t.Errorf("%s pending_tags: got %v, want %v", c.date, got.PendingTags, c.wantPending)
			continue
		}
		for i := range c.wantPending {
			if got.PendingTags[i] != c.wantPending[i] {
				t.Errorf("%s pending_tags[%d]: got %q, want %q", c.date, i, got.PendingTags[i], c.wantPending[i])
			}
		}
	}
}

// TestNormalizeTagColumns verifies that NULL/empty/non-JSON tag columns are
// rewritten to '[]' while valid JSON arrays are left untouched.
func TestNormalizeTagColumns(t *testing.T) {
	logger := slog.Default()
	db, err := openSqlite(logger, ":memory:", false)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := autoMigrateModels(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	familyID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339)

	cases := []struct {
		date     string
		tags     string // raw column literal (may be non-JSON)
		nullTags bool
		want     string // expected raw column after normalization
	}{
		{date: "2024-02-01", tags: "''", want: "[]"},               // empty string
		{date: "2024-02-02", nullTags: true, want: "[]"},           // NULL
		{date: "2024-02-03", tags: "'garbage'", want: "[]"},        // non-JSON text
		{date: "2024-02-04", tags: `'["work"]'`, want: `["work"]`}, // valid, untouched
	}

	for _, c := range cases {
		tagsLiteral := c.tags
		if c.nullTags {
			tagsLiteral = "NULL"
		}
		if err := db.Exec(
			`INSERT INTO items (id, family_id, date, title, tags, pending_tags, created_at, updated_at)
			 VALUES (?, ?, ?, 't', `+tagsLiteral+`, '[]', ?, ?)`,
			uuid.New().String(), familyID, c.date, now, now,
		).Error; err != nil {
			t.Fatalf("seed %s: %v", c.date, err)
		}
	}

	if err := normalizeTagColumns(logger, db); err != nil {
		t.Fatalf("normalizeTagColumns: %v", err)
	}

	for _, c := range cases {
		var raw string
		if err := db.Raw(
			"SELECT tags FROM items WHERE date = ? AND family_id = ?", c.date, familyID,
		).Scan(&raw).Error; err != nil {
			t.Fatalf("fetch %s: %v", c.date, err)
		}
		if raw != c.want {
			t.Errorf("%s tags raw: got %q, want %q", c.date, raw, c.want)
		}
	}
}
