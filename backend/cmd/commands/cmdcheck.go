package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ya-breeze/diary.be/pkg/checker"
	"github.com/ya-breeze/diary.be/pkg/database"
)

var allChecks = []checker.Check{
	checker.MimeCheck{},
	checker.OrphansCheck{},
	checker.RefsCheck{},
}

func CmdCheck() *cobra.Command {
	var fix bool
	var format string
	var checksFlag string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check storage health and optionally repair issues",
		Long: `Scans asset files and diary entries for inconsistencies.

Available checks:
  mime     - asset files with wrong extension (e.g. video saved as .jpg)
  orphans  - asset files not referenced by any diary entry
  refs     - diary entries referencing missing asset files

Exits with code 0 if no issues are found, 1 if issues exist.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, logger, err := createConfigAndLogger(cmd)
			if err != nil {
				return err
			}

			// Select checks
			var selected []checker.Check
			if checksFlag == "all" || checksFlag == "" {
				selected = allChecks
			} else {
				names := strings.Split(checksFlag, ",")
				nameSet := make(map[string]bool, len(names))
				for _, n := range names {
					nameSet[strings.TrimSpace(n)] = true
				}
				for _, c := range allChecks {
					if nameSet[c.Name()] {
						selected = append(selected, c)
					}
				}
				if len(selected) == 0 {
					return fmt.Errorf("no valid checks selected (available: mime, orphans, refs)")
				}
			}

			// Open DB
			db := database.NewStorage(logger, cfg)
			if err := db.Open(); err != nil {
				return fmt.Errorf("opening database: %w", err)
			}
			defer db.Close() //nolint:errcheck

			runner := checker.NewRunner(logger, selected)
			issues, err := runner.Run(db, cfg, fix, os.Stdout, format == "json")
			if err != nil {
				return err
			}

			if len(issues) > 0 {
				// After fixing, check if any issues remain unfixed
				remaining := 0
				for _, i := range issues {
					if i.Fixable() {
						remaining++
					}
				}
				if !fix || remaining > 0 {
					os.Exit(1)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "automatically repair fixable issues")
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	cmd.Flags().StringVar(&checksFlag, "checks", "all", "comma-separated list of checks to run (mime,orphans,refs) or 'all'")

	return cmd
}
