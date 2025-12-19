package server

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/server/common"
	"golang.org/x/time/rate"
)

func AuthMiddleware(logger *slog.Logger, cfg *config.Config) mux.MiddlewareFunc {
	// Initialize cookie store
	cookieStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))

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

			checkToken(logger, cfg.Issuer, cfg.JWTSecret, cfg.CookieName, cookieStore, next, writer, req)
		})
	}
}

func checkToken(
	logger *slog.Logger, issuer, jwtSecret, cookieName string, cookieStore *sessions.CookieStore,
	next http.Handler, writer http.ResponseWriter, req *http.Request,
) {
	var token string

	// 1. Try Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" {
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) == 2 && authHeaderParts[0] == "Bearer" {
			token = authHeaderParts[1]
		}
	}

	// 2. Try Session Cookie if header is missing
	if token == "" {
		if cookieName == "" {
			cookieName = "diarycookie"
		}
		if session, err := cookieStore.Get(req, cookieName); err == nil {
			if t, ok := session.Values["token"].(string); ok {
				token = t
			}
		}
	}

	if token == "" {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the token
	userID, err := auth.CheckJWT(token, issuer, jwtSecret)
	if err != nil {
		logger.With("err", err).Warn("Invalid token")
		http.Error(writer, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Log successful authentication with user ID
	logger.Info("Request authenticated", "userID", userID, "path", req.URL.Path, "method", req.Method)

	req = req.WithContext(context.WithValue(req.Context(), common.UserIDKey, userID))
	next.ServeHTTP(writer, req)
}

// RateLimiterStore manages per-IP rate limiters for authentication endpoints
type RateLimiterStore struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiterStore creates a new rate limiter store
func NewRateLimiterStore() *RateLimiterStore {
	return &RateLimiterStore{
		limiters: make(map[string]*rate.Limiter),
	}
}

// GetLimiter returns a rate limiter for the given IP address
// Creates a new limiter if one doesn't exist (5 requests per minute per IP)
func (s *RateLimiterStore) GetLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	limiter, exists := s.limiters[ip]
	if !exists {
		// 5 requests per minute per IP (rate.Limit = 5/60 = ~0.083 per second)
		limiter = rate.NewLimiter(rate.Limit(5.0/60.0), 1)
		s.limiters[ip] = limiter
	}
	return limiter
}

// RateLimitMiddleware creates a middleware that applies rate limiting to the /v1/authorize endpoint
func RateLimitMiddleware(logger *slog.Logger, store *RateLimiterStore, disableRateLimit bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			// Only apply rate limiting to the authorize endpoint if not disabled
			if !disableRateLimit && req.URL.Path == "/v1/authorize" && req.Method == "POST" {
				// Extract client IP address
				clientIP := getClientIP(req)

				// Get rate limiter for this IP
				limiter := store.GetLimiter(clientIP)

				// Check if request is allowed
				if !limiter.Allow() {
					logger.Warn("Rate limit exceeded for auth endpoint", "ip", clientIP, "path", req.URL.Path)
					http.Error(writer, "Too Many Requests", http.StatusTooManyRequests)
					return
				}
			}

			next.ServeHTTP(writer, req)
		})
	}
}

// getClientIP extracts the client IP address from the request
// Checks X-Forwarded-For header first (for proxied requests), then falls back to RemoteAddr
func getClientIP(req *http.Request) string {
	// Check X-Forwarded-For header (set by proxies like Nginx)
	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (alternative proxy header)
	if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return ip
}
