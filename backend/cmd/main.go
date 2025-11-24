//nolint:forbidigo // it's okay to use fmt in this file
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/ya-breeze/diary.be/cmd/commands"
	"github.com/ya-breeze/diary.be/pkg/config"
)

// Write to stdout in gray color
type grayWriter struct{}

func (w *grayWriter) Write(p []byte) (int, error) {
	return fmt.Fprint(os.Stdout, "\x1b[90m", string(p), "\x1b[0m")
}

func newRootCmd(cfgFile *string, logger *slog.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "geekbudget",
		Short: "GeekBudget is a personal finance manager",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cfg, err := config.InitiateConfig(*cfgFile)
			if err != nil {
				fmt.Printf("ERROR: %s", err)
				os.Exit(1)
			}
			cmd.SetContext(context.WithValue(cmd.Context(), commands.ConfigKey, cfg))
		},
	}

	rootCmd.PersistentFlags().StringVar(cfgFile, "config", "", "config file")
	rootCmd.AddCommand(
		commands.CmdUser(logger),
		commands.CmdServer(),
	)

	return rootCmd
}

func main() {
	var cfgFile string

	logger := slog.New(slog.NewJSONHandler(&grayWriter{}, nil))

	rootCmd := newRootCmd(&cfgFile, logger)
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
