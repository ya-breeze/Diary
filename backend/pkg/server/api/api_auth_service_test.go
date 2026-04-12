package api_test

import (
	"context"
	"log/slog"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kinauth "github.com/ya-breeze/kin-core/auth"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/api"
)

var _ = Describe("AuthAPIService", func() {
	var (
		service   goserver.AuthAPIService
		logger    *slog.Logger
		cfg       *config.Config
		storage   database.Storage
		ctx       context.Context
		testEmail string
		testPass  string
		tempDir   string
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		var err error
		tempDir, err = os.MkdirTemp("", "auth_test")
		Expect(err).NotTo(HaveOccurred())

		cfg = &config.Config{
			DataPath:     tempDir,
			JWTSecret:    "test-secret-key-for-jwt-tokens",
			CookieSecure: false,
		}
		storage = database.NewStorage(logger, cfg)
		Expect(storage.Open()).To(Succeed())

		service = api.NewAuthAPIService(logger, storage, cfg)
		ctx = context.Background()
		testEmail = "test@test.com"
		testPass = "testpassword123"

		// Create family and user with kin-core compatible bcrypt hash
		family, err := storage.CreateFamily("TestFamily")
		Expect(err).ToNot(HaveOccurred())

		hashedPass, err := kinauth.HashPassword(testPass)
		Expect(err).ToNot(HaveOccurred())

		_, err = storage.CreateUser(testEmail, hashedPass, family.ID)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	Describe("Authorize", func() {
		Context("with valid credentials", func() {
			It("should return a JWT token", func() {
				authData := goserver.AuthData{
					Email:    testEmail,
					Password: testPass,
				}

				response, err := service.Authorize(ctx, authData)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				responseBody, ok := response.Body.(goserver.Authorize200Response)
				Expect(ok).To(BeTrue())
				Expect(responseBody.Token).ToNot(BeEmpty())

				// Verify the token is a valid kin-core JWT
				claims, parseErr := kinauth.ParseToken(responseBody.Token, []byte(cfg.JWTSecret))
				Expect(parseErr).ToNot(HaveOccurred())
				Expect(claims.UserID).ToNot(Equal(""))
			})
		})

		Context("with invalid email", func() {
			It("should return 401 unauthorized", func() {
				authData := goserver.AuthData{
					Email:    "nonexistent@example.com",
					Password: testPass,
				}

				response, err := service.Authorize(ctx, authData)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(401))
				Expect(response.Body).To(BeNil())
			})
		})

		Context("with invalid password", func() {
			It("should return 401 unauthorized", func() {
				authData := goserver.AuthData{
					Email:    testEmail,
					Password: "wrongpassword",
				}

				response, err := service.Authorize(ctx, authData)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(401))
				Expect(response.Body).To(BeNil())
			})
		})

		Context("with empty credentials", func() {
			It("should return 401 unauthorized for empty email", func() {
				authData := goserver.AuthData{
					Email:    "",
					Password: testPass,
				}

				response, err := service.Authorize(ctx, authData)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(401))
				Expect(response.Body).To(BeNil())
			})

			It("should return 401 unauthorized for empty password", func() {
				authData := goserver.AuthData{
					Email:    testEmail,
					Password: "",
				}

				response, err := service.Authorize(ctx, authData)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(401))
				Expect(response.Body).To(BeNil())
			})
		})
	})
})
