package checker

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
)

// Issue represents a single problem found by a check.
type Issue struct {
	Check   string `json:"check"`
	UserID  string `json:"userID"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Fixable bool   `json:"fixable"`
	fix     func() error
}

// Fix runs the automated repair for this issue.
func (i *Issue) Fix() error {
	if i.fix == nil {
		return nil
	}
	return i.fix()
}

// Check is the interface every health check must implement.
type Check interface {
	Name() string
	Run(db database.Storage, cfg *config.Config, logger *slog.Logger) ([]Issue, error)
}

// Runner executes a set of checks and collects results.
type Runner struct {
	checks []Check
	logger *slog.Logger
}

func NewRunner(logger *slog.Logger, checks []Check) *Runner {
	return &Runner{checks: checks, logger: logger}
}

// RunForUser executes all checks and filters results to the given userID.
// If fix is true, automated fixes are applied only for that user's issues.
func (r *Runner) RunForUser(db database.Storage, cfg *config.Config, userID string, fix bool) ([]Issue, error) {
	// Always run without fix to avoid touching other users' data
	all, err := r.Run(db, cfg, false, io.Discard, false)
	if err != nil {
		return nil, err
	}
	var filtered []Issue
	for _, i := range all {
		if i.UserID == userID {
			filtered = append(filtered, i)
		}
	}
	if fix {
		for idx := range filtered {
			if !filtered[idx].Fixable {
				continue
			}
			if err := filtered[idx].Fix(); err != nil {
				r.logger.Error("Fix failed", "check", filtered[idx].Check, "path", filtered[idx].Path, "error", err)
				continue
			}
			r.logger.Info("Fixed", "check", filtered[idx].Check, "path", filtered[idx].Path)
			filtered[idx].Fixable = false
			filtered[idx].fix = nil
		}
	}
	return filtered, nil
}

// Run executes all checks. If fix is true, applies automated fixes.
// Returns the list of issues found and whether any issues remain after fixing.
func (r *Runner) Run(db database.Storage, cfg *config.Config, fix bool, w io.Writer, jsonFmt bool) ([]Issue, error) {
	var all []Issue

	for _, c := range r.checks {
		r.logger.Info("Running check", "check", c.Name())
		issues, err := c.Run(db, cfg, r.logger)
		if err != nil {
			return all, fmt.Errorf("check %q failed: %w", c.Name(), err)
		}
		all = append(all, issues...)
	}

	if fix {
		for i := range all {
			if !all[i].Fixable {
				continue
			}
			if err := all[i].Fix(); err != nil {
				r.logger.Error("Fix failed", "check", all[i].Check, "path", all[i].Path, "error", err)
				continue
			}
			r.logger.Info("Fixed", "check", all[i].Check, "path", all[i].Path)
			all[i].Fixable = false
			all[i].fix = nil
		}
	}

	if jsonFmt {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(all)
	} else {
		if len(all) == 0 {
			fmt.Fprintln(w, "No issues found.")
		}
		for _, issue := range all {
			status := ""
			if fix && !issue.Fixable {
				status = " [fixed]"
			} else if issue.Fixable {
				status = " [fixable]"
			}
			fmt.Fprintf(w, "[%s] user=%s path=%s: %s%s\n",
				issue.Check, issue.UserID, issue.Path, issue.Message, status)
		}
	}

	return all, nil
}
