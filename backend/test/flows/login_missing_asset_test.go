package flows_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/auth"
	"github.com/ya-breeze/diary.be/pkg/generated/goclient"
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
				// Step 1: Login via API
				authData := goclient.AuthData{
					Email:    setup.TestEmail,
					Password: setup.TestPass,
				}

				authResponse, httpResponse, err := setup.APIClient.AuthAPI.Authorize(context.Background()).AuthData(authData).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				Expect(authResponse.Token).ToNot(BeEmpty())

				// Step 2: Configure client with JWT token for subsequent requests
				clientConfig := setup.APIClient.GetConfig()
				clientConfig.AddDefaultHeader("Authorization", "Bearer "+authResponse.Token)

				// Step 3: Try to get a missing asset
				missingAssetPath := "nonexistent/missing-image.jpg"

				_, httpResponse, err = setup.APIClient.AssetsAPI.GetAsset(context.Background()).Path(missingAssetPath).Execute()

				// We expect this to fail with 404
				Expect(err).To(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusNotFound))

				// Verify the error is a GenericOpenAPIError with 404 status
				var openAPIErr *goclient.GenericOpenAPIError
				if errors.As(err, &openAPIErr) {
					Expect(openAPIErr.Error()).To(ContainSubstring("404"))
				}
			})
		})

		Context("when user tries to access asset without authentication", func() {
			It("should receive 401 unauthorized", func() {
				// Try to get an asset without authentication
				missingAssetPath := "some-asset.jpg"

				_, httpResponse, err := setup.APIClient.AssetsAPI.GetAsset(context.Background()).Path(missingAssetPath).Execute()

				// We expect this to fail with 401
				Expect(err).To(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when user provides invalid credentials", func() {
			It("should receive 401 authentication failed", func() {
				// Try to login with invalid credentials
				authData := goclient.AuthData{
					Email:    setup.TestEmail,
					Password: "setup.Trongpassword",
				}

				_, httpResponse, err := setup.APIClient.AuthAPI.Authorize(context.Background()).AuthData(authData).Execute()

				// We expect this to fail with 401
				Expect(err).To(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when user logs in and fetches an existing asset", func() {
			It("should successfully retrieve the asset", func() {
				// Step 1: Login via API
				authData := goclient.AuthData{
					Email:    setup.TestEmail,
					Password: setup.TestPass,
				}

				authResponse, httpResponse, err := setup.APIClient.AuthAPI.Authorize(context.Background()).AuthData(authData).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				Expect(authResponse.Token).ToNot(BeEmpty())

				// Step 2: Configure client with JWT token for subsequent requests
				clientConfig := setup.APIClient.GetConfig()
				clientConfig.AddDefaultHeader("Authorization", "Bearer "+authResponse.Token)

				// Step 3: Create a test asset file in the user's directory
				// First, we need to get the user ID from the JWT token to create the correct directory structure
				userID, err := auth.CheckJWT(authResponse.Token, setup.Cfg.Issuer, setup.Cfg.JWTSecret)
				Expect(err).ToNot(HaveOccurred())

				userAssetDir := filepath.Join(setup.TempDir, userID)
				err = os.MkdirAll(userAssetDir, 0o755)
				Expect(err).ToNot(HaveOccurred())

				testAssetPath := "images/photos/test-photo.jpg"
				testAssetFullPath := filepath.Join(userAssetDir, testAssetPath)
				testAssetDir := filepath.Dir(testAssetFullPath)
				err = os.MkdirAll(testAssetDir, 0o755)
				Expect(err).ToNot(HaveOccurred())

				testAssetContent := []byte("fake image content for testing")
				err = os.WriteFile(testAssetFullPath, testAssetContent, 0o600)
				Expect(err).ToNot(HaveOccurred())

				// Step 4: Fetch the existing asset
				assetFile, httpResponse, err := setup.APIClient.AssetsAPI.GetAsset(context.Background()).Path(testAssetPath).Execute()

				// We expect this to succeed with 200
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				Expect(assetFile).ToNot(BeNil())

				// Step 5: Verify the content (optional - read and compare)
				defer assetFile.Close()
				retrievedContent, err := io.ReadAll(assetFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(retrievedContent).To(Equal(testAssetContent))
			})
		})
	})
})
