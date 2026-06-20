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

	"github.com/google/uuid"
	"github.com/ya-breeze/diary.be/pkg/ai"
	"github.com/ya-breeze/diary.be/pkg/checker"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
)

// ErrInvalidInput is returned when caller-supplied values (filename, date) fail validation.
var ErrInvalidInput = errors.New("invalid input")

// noopWriter discards all output (used when running checks silently in background).
type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }

var _ io.Writer = noopWriter{}

// baseChecks are the always-available, dependency-free checks.
var baseChecks = []checker.Check{
	checker.MimeCheck{},
	checker.OrphansCheck{},
	checker.RefsCheck{},
}

// FamilyResult holds the last health check results for a single family.
type FamilyResult = UserResult

// UserResult holds the last health check results for a single family.
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
	results  map[uuid.UUID]*UserResult
	interval time.Duration
	checks   []checker.Check
}

func NewCheckerTask(
	logger *slog.Logger, db database.Storage, cfg *config.Config, suggester ai.Suggester,
) *CheckerTask {
	interval := 24 * time.Hour
	if cfg.HealthCheckInterval != "" {
		if d, err := time.ParseDuration(cfg.HealthCheckInterval); err == nil {
			interval = d
		} else {
			logger.Warn("Invalid health_check_interval, using 24h", "value", cfg.HealthCheckInterval, "error", err)
		}
	}
	checks := append([]checker.Check{}, baseChecks...)
	checks = append(checks, checker.UntaggedCheck{Suggester: suggester})
	return &CheckerTask{
		db:       db,
		cfg:      cfg,
		logger:   logger,
		results:  make(map[uuid.UUID]*UserResult),
		interval: interval,
		checks:   checks,
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

// GetIssues returns the stored results for a family, or nil if no check has run yet.
func (t *CheckerTask) GetIssues(familyID uuid.UUID) *UserResult {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.results[familyID]
}

// RunFix re-runs the named checks with fix=true for a specific family, then re-scans to get clean results.
// If checks is empty, all checks are run.
func (t *CheckerTask) RunFix(familyID uuid.UUID, checks []string) (*UserResult, error) {
	selected, err := t.selectChecks(checks)
	if err != nil {
		return nil, err
	}
	runner := checker.NewRunner(t.logger, selected)

	// First pass: apply fixes
	if _, err := runner.RunForFamily(t.db, t.cfg, familyID, true); err != nil {
		return nil, err
	}

	// Second pass: re-scan to get true current state (fixed issues should be gone)
	issues, err := runner.RunForFamily(t.db, t.cfg, familyID, false)
	if err != nil {
		return nil, err
	}

	ignored, _ := t.db.GetIgnoredOrphans(familyID)

	// Collect the names of checks we just ran so we can replace only those in the cache.
	selectedNames := make(map[string]bool, len(selected))
	for _, c := range selected {
		selectedNames[c.Name()] = true
	}

	t.mu.Lock()
	existing := t.results[familyID]
	var mergedIssues []checker.Issue
	if existing != nil {
		for _, i := range existing.Issues {
			if !selectedNames[i.Check] {
				mergedIssues = append(mergedIssues, i)
			}
		}
	}
	mergedIssues = append(mergedIssues, issues...)
	result := &UserResult{Issues: mergedIssues, LastChecked: time.Now(), IgnoredOrphans: ignored}
	t.results[familyID] = result
	t.mu.Unlock()
	return result, nil
}

// GetIgnoredOrphans returns the list of filenames the family has chosen to ignore.
func (t *CheckerTask) GetIgnoredOrphans(familyID uuid.UUID) ([]string, error) {
	return t.db.GetIgnoredOrphans(familyID)
}

// DeleteOrphan deletes a single orphaned asset file for the family then re-runs orphan checks.
func (t *CheckerTask) DeleteOrphan(familyID uuid.UUID, filename string) (*UserResult, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	familyDir := filepath.Join(t.cfg.DataPath, config.AssetsDirName, familyID.String())
	filePath := filepath.Join(familyDir, filename)
	if err := os.Remove(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("orphan %q not found on disk: %w", filename, database.ErrNotFound)
		}
		return nil, fmt.Errorf("deleting orphan %q: %w", filename, err)
	}
	t.logger.Info("Deleted orphan", "file", filename, "familyID", familyID)
	return t.refreshOrphansForFamily(familyID)
}

// AttachOrphan inserts a markdown image reference into a diary entry (creating it if needed).
func (t *CheckerTask) AttachOrphan(familyID uuid.UUID, filename, date string) (*UserResult, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	if err := validateDate(date); err != nil {
		return nil, err
	}
	item, err := t.db.GetItem(familyID, date)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, fmt.Errorf("getting item %s: %w", date, err)
		}
		// Create a new entry for that date
		item = &models.Item{Date: date}
	}
	item.Body += fmt.Sprintf("\n\n![%s](%s)", filename, filename)
	if err := t.db.PutItem(familyID, item); err != nil {
		return nil, fmt.Errorf("saving item %s: %w", date, err)
	}
	t.logger.Info("Attached orphan to entry", "file", filename, "date", date, "familyID", familyID)
	return t.refreshOrphansForFamily(familyID)
}

