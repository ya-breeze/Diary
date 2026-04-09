package webapp

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	kinauth "github.com/ya-breeze/kin-core/auth"
	"github.com/ya-breeze/kin-core/authdb"
	kincookies "github.com/ya-breeze/kin-core/cookies"
	"github.com/ya-breeze/kin-core/middleware"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 365 * 24 * time.Hour
)

// isValidRedirectURL validates that a redirect URL is safe to use
// It prevents open redirect vulnerabilities by ensuring the URL is internal
func isValidRedirectURL(redirectURL string) bool {
	if redirectURL == "" {
		return false
	}

	// Parse the URL
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return false
	}

	// Must be a relative URL (no scheme, no host)
	if parsedURL.Scheme != "" || parsedURL.Host != "" {
		return false
	}

	// Must start with "/" but not "//" (to prevent protocol-relative URLs)
	if !strings.HasPrefix(redirectURL, "/") || strings.HasPrefix(redirectURL, "//") {
		return false
	}

	// Additional security: prevent URLs that could be interpreted as external
	// Block URLs with backslashes, which could be used in some browsers
	if strings.Contains(redirectURL, "\\") {
		return false
	}

	// Block URLs with encoded characters that could bypass validation
	if strings.Contains(redirectURL, "%2F") || strings.Contains(redirectURL, "%5C") {
		return false
	}

	return true
}

func (r *WebAppRouter) loginHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	username := req.Form.Get("username")
	password := req.Form.Get("password")
	redirectURL := req.Form.Get("redirect")

	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Timing-safe credential verification
	hash := kinauth.DummyHash
	user, err := r.db.GetUserByUsername(username)
	if err == nil {
		hash = user.PasswordHash
	}
	if !kinauth.VerifyPassword(password, hash) || err != nil {
		r.logger.Warn("Authentication failed", "username", username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	familyID := user.FamilyID
	accessToken, err := kinauth.GenerateAccessToken(user.ID, &familyID, []byte(r.cfg.JWTSecret), accessTokenTTL)
	if err != nil {
		r.logger.Error("Failed to generate access token", "error", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}
	rt, rtErr := authdb.CreateRefreshToken(r.gormDB, user.ID, refreshTokenTTL)
	if rtErr != nil {
		r.logger.Error("Failed to create refresh token", "error", rtErr)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	kincookies.SetAccessCookie(w, accessToken, int(accessTokenTTL.Seconds()), r.cookieCfg)
	kincookies.SetRefreshCookie(w, rt.Token, int(refreshTokenTTL.Seconds()), r.cookieCfg)

	r.logger.Info("User logged in", "username", username, "familyID", familyID)

	// Determine redirect destination with security validation
	destination := "/"
	if isValidRedirectURL(redirectURL) {
		destination = redirectURL
	}

	http.Redirect(w, req, destination, http.StatusSeeOther)
}

func (r *WebAppRouter) logoutHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Blacklist current access token and revoke refresh token
	if tokenStr := kincookies.GetAccessToken(req); tokenStr != "" {
		if claims, err := kinauth.ParseToken(tokenStr, []byte(r.cfg.JWTSecret)); err == nil {
			_ = authdb.BlacklistToken(r.gormDB, tokenStr, claims.ExpiresAt.Time)
		}
	}
	if rtStr := kincookies.GetRefreshToken(req); rtStr != "" {
		_ = authdb.RevokeRefreshToken(r.gormDB, rtStr)
	}

	kincookies.ClearAuthCookies(w, r.cookieCfg)

	tmpl, err := r.loadTemplates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := utils.CreateTemplateData(req, "login")

	// Check if there's a redirect parameter in the logout request
	redirectURL := req.URL.Query().Get("redirect")
	if isValidRedirectURL(redirectURL) {
		data["RedirectURL"] = redirectURL
	}

	if err := tmpl.ExecuteTemplate(w, "login.tpl", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetFamilyIDFromCookie validates the kin_access cookie and returns the familyID.
func (r *WebAppRouter) GetFamilyIDFromCookie(req *http.Request) (uuid.UUID, int, error) {
	claims, err := middleware.ValidateRequest(req, r.kinCfg)
	if err != nil {
		r.logger.Warn("Invalid or missing access token", "error", err)
		return uuid.Nil, http.StatusUnauthorized, err
	}
	if claims.FamilyID == nil {
		return uuid.Nil, http.StatusUnauthorized, fmt.Errorf("no family in token")
	}
	r.logger.Info("Request authenticated", "userID", claims.UserID, "familyID", claims.FamilyID)
	return *claims.FamilyID, http.StatusOK, nil
}

// ValidateFamilyID is the template-aware auth check used by webapp handlers.
func (r *WebAppRouter) ValidateFamilyID(
	tmpl *template.Template, w http.ResponseWriter, req *http.Request,
) (uuid.UUID, error) {
	familyID, statusCode, err := r.GetFamilyIDFromCookie(req)
	if err != nil {
		// Capture the current request URL for redirect after login
		redirectURL := req.URL.String()

		data := map[string]any{
			"RedirectURL": redirectURL,
		}

		w.WriteHeader(statusCode)

		if errTmpl := tmpl.ExecuteTemplate(w, "login.tpl", data); errTmpl != nil {
			r.logger.Warn("failed to execute login template", "error", errTmpl)
			http.Error(w, errTmpl.Error(), http.StatusInternalServerError)
		}
		return uuid.Nil, fmt.Errorf("failed to get family ID from cookie: %w", err)
	}

	return familyID, nil
}
