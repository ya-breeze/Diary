package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ya-breeze/diary.be/pkg/checker"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
)

// noopWriter discards all output (used when running checks silently in background).
type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }

var _ io.Writer = noopWriter{}

var allChecks = []checker.Check{
	checker.MimeCheck{},
	checker.OrphansCheck{},
	checker.RefsCheck{},
}

// UserResult holds the last health check results for a single user.
type UserResult struct {
	Issues         []checker.Issue
	LastChecked    time.Time
	IgnoredOrphans []string
}

// CheckerTask runs health checks periodically and stores results in memory.
type CheckerTask struct {
	db       database.Storage
	cfg      *config.Config
	logger   *slog.Logger
	mu       sync.RWMutex
	results  map[string]*UserResult // userID → result
	interval time.Duration
}

func NewCheckerTask(logger *slog.Logger, db database.Storage, cfg *config.Config) *CheckerTask {
	interval := 24 * time.Hour
	if cfg.HealthCheckInterval != "" {
		if d, err := time.ParseDuration(cfg.HealthCheckInterval); err == nil {
			interval = d
		} else {
			logger.Warn("Invalid health_check_interval, using 24h", "value", cfg.HealthCheckInterval, "error", err)
		}
	}
	return &CheckerTask{
		db:       db,
		cfg:      cfg,
		logger:   logger,
		results:  make(map[string]*UserResult),
		interval: interval,
	}
}

// Start launches the background goroutine. It runs once after a 30s delay, then on the ticker.
func (t *CheckerTask) Start(ctx context.Context) {
	go func() {
		select {
		case <-time.After(30 * time.Second):
		case <-ctx.Done():
			return
		}
		t.runAll()
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.runAll()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// GetIssues returns the stored results for a user, or nil if no check has run yet.
func (t *CheckerTask) GetIssues(userID string) *UserResult {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.results[userID]
}

// RunFix re-runs the named checks with fix=true for a specific user, then re-scans to get clean results.
// If checks is empty, all checks are run.
func (t *CheckerTask) RunFix(userID string, checks []string) (*UserResult, error) {
	selected, err := selectChecks(checks)
	if err != nil {
		return nil, err
	}
	runner := checker.NewRunner(t.logger, selected)

	// First pass: apply fixes
	if _, err := runner.RunForUser(t.db, t.cfg, userID, true); err != nil {
		return nil, err
	}

	// Second pass: re-scan to get true current state (fixed issues should be gone)
	issues, err := runner.RunForUser(t.db, t.cfg, userID, false)
	if err != nil {
		return nil, err
	}

	ignored, _ := t.db.GetIgnoredOrphans(userID)
	result := &UserResult{Issues: issues, LastChecked: time.Now(), IgnoredOrphans: ignored}
	t.mu.Lock()
	t.results[userID] = result
	t.mu.Unlock()
	return result, nil
}

// GetIgnoredOrphans returns the list of filenames the user has chosen to ignore.
func (t *CheckerTask) GetIgnoredOrphans(userID string) ([]string, error) {
	return t.db.GetIgnoredOrphans(userID)
}

// DeleteOrphan deletes a single orphaned asset file for the user then re-runs orphan checks.
func (t *CheckerTask) DeleteOrphan(userID, filename string) (*UserResult, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	userDir := filepath.Join(t.cfg.DataPath, config.AssetsDirName, userID)
	filePath := filepath.Join(userDir, filename)
	if err := os.Remove(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("orphan not found: %w", database.ErrNotFound)
		}
		return nil, fmt.Errorf("deleting orphan %q: %w", filename, err)
	}
	t.logger.Info("Deleted orphan", "file", filename, "userID", userID)
	return t.refreshOrphansForUser(userID)
}

// AttachOrphan inserts a markdown image reference into a diary entry (creating it if needed).
func (t *CheckerTask) AttachOrphan(userID, filename, date string) (*UserResult, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	item, err := t.db.GetItem(userID, date)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, fmt.Errorf("getting item %s: %w", date, err)
		}
		// Create a new entry for that date
		item = &models.Item{UserID: userID, Date: date}
	}
	item.Body += fmt.Sprintf("\n\n![%s](%s)", filename, filename)
	if err := t.db.PutItem(userID, item); err != nil {
		return nil, fmt.Errorf("saving item %s: %w", date, err)
	}
	t.logger.Info("Attached orphan to entry", "file", filename, "date", date, "userID", userID)
	return t.refreshOrphansForUser(userID)
}

