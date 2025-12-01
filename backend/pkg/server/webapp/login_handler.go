package webapp

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/utils"
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

func (r *WebAppRouter) setSessionToken(w http.ResponseWriter, req *http.Request, token string) error {
	session, err := r.cookies.Get(req, r.cfg.CookieName)
	if err != nil {
		return err
	}
	session.Values["token"] = token
	// Use configured Secure flag (defaults to true for production security)
	// Set to false only for local development without HTTPS
	session.Options.Secure = r.cfg.CookieSecure
	session.Options.SameSite = http.SameSiteLaxMode
	if err := session.Save(req, w); err != nil {
		return err
	}
	return nil
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

	// Use AuthAPIService to authenticate user
	authData := goserver.AuthData{
		Email:    username,
		Password: password,
	}

	response, err := r.authService.Authorize(req.Context(), authData)
	if err != nil {
		r.logger.Error("Authentication failed", "username", username, "error", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Check if authentication was successful
	if response.Code != 200 {
		r.logger.Warn("Authentication failed", "username", username, "status", response.Code)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Extract token from response
	authResponse, ok := response.Body.(goserver.Authorize200Response)
	if !ok {
		r.logger.Error("Invalid response type from auth service", "username", username)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	token := authResponse.Token

	// set JWT token in cookie
	if err := r.setSessionToken(w, req, token); err != nil {
		r.logger.Warn("failed to save session", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	c := &http.Cookie{
		Name:     r.cfg.CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
	http.SetCookie(w, c)

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

func (r *WebAppRouter) GetUserIDFromSession(req *http.Request) (string, int, error) {
	session, err := r.cookies.Get(req, r.cfg.CookieName)
	if err != nil {
		r.logger.Error("Failed to get session", "error", err)
		return "", http.StatusUnauthorized, err
	}

	token, ok := session.Values["token"].(string)
	if !ok {
		r.logger.Warn("failed to get token from session")
		return "", http.StatusUnauthorized, errors.New("token not found in session")
	}

	userID, err := auth.CheckJWT(token, r.cfg.Issuer, r.cfg.JWTSecret)
	if err != nil {
		r.logger.With("err", err).Warn("Invalid token")
		return "", http.StatusUnauthorized, err
	}

	// Log successful authentication with user ID from cookie
	r.logger.Info("Request authenticated", "userID", userID, "source", "cookie", "path", req.URL.Path, "method", req.Method)

	return userID, http.StatusOK, nil
}

func (r *WebAppRouter) ValidateUserID(
	tmpl *template.Template, w http.ResponseWriter, req *http.Request,
) (string, error) {
	userID, statusCode, err := r.GetUserIDFromSession(req)
	if err != nil {
		// Capture the current request URL for redirect after login
		redirectURL := req.URL.String()

		// Create template data with redirect URL
		data := map[string]any{
			"RedirectURL": redirectURL,
		}

		// Set the status code before writing the response
		w.WriteHeader(statusCode)

		if errTmpl := tmpl.ExecuteTemplate(w, "login.tpl", data); errTmpl != nil {
			r.logger.Warn("failed to execute login template", "error", errTmpl)
			http.Error(w, errTmpl.Error(), http.StatusInternalServerError)
		}
		return "", fmt.Errorf("failed to get user ID from session: %w", err)
	}

	return userID, nil
}