// IgnoreOrphan adds a filename to the family's ignore list.
func (t *CheckerTask) IgnoreOrphan(familyID uuid.UUID, filename string) (*UserResult, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	if err := t.db.AddIgnoredOrphan(familyID, filename); err != nil {
		return nil, fmt.Errorf("adding ignored orphan %q: %w", filename, err)
	}
	t.logger.Info("Ignored orphan", "file", filename, "familyID", familyID)
	return t.refreshOrphansForFamily(familyID)
}

// UnignoreOrphan removes a filename from the family's ignore list.
func (t *CheckerTask) UnignoreOrphan(familyID uuid.UUID, filename string) (*UserResult, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	if err := t.db.RemoveIgnoredOrphan(familyID, filename); err != nil {
		return nil, fmt.Errorf("removing ignored orphan %q: %w", filename, err)
	}
	t.logger.Info("Unignored orphan", "file", filename, "familyID", familyID)
	return t.refreshOrphansForFamily(familyID)
}

// refreshOrphansForFamily re-runs the orphans check for a single family and merges the fresh
// orphan results with whatever non-orphan issues are currently cached in memory.
func (t *CheckerTask) refreshOrphansForFamily(familyID uuid.UUID) (*UserResult, error) {
	runner := checker.NewRunner(t.logger, []checker.Check{checker.OrphansCheck{}})
	issues, err := runner.RunForFamily(t.db, t.cfg, familyID, false)
	if err != nil {
		return nil, err
	}
	ignored, _ := t.db.GetIgnoredOrphans(familyID)

	t.mu.Lock()
	existing := t.results[familyID]
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
	t.results[familyID] = result
	t.mu.Unlock()
	return result, nil
}

// validateFilename ensures a filename is safe (no path separators or traversal).
func validateFilename(filename string) error {
	if filename == "" || strings.ContainsAny(filename, "/\\") || strings.Contains(filename, "..") {
		return fmt.Errorf("invalid filename %q: %w", filename, ErrInvalidInput)
	}
	return nil
}

// validateDate ensures the date is a valid YYYY-MM-DD string.
func validateDate(date string) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("invalid date %q (expected YYYY-MM-DD): %w", date, ErrInvalidInput)
	}
	return nil
}

func (t *CheckerTask) runAll() {
	t.logger.Info("Running health checks")
	runner := checker.NewRunner(t.logger, t.checks)
	allIssues, err := runner.Run(t.db, t.cfg, false, noopWriter{}, false)
	if err != nil {
		t.logger.Error("Health check failed", "error", err)
		return
	}

	// Group issues by familyID
	byFamily := make(map[uuid.UUID][]checker.Issue)
	for _, issue := range allIssues {
		fid, err := uuid.Parse(issue.FamilyID)
		if err != nil {
			t.logger.Warn("Skipping issue with invalid familyID", "familyID", issue.FamilyID)
			continue
		}
		byFamily[fid] = append(byFamily[fid], issue)
	}

	now := time.Now()
	users, err := t.db.GetAllUsers()
	if err != nil {
		t.logger.Error("Health check: failed to get users", "error", err)
		return
	}
	// Deduplicate families (multiple users may share a family)
	seenFamilies := make(map[uuid.UUID]bool)
	t.mu.Lock()
	for _, user := range users {
		fid := user.FamilyID
		if seenFamilies[fid] {
			continue
		}
		seenFamilies[fid] = true
		ignored, _ := t.db.GetIgnoredOrphans(fid)
		t.results[fid] = &UserResult{
			Issues:         byFamily[fid],
			LastChecked:    now,
			IgnoredOrphans: ignored,
		}
		t.logger.Info("Health check complete", "familyID", fid, "issues", len(byFamily[fid]))
	}
	t.mu.Unlock()
}

func (t *CheckerTask) selectChecks(names []string) ([]checker.Check, error) {
	if len(names) == 0 {
		return t.checks, nil
	}
	known := make(map[string]checker.Check, len(t.checks))
	for _, c := range t.checks {
		known[c.Name()] = c
	}
	var selected []checker.Check
	for _, n := range names {
		c, ok := known[n]
		if !ok {
			return nil, fmt.Errorf("unknown check %q (available: mime, orphans, refs, untagged)", n)
		}
		selected = append(selected, c)
	}
	return selected, nil
}
