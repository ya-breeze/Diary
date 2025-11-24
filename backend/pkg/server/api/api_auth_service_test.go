package api_test

import (
	"context"
	"encoding/base64"
	"log/slog"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/api"
)

var _ = Describe("AuthAPIService", func() {
	var (
		service    goserver.AuthAPIService
		logger     *slog.Logger
		cfg        *config.Config
		storage    database.Storage
		ctx        context.Context
		testEmail  string
		testPass   string
		hashedPass string
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		cfg = &config.Config{
			DBPath:    ":memory:",
			Issuer:    "test-issuer",
			JWTSecret: "test-secret-key-for-jwt-tokens",
		}
		storage = database.NewStorage(logger, cfg)
		Expect(storage.Open()).To(Succeed())

		service = api.NewAuthAPIService(logger, storage, cfg)
		ctx = context.Background()
		testEmail = "test@test.com"
		testPass = "testpassword123"

		// Create a test user
		hashedPassBytes, err := auth.HashPassword([]byte(testPass))
		Expect(err).ToNot(HaveOccurred())
		hashedPass = base64.StdEncoding.EncodeToString(hashedPassBytes)

		_, err = storage.CreateUser(testEmail, hashedPass)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		storage.Close()
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

				// Check that response body contains a token
				responseBody, ok := response.Body.(goserver.Authorize200Response)
				Expect(ok).To(BeTrue())
				Expect(responseBody.Token).ToNot(BeEmpty())

				// Verify the token is valid
				userID, err := auth.CheckJWT(responseBody.Token, cfg.Issuer, cfg.JWTSecret)
				Expect(err).ToNot(HaveOccurred())
				Expect(userID).ToNot(BeEmpty())
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
