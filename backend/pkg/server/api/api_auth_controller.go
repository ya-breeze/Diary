package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	kinauth "github.com/ya-breeze/kin-core/auth"
	"github.com/ya-breeze/kin-core/authdb"
	kincookies "github.com/ya-breeze/kin-core/cookies"
	"gorm.io/gorm"
)

// CustomAuthAPIController wraps the generated AuthAPIController to add kin-core cookie support.
type CustomAuthAPIController struct {
	service      goserver.AuthAPIServicer
	errorHandler goserver.ErrorHandler
	logger       *slog.Logger
	cfg          *config.Config
	db           database.Storage
	gormDB       *gorm.DB
	cookieCfg    kincookies.Config
}

// NewCustomAuthAPIController creates a custom auth controller with kin-core cookie support.
func NewCustomAuthAPIController(
	service goserver.AuthAPIServicer, logger *slog.Logger, cfg *config.Config,
	db database.Storage, gormDB *gorm.DB,
) *CustomAuthAPIController {
	return &CustomAuthAPIController{
		service:      service,
		errorHandler: goserver.DefaultErrorHandler,
		logger:       logger,
		cfg:          cfg,
		db:           db,
		gormDB:       gormDB,
		cookieCfg:    kincookies.Config{Secure: cfg.CookieSecure},
	}
}

// Routes returns all the api routes for the CustomAuthAPIController.
func (c *CustomAuthAPIController) Routes() goserver.Routes {
	return goserver.Routes{
		"Authorize": goserver.Route{
			Method:      strings.ToUpper("Post"),
			Pattern:     "/v1/authorize",
			HandlerFunc: c.Authorize,
		},
		"Logout": goserver.Route{
			Method:      strings.ToUpper("Post"),
			Pattern:     "/v1/logout",
			HandlerFunc: c.Logout,
		},
		"Refresh": goserver.Route{
			Method:      strings.ToUpper("Post"),
			Pattern:     "/v1/refresh",
			HandlerFunc: c.Refresh,
		},
	}
}

// Authorize validates credentials and sets kin-core access + refresh cookies.
func (c *CustomAuthAPIController) Authorize(w http.ResponseWriter, r *http.Request) {
	authDataParam := goserver.AuthData{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&authDataParam); err != nil {
		c.errorHandler(w, r, &goserver.ParsingError{Err: err}, nil)
		return
	}
	if err := goserver.AssertAuthDataRequired(authDataParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}

	result, err := c.service.Authorize(r.Context(), authDataParam)
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	if result.Code == 200 {
		authResponse, ok := result.Body.(goserver.Authorize200Response)
		if ok && authResponse.Token != "" {
			// Parse access token to get userID for refresh token creation
			claims, parseErr := kinauth.ParseToken(authResponse.Token, []byte(c.cfg.JWTSecret))
			if parseErr == nil {
				rt, rtErr := authdb.CreateRefreshToken(c.gormDB, claims.UserID, refreshTokenTTL)
				if rtErr != nil {
					c.logger.Warn("Failed to create refresh token", "error", rtErr)
				} else {
					kincookies.SetRefreshCookie(w, rt.Token, int(refreshTokenTTL.Seconds()), c.cookieCfg)
				}
			}
			kincookies.SetAccessCookie(w, authResponse.Token, int(accessTokenTTL.Seconds()), c.cookieCfg)
			c.logger.Info("Auth cookies set successfully")
		}
	}

	_ = goserver.EncodeJSONResponse(result.Body, &result.Code, w)
}

// Logout blacklists the access token, revokes the refresh token, and clears cookies.
func (c *CustomAuthAPIController) Logout(w http.ResponseWriter, r *http.Request) {
	if tokenStr := kincookies.GetAccessToken(r); tokenStr != "" {
		if claims, err := kinauth.ParseToken(tokenStr, []byte(c.cfg.JWTSecret)); err == nil {
			if err := authdb.BlacklistToken(c.gormDB, tokenStr, claims.ExpiresAt.Time); err != nil {
				c.logger.Warn("Failed to blacklist token", "error", err)
			}
		}
	}
	if rtStr := kincookies.GetRefreshToken(r); rtStr != "" {
		if err := authdb.RevokeRefreshToken(c.gormDB, rtStr); err != nil {
			c.logger.Warn("Failed to revoke refresh token", "error", err)
		}
	}

	kincookies.ClearAuthCookies(w, c.cookieCfg)
	c.logger.Info("User logged out, cookies cleared")
	w.WriteHeader(http.StatusOK)
}

// Refresh rotates the refresh token and issues a new access token.
func (c *CustomAuthAPIController) Refresh(w http.ResponseWriter, r *http.Request) {
	rtStr := kincookies.GetRefreshToken(r)
	if rtStr == "" {
		http.Error(w, "No refresh token", http.StatusUnauthorized)
		return
	}

	newRT, err := authdb.RotateRefreshToken(c.gormDB, rtStr, refreshTokenTTL)
	if err != nil {
		if err == authdb.ErrTokenCompromised {
			c.logger.Warn("Refresh token reuse detected — all sessions revoked")
			kincookies.ClearAuthCookies(w, c.cookieCfg)
			http.Error(w, "Token compromised", http.StatusUnauthorized)
			return
		}
		c.logger.Warn("Failed to rotate refresh token", "error", err)
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Look up the user to include familyID in the new access token
	user, err := c.db.GetUser(newRT.UserID)
	if err != nil {
		c.logger.Error("Failed to get user on refresh", "userID", newRT.UserID, "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	familyID := user.FamilyID
	accessToken, err := kinauth.GenerateAccessToken(newRT.UserID, &familyID, []byte(c.cfg.JWTSecret), accessTokenTTL)
	if err != nil {
		c.logger.Error("Failed to generate access token on refresh", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	kincookies.SetAccessCookie(w, accessToken, int(accessTokenTTL.Seconds()), c.cookieCfg)
	kincookies.SetRefreshCookie(w, newRT.Token, int(refreshTokenTTL.Seconds()), c.cookieCfg)
	c.logger.Info("Token refreshed", "userID", newRT.UserID)
	w.WriteHeader(http.StatusOK)
}
