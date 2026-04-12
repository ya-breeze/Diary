package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kinauth "github.com/ya-breeze/kin-core/auth"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
)

func TestMiddlewares(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Middlewares")
}

var _ = Describe("AuthMiddleware", func() {
	var (
		logger    *slog.Logger
		cfg       *config.Config
		storage   database.Storage
		tempDir   string
		jwtSecret string
		userID    uuid.UUID
		familyID  uuid.UUID
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		jwtSecret = "secret"
		var err error
		tempDir, err = os.MkdirTemp("", "middleware_test")
		Expect(err).NotTo(HaveOccurred())

		cfg = &config.Config{
			JWTSecret:    jwtSecret,
			CookieSecure: false,
			DataPath:     tempDir,
		}

		storage = database.NewStorage(logger, cfg)
		Expect(storage.Open()).To(Succeed())

		userID = uuid.New()
		familyID = uuid.New()
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	Context("when authenticating requests", func() {
		It("should allow request with valid Authorization header (Bearer token)", func() {
			middleware := AuthMiddleware(logger, cfg, storage.GetDB())
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			token, err := kinauth.GenerateAccessToken(userID, &familyID, []byte(jwtSecret), 15*time.Minute)
			Expect(err).NotTo(HaveOccurred())

			req := httptest.NewRequest("GET", "/v1/items", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should allow request with valid kin_access cookie", func() {
			middleware := AuthMiddleware(logger, cfg, storage.GetDB())
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			token, err := kinauth.GenerateAccessToken(userID, &familyID, []byte(jwtSecret), 15*time.Minute)
			Expect(err).NotTo(HaveOccurred())

			req := httptest.NewRequest("GET", "/v1/media/123", nil)
			req.AddCookie(&http.Cookie{Name: "kin_access", Value: token})
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusOK))
		})

		It("should reject request with no auth", func() {
			middleware := AuthMiddleware(logger, cfg, storage.GetDB())
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/v1/items", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should reject request with invalid token", func() {
			middleware := AuthMiddleware(logger, cfg, storage.GetDB())
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/v1/items", nil)
			req.Header.Set("Authorization", "Bearer invalid-token")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("RateLimitMiddleware", func() {
	var (
		logger *slog.Logger
		store  *RateLimiterStore
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		store = NewRateLimiterStore()
	})

	Describe("Rate limiting on /v1/authorize endpoint", func() {
		Context("when rate limiting is enabled", func() {
			It("should allow first request", func() {
				middleware := RateLimitMiddleware(logger, store, false)
				handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				req := httptest.NewRequest("POST", "/v1/authorize", nil)
				req.RemoteAddr = "127.0.0.1:12345"
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)
				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should reject rapid successive requests", func() {
				middleware := RateLimitMiddleware(logger, store, false)
				handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				// First request should succeed
				req := httptest.NewRequest("POST", "/v1/authorize", nil)
				req.RemoteAddr = "127.0.0.1:12345"
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).To(Equal(http.StatusOK))

				// Second request should be rate limited (burst is 1)
				req = httptest.NewRequest("POST", "/v1/authorize", nil)
				req.RemoteAddr = "127.0.0.1:12345"
				w = httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).To(Equal(http.StatusTooManyRequests))
			})

			It("should apply rate limiting per IP address", func() {
				middleware := RateLimitMiddleware(logger, store, false)
				handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				// First request from IP1 should succeed
				req := httptest.NewRequest("POST", "/v1/authorize", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).To(Equal(http.StatusOK))

				// First request from IP2 should also succeed (different limiter)
				req = httptest.NewRequest("POST", "/v1/authorize", nil)
				req.RemoteAddr = "192.168.1.2:12345"
				w = httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).To(Equal(http.StatusOK))
			})

			It("should not apply rate limiting to non-authorize endpoints", func() {
				middleware := RateLimitMiddleware(logger, store, false)
				handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				// Make 10 requests to /v1/items (should all succeed)
				for i := 0; i < 10; i++ {
					req := httptest.NewRequest("GET", "/v1/items", nil)
					req.RemoteAddr = "127.0.0.1:12345"
					w := httptest.NewRecorder()
					handler.ServeHTTP(w, req)
					Expect(w.Code).To(Equal(http.StatusOK))
				}
			})
		})

		Context("when rate limiting is disabled", func() {
			It("should allow unlimited requests", func() {
				middleware := RateLimitMiddleware(logger, store, true)
				handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				// Make 100 requests (should all succeed)
				for i := 0; i < 100; i++ {
					req := httptest.NewRequest("POST", "/v1/authorize", nil)
					req.RemoteAddr = "127.0.0.1:12345"
					w := httptest.NewRecorder()
					handler.ServeHTTP(w, req)
					Expect(w.Code).To(Equal(http.StatusOK))
				}
			})
		})
	})

	Describe("getClientIP function", func() {
		It("should extract IP from X-Forwarded-For header", func() {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
			ip := getClientIP(req)
			Expect(ip).To(Equal("203.0.113.1"))
		})

		It("should extract IP from X-Real-IP header", func() {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Real-IP", "203.0.113.2")
			ip := getClientIP(req)
			Expect(ip).To(Equal("203.0.113.2"))
		})

		It("should fall back to RemoteAddr", func() {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			ip := getClientIP(req)
			Expect(ip).To(Equal("127.0.0.1"))
		})
	})
})
