package commands

import (
	"errors"
	"log/slog"
	"os"

	"github.com/dusted-go/logging/prettylog"
	"github.com/spf13/cobra"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/server"
	"golang.org/x/term"
)

type ContextKey string

const ConfigKey ContextKey = "config"

func CmdServer() *cobra.Command {
	res := &cobra.Command{
		Use:   "server",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, logger, err := createConfigAndLogger(cmd)
			if err != nil {
				return err
			}
			return server.Server(logger, cfg)
		},
	}

	return res
}

func createConfigAndLogger(cmd *cobra.Command) (*config.Config, *slog.Logger, error) {
	cfg, ok := cmd.Context().Value(ConfigKey).(*config.Config)
	if !ok {
		return nil, nil, errors.New("could not retrieve config from context")
	}

	var h slog.Handler
	if term.IsTerminal(int(os.Stdout.Fd())) {
		h = prettylog.NewHandler(&slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: false,
		})
	} else {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	logger := slog.New(h)
	logger.Info("Config loaded", "config", cfg)
	return cfg, logger, nil
}