// IgnoreOrphan adds a filename to the user's ignore list.
func (t *CheckerTask) IgnoreOrphan(userID, filename string) (*UserResult, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	if err := t.db.AddIgnoredOrphan(userID, filename); err != nil {
		return nil, fmt.Errorf("adding ignored orphan %q: %w", filename, err)
	}
	t.logger.Info("Ignored orphan", "file", filename, "userID", userID)
	return t.refreshOrphansForUser(userID)
}

// UnignoreOrphan removes a filename from the user's ignore list.
func (t *CheckerTask) UnignoreOrphan(userID, filename string) (*UserResult, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	if err := t.db.RemoveIgnoredOrphan(userID, filename); err != nil {
		return nil, fmt.Errorf("removing ignored orphan %q: %w", filename, err)
	}
	t.logger.Info("Unignored orphan", "file", filename, "userID", userID)
	return t.refreshOrphansForUser(userID)
}

// refreshOrphansForUser re-runs the orphans check and updates stored results.
func (t *CheckerTask) refreshOrphansForUser(userID string) (*UserResult, error) {
	runner := checker.NewRunner(t.logger, []checker.Check{checker.OrphansCheck{}})
	issues, err := runner.RunForUser(t.db, t.cfg, userID, false)
	if err != nil {
		return nil, err
	}
	ignored, _ := t.db.GetIgnoredOrphans(userID)

	t.mu.Lock()
	existing := t.results[userID]
	var mergedIssues []checker.Issue
	if existing != nil {
		for _, i := range existing.Issues {
			if i.Check != "orphans" {
				mergedIssues = append(mergedIssues, i)
			}
		}
	}
	mergedIssues = append(mergedIssues, issues...)
	result := &UserResult{Issues: mergedIssues, LastChecked: time.Now(), IgnoredOrphans: ignored}
	t.results[userID] = result
	t.mu.Unlock()
	return result, nil
}

// validateFilename ensures a filename is safe (no path separators or traversal).
func validateFilename(filename string) error {
	if filename == "" || strings.ContainsAny(filename, "/\\") || strings.Contains(filename, "..") {
		return fmt.Errorf("invalid filename %q", filename)
	}
	return nil
}

func (t *CheckerTask) runAll() {
	t.logger.Info("Running health checks")
	runner := checker.NewRunner(t.logger, allChecks)
	allIssues, err := runner.Run(t.db, t.cfg, false, noopWriter{}, false)
	if err != nil {
		t.logger.Error("Health check failed", "error", err)
		return
	}

	// Group issues by userID
	byUser := make(map[string][]checker.Issue)
	for _, issue := range allIssues {
		byUser[issue.UserID] = append(byUser[issue.UserID], issue)
	}

	now := time.Now()
	// Also mark users with no issues as checked
	users, err := t.db.GetAllUsers()
	if err != nil {
		t.logger.Error("Health check: failed to get users", "error", err)
		return
	}
	t.mu.Lock()
	for _, user := range users {
		userID := user.ID.String()
		ignored, _ := t.db.GetIgnoredOrphans(userID)
		t.results[userID] = &UserResult{
			Issues:         byUser[userID],
			LastChecked:    now,
			IgnoredOrphans: ignored,
		}
		t.logger.Info("Health check complete", "userID", userID, "issues", len(byUser[userID]))
	}
	t.mu.Unlock()
}

func selectChecks(names []string) ([]checker.Check, error) {
	if len(names) == 0 {
		return allChecks, nil
	}
	known := make(map[string]checker.Check, len(allChecks))
	for _, c := range allChecks {
		known[c.Name()] = c
	}
	var selected []checker.Check
	for _, n := range names {
		c, ok := known[n]
		if !ok {
			return nil, fmt.Errorf("unknown check %q (available: mime, orphans, refs)", n)
		}
		selected = append(selected, c)
	}
	return selected, nil
}
