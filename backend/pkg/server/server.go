package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/ya-breeze/diary.be/pkg/ai"
	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/api"
	"github.com/ya-breeze/diary.be/pkg/server/tasks"
	"github.com/ya-breeze/diary.be/pkg/server/webapp"
	kinauth "github.com/ya-breeze/kin-core/auth"
	"github.com/ya-breeze/kin-core/authdb"
	"gorm.io/gorm"
)

func Server(logger *slog.Logger, cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := database.NewStorage(logger, cfg)
	if err := storage.Open(); err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}

	_, finishChan, err := Serve(ctx, logger, storage, cfg)
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan
	logger.Info("Received signal. Shutting down server...")

	cancel()
	<-finishChan
	return nil
}

func createControllers(
	logger *slog.Logger, cfg *config.Config, db database.Storage,
	checkerTask *tasks.CheckerTask, suggester ai.Suggester,
) goserver.CustomControllers {
	return goserver.CustomControllers{
		AuthAPIService:   api.NewAuthAPIService(logger, db, cfg),
		FamilyAPIService: api.NewFamilyAPIService(logger, db),
		UserAPIService:   api.NewUserAPIService(logger, db),
		AssetsAPIService: api.NewAssetsAPIService(logger, cfg),
		HealthAPIService: api.NewHealthAPIServiceImpl(checkerTask),
		ItemsAPIService:  api.NewItemsAPIService(logger, db, suggester),
		SyncAPIService:   api.NewSyncAPIService(logger, db),
	}
}

func Serve(
	ctx context.Context, logger *slog.Logger,
	storage database.Storage, cfg *config.Config,
) (net.Addr, chan int, error) {
	commit := func() string {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					return setting.Value
				}
			}
		}
		return ""
	}()
	logger.Info("Built from git commit: " + commit)

	if cfg.JWTSecret == "" {
		logger.Warn("JWT secret is not set. Creating random secret...")
		cfg.JWTSecret = auth.GenerateRandomString(32)
	}

	logger.Info("Starting Diary server...")

	gormDB := storage.GetDB()

	// Seed users from DIARY_SEED_USERS (format: "Family:Username:Password,...")
	if cfg.SeedUsers != "" {
		logger.Info("Seeding users...")
		for entry := range strings.SplitSeq(cfg.SeedUsers, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			if err := upsertSeedUser(storage, entry, logger); err != nil {
				return nil, nil, fmt.Errorf("failed to seed user %q: %w", entry, err)
			}
		}
	} else {
		logger.Info("No seed users defined in configuration")
	}

	// Start token cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				authdb.CleanupExpiredBlacklist(gormDB)
				authdb.CleanupExpiredRefreshTokens(gormDB)
			}
		}
	}()

	// Start background health-check task
	checkerTask := tasks.NewCheckerTask(logger, storage, cfg)
	checkerTask.Start(ctx)

	// Start background backup task
	backupTask := tasks.NewBackupTask(logger, cfg)
	backupTask.Start(ctx)

	// Construct the AI tag suggester (disabled gracefully if GEMINI_API_KEY unset)
	suggester, err := ai.NewSuggester(ctx, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AI suggester: %w", err)
	}

	// Create controllers
	controllers := createControllers(logger, cfg, storage, checkerTask, suggester)

	// Add extra routers
	extraRouters := []goserver.Router{webapp.NewWebAppRouter(controllers, commit, logger, cfg, storage, gormDB)}
	extraRouters = append(extraRouters, api.NewAssetsBatchRouter(logger, cfg))
	extraRouters = append(extraRouters, api.NewCustomAuthAPIController(controllers.AuthAPIService, logger, cfg, storage, gormDB))

	return goserver.Serve(ctx, logger, cfg,
		controllers,
		extraRouters,
		createMiddlewares(logger, cfg, gormDB)...)
}

// upsertSeedUser creates or updates a user from a "Family:Username:Password" entry.
func upsertSeedUser(storage database.Storage, entry string, logger *slog.Logger) error {
	tokens := strings.Split(entry, ":")
	if len(tokens) != 3 {
		return fmt.Errorf("invalid seed user format %q, expected Family:Username:Password", entry)
	}
	familyName, username, password := tokens[0], tokens[1], tokens[2]

	// Ensure family exists
	family, err := storage.GetFamilyByName(familyName)
	if err != nil {
		family, err = storage.CreateFamily(familyName)
		if err != nil {
			return fmt.Errorf("failed to create family %q: %w", familyName, err)
		}
		logger.Info("Created family", "name", familyName, "id", family.ID)
	}

	// Upsert user
	existing, err := storage.GetUserByUsername(username)
	if err != nil {
		// User doesn't exist — create
		hash, hashErr := kinauth.HashPassword(password)
		if hashErr != nil {
			return fmt.Errorf("failed to hash password for %q: %w", username, hashErr)
		}
		user, createErr := storage.CreateUser(username, hash, family.ID)
		if createErr != nil {
			return fmt.Errorf("failed to create user %q: %w", username, createErr)
		}
		logger.Info("Created seed user", "username", username, "id", user.ID)
	} else {
		// User exists — update password
		hash, hashErr := kinauth.HashPassword(password)
		if hashErr != nil {
			return fmt.Errorf("failed to hash password for %q: %w", username, hashErr)
		}
		existing.PasswordHash = hash
		if putErr := storage.PutUser(existing); putErr != nil {
			return fmt.Errorf("failed to update user %q: %w", username, putErr)
		}
		logger.Info("Updated seed user password", "username", username)
	}

	return nil
}

func createMiddlewares(logger *slog.Logger, cfg *config.Config, gormDB *gorm.DB) []mux.MiddlewareFunc {
	rateLimiterStore := NewRateLimiterStore()
	return []mux.MiddlewareFunc{
		RateLimitMiddleware(logger, rateLimiterStore, cfg.DisableRateLimit),
		AuthMiddleware(logger, cfg, gormDB),
	}
}
