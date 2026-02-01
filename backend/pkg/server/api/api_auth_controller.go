package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
)

// CustomAuthAPIController wraps the generated AuthAPIController to add cookie support
type CustomAuthAPIController struct {
	service      goserver.AuthAPIServicer
	errorHandler goserver.ErrorHandler
	logger       *slog.Logger
	cfg          *config.Config
	cookies      *sessions.CookieStore
}

// NewCustomAuthAPIController creates a custom auth controller with cookie support
func NewCustomAuthAPIController(
	service goserver.AuthAPIServicer, logger *slog.Logger, cfg *config.Config,
) *CustomAuthAPIController {
	return &CustomAuthAPIController{
		service:      service,
		errorHandler: goserver.DefaultErrorHandler,
		logger:       logger,
		cfg:          cfg,
		cookies:      sessions.NewCookieStore([]byte(cfg.SessionSecret)),
	}
}

// Routes returns all the api routes for the CustomAuthAPIController
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
	}
}

// setSessionToken sets the JWT token in a session cookie
func (c *CustomAuthAPIController) setSessionToken(w http.ResponseWriter, req *http.Request, token string) error {
	// Use configured cookie name, or default if not set
	cookieName := c.cfg.CookieName
	if cookieName == "" {
		cookieName = "diarycookie"
	}

	session, _ := c.cookies.Get(req, cookieName)
	session.Values["token"] = token
	// Ensure option fields are set from config
	c.configureSessionOptions(session)

	if err := session.Save(req, w); err != nil {
		return err
	}
	return nil
}

func (c *CustomAuthAPIController) configureSessionOptions(session *sessions.Session) {
	// Use configured Secure flag (defaults to true for production security)
	// Set to false only for local development without HTTPS
	session.Options.Secure = c.cfg.CookieSecure
	session.Options.SameSite = http.SameSiteLaxMode
	session.Options.HttpOnly = true
	session.Options.Path = "/"
	session.Options.MaxAge = 24 * 60 * 60 // 24 hours
}

// Authorize - validate user/password and return token
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
	if err := goserver.AssertAuthDataConstraints(authDataParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.Authorize(r.Context(), authDataParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	// If authentication was successful (200), set the cookie
	if result.Code == 200 {
		authResponse, ok := result.Body.(goserver.Authorize200Response)
		if ok && authResponse.Token != "" {
			if err := c.setSessionToken(w, r, authResponse.Token); err != nil {
				c.logger.Warn("Failed to set session cookie", "error", err)
				// Don't fail the request if cookie setting fails, just log it
			} else {
				c.logger.Info("Session cookie set successfully for API login")
			}
		}
	}

	// If no error, encode the body and the result code
	_ = goserver.EncodeJSONResponse(result.Body, &result.Code, w)
}

// clearSessionToken clears the session cookie
func (c *CustomAuthAPIController) clearSessionToken(w http.ResponseWriter, req *http.Request) error {
	cookieName := c.cfg.CookieName
	if cookieName == "" {
		cookieName = "diarycookie"
	}

	session, _ := c.cookies.Get(req, cookieName)

	c.configureSessionOptions(session)
	session.Options.MaxAge = -1

	if err := session.Save(req, w); err != nil {
		return err
	}
	return nil
}

// Logout - clear session cookie
func (c *CustomAuthAPIController) Logout(w http.ResponseWriter, r *http.Request) {
	if err := c.clearSessionToken(w, r); err != nil {
		c.logger.Error("Failed to clear session cookie", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.logger.Info("User logged out, cookie cleared")
	w.WriteHeader(http.StatusOK)
}
