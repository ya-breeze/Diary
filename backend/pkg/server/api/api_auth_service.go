package api

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"

	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
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

// Authorize - validate user/password and return token
func (s *AuthAPIServiceImpl) Authorize(ctx context.Context, authData goserver.AuthData) (goserver.ImplResponse, error) {
	s.logger.Info("Authorize request", "email", authData.Email)

	// Get user ID by email
	userID, err := s.db.GetUserID(authData.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			s.logger.Warn("User not found", "email", authData.Email)
			return goserver.Response(401, nil), nil
		}
		s.logger.Error("Failed to get user ID", "email", authData.Email, "error", err)
		return goserver.Response(500, nil), nil
	}

	// Get user details
	user, err := s.db.GetUser(userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			s.logger.Warn("User not found", "userID", userID)
			return goserver.Response(401, nil), nil
		}
		s.logger.Error("Failed to get user", "userID", userID, "error", err)
		return goserver.Response(500, nil), nil
	}

	// Verify password - decode base64 encoded hash from database
	hashedPassword, err := base64.StdEncoding.DecodeString(user.HashedPassword)
	if err != nil {
		s.logger.Error("Failed to decode hashed password", "email", authData.Email, "error", err)
		return goserver.Response(500, nil), nil
	}

	if !auth.CheckPasswordHash([]byte(authData.Password), hashedPassword) {
		s.logger.Warn("Invalid password", "email", authData.Email)
		return goserver.Response(401, nil), nil
	}

	// Generate JWT token
	token, err := auth.CreateJWT(userID, s.cfg.Issuer, s.cfg.JWTSecret)
	if err != nil {
		s.logger.Error("Failed to create JWT token", "userID", userID, "error", err)
		return goserver.Response(500, nil), nil
	}

	s.logger.Info("User authenticated successfully", "email", authData.Email, "userID", userID)

	// Return token response
	response := goserver.Authorize200Response{
		Token: token,
	}

	return goserver.Response(200, response), nil
}
