package server

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

func AuthMiddleware(logger *slog.Logger, cfg *config.Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			log.Printf(
				"%s %s",
				req.Method,
				req.RequestURI)

			// Skip authorization for the root endpoint
			if req.URL.Path == "/" || strings.HasPrefix(req.URL.Path, "/web/") {
				next.ServeHTTP(writer, req)
				return
			}

			// Skip authorization for the authorize endpoint - there is no way to do it with
			// go-server openapi templates now :(
			if req.URL.Path == "/v1/authorize" {
				next.ServeHTTP(writer, req)
				return
			}

			checkToken(logger, cfg.Issuer, cfg.JWTSecret, next, writer, req)
		})
	}
}

func checkToken(
	logger *slog.Logger, issuer, jwtSecret string, next http.Handler,
	writer http.ResponseWriter, req *http.Request,
) {
	// Authorization logic - only check Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
		http.Error(writer, "Invalid authorization header", http.StatusUnauthorized)
		return
	}
	bearerToken := authHeaderParts[1]

	// Parse the token
	userID, err := auth.CheckJWT(bearerToken, issuer, jwtSecret)
	if err != nil {
		logger.With("err", err).Warn("Invalid token")
		http.Error(writer, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Log successful authentication with user ID
	logger.Info("Request authenticated", "userID", userID, "source", "header", "path", req.URL.Path, "method", req.Method)

	req = req.WithContext(context.WithValue(req.Context(), common.UserIDKey, userID))
	next.ServeHTTP(writer, req)
}
