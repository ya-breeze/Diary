package flows_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/config"
)

var _ = Describe("Login and Missing Asset Flow", func() {
	var setup *SharedTestSetup

	BeforeEach(func() {
		setup = SetupTestEnvironment()
	})

	AfterEach(func() {
		setup.TeardownTestEnvironment()
	})

	Describe("Authentication and Asset Access Flow", func() {
		Context("when user logs in successfully", func() {
			It("should authenticate and then receive 404 for missing asset", func() {
				authResp, httpResp, err := setup.APIClient.Authorize(
					context.Background(), setup.TestEmail, setup.TestPass,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(authResp.Token).ToNot(BeEmpty())

				setup.APIClient.SetToken(authResp.Token)

				resp, err := setup.APIClient.GetAsset(context.Background(), "nonexistent/missing-image.jpg")
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Context("when user tries to access asset without authentication", func() {
			It("should receive 401 unauthorized", func() {
				resp, err := setup.APIClient.GetAsset(context.Background(), "some-asset.jpg")
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when user provides invalid credentials", func() {
			It("should receive 401 authentication failed", func() {
				_, httpResp, err := setup.APIClient.Authorize(
					context.Background(), setup.TestEmail, "wrongpassword",
				)
				Expect(err).To(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when user logs in and fetches an existing asset", func() {
			It("should successfully retrieve the asset", func() {
				authResp, httpResp, err := setup.APIClient.Authorize(
					context.Background(), setup.TestEmail, setup.TestPass,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				setup.APIClient.SetToken(authResp.Token)

				// Create a test asset in the user's directory
				userID, err := auth.CheckJWT(authResp.Token, setup.Cfg.Issuer, setup.Cfg.JWTSecret)
				Expect(err).ToNot(HaveOccurred())

				userAssetDir := filepath.Join(setup.TempDir, config.AssetsDirName, userID)
				Expect(os.MkdirAll(userAssetDir, 0o755)).To(Succeed())

				testAssetPath := "images/photos/test-photo.jpg"
				testAssetFullPath := filepath.Join(userAssetDir, testAssetPath)
				Expect(os.MkdirAll(filepath.Dir(testAssetFullPath), 0o755)).To(Succeed())

				testContent := []byte("fake image content for testing")
				Expect(os.WriteFile(testAssetFullPath, testContent, 0o600)).To(Succeed())

				resp, err := setup.APIClient.GetAsset(context.Background(), testAssetPath)
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				retrievedContent, err := io.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(retrievedContent).To(Equal(testContent))
			})
		})
	})
})
