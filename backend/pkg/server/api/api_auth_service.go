package api

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	kinauth "github.com/ya-breeze/kin-core/auth"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 365 * 24 * time.Hour
)

type AuthAPIServiceImpl struct {
	logger *slog.Logger
	db     database.Storage
	cfg    *config.Config
}

func NewAuthAPIService(logger *slog.Logger, db database.Storage, cfg *config.Config) goserver.AuthAPIService {
	return &AuthAPIServiceImpl{
		logger: logger,
		db:     db,
		cfg:    cfg,
	}
}

// Authorize validates credentials and returns a signed access token.
// Cookie setting (access + refresh) is handled by CustomAuthAPIController.
func (s *AuthAPIServiceImpl) Authorize(ctx context.Context, authData goserver.AuthData) (goserver.ImplResponse, error) {
	s.logger.Info("Authorize request", "email", authData.Email)

	// Timing-safe credential verification
	hash := kinauth.DummyHash
	user, err := s.db.GetUserByUsername(authData.Email)
	if err == nil {
		hash = user.PasswordHash
	}
	if !kinauth.VerifyPassword(authData.Password, hash) || err != nil {
		if errors.Is(err, database.ErrNotFound) {
			s.logger.Warn("User not found", "email", authData.Email)
		} else if err != nil {
			s.logger.Error("Failed to get user", "email", authData.Email, "error", err)
		} else {
			s.logger.Warn("Invalid password", "email", authData.Email)
		}
		return goserver.Response(401, nil), nil
	}

	familyID := user.FamilyID
	accessToken, err := kinauth.GenerateAccessToken(user.ID, &familyID, []byte(s.cfg.JWTSecret), accessTokenTTL)
	if err != nil {
		s.logger.Error("Failed to create access token", "error", err)
		return goserver.Response(500, nil), nil
	}

	s.logger.Info("User authenticated successfully", "email", authData.Email, "userID", user.ID)

	return goserver.Response(200, goserver.Authorize200Response{Token: accessToken}), nil
}
