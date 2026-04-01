package tasks

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/ya-breeze/diary.be/pkg/checker"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
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
	Issues      []checker.Issue
	LastChecked time.Time
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

	result := &UserResult{Issues: issues, LastChecked: time.Now()}
	t.mu.Lock()
	t.results[userID] = result
	t.mu.Unlock()
	return result, nil
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
		t.results[userID] = &UserResult{
			Issues:      byUser[userID],
			LastChecked: now,
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
