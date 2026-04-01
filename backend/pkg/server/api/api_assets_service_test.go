package api_test

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/server/api"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

// Helper function to create context with user ID for assets tests
func createContextWithUserIDForAssets(userID string) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, common.UserIDKey, userID)
}

var _ = Describe("AssetsAPIService", func() {
	var (
		service  *api.AssetsAPIServiceImpl
		logger   *slog.Logger
		cfg      *config.Config
		tempDir  string
		userID   string
		testFile string
		ctx      context.Context
	)

	// Create context outside of BeforeEach to avoid fatcontext linting issue
	userID = "test-user-123"
	ctx = createContextWithUserIDForAssets(userID)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

		var err error
		tempDir, err = os.MkdirTemp("", "assets_test")
		Expect(err).NotTo(HaveOccurred())

		cfg = &config.Config{
			DataPath: tempDir,
		}

		// Create user directory and test file
		userDir := filepath.Join(tempDir, config.AssetsDirName, userID)
		err = os.MkdirAll(userDir, 0o755)
		Expect(err).NotTo(HaveOccurred())

		testFile = filepath.Join(userDir, "test-image.jpg")
		err = os.WriteFile(testFile, []byte("fake image content"), 0o600)
		Expect(err).NotTo(HaveOccurred())

		serviceInterface := api.NewAssetsAPIService(logger, cfg)
		var ok bool
		service, ok = serviceInterface.(*api.AssetsAPIServiceImpl)
		Expect(ok).To(BeTrue(), "Failed to cast service to AssetsAPIServiceImpl")
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("GetAsset", func() {
		Context("when user ID is missing from context", func() {
			It("should return unauthorized", func() {
				emptyCtx := context.Background()
				response, err := service.GetAsset(emptyCtx, "test-image.jpg")

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when path contains directory traversal", func() {
			It("should return bad request for .. in path", func() {
				response, err := service.GetAsset(ctx, "../secret.txt")

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for absolute path", func() {
				response, err := service.GetAsset(ctx, "/etc/passwd")

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when accessing files in subdirectories", func() {
			BeforeEach(func() {
				// Create a subdirectory with a file
				subDir := filepath.Join(tempDir, config.AssetsDirName, userID, "images")
				err := os.MkdirAll(subDir, 0o755)
				Expect(err).NotTo(HaveOccurred())

				subFile := filepath.Join(subDir, "photo.jpg")
				err = os.WriteFile(subFile, []byte("photo content"), 0o600)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow access to files in subdirectories", func() {
				response, err := service.GetAsset(ctx, "images/photo.jpg")

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusOK))
				Expect(response.Body).To(BeAssignableToTypeOf(&os.File{}))

				// Verify we can read from the file
				file, ok := response.Body.(*os.File)
				Expect(ok).To(BeTrue(), "Failed to cast response body to *os.File")
				defer file.Close()

				content, err := os.ReadFile(file.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("photo content"))
			})

			It("should handle nested subdirectories", func() {
				// Create deeper nesting
				deepDir := filepath.Join(tempDir, config.AssetsDirName, userID, "docs", "2023", "reports")
				err := os.MkdirAll(deepDir, 0o755)
				Expect(err).NotTo(HaveOccurred())

				deepFile := filepath.Join(deepDir, "report.pdf")
				err = os.WriteFile(deepFile, []byte("report content"), 0o600)
				Expect(err).NotTo(HaveOccurred())

				response, err := service.GetAsset(ctx, "docs/2023/reports/report.pdf")

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when file does not exist", func() {
			It("should return not found", func() {
				response, err := service.GetAsset(ctx, "nonexistent.jpg")

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("when file exists", func() {
			It("should return the file successfully", func() {
				response, err := service.GetAsset(ctx, "test-image.jpg")

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusOK))
				Expect(response.Body).To(BeAssignableToTypeOf(&os.File{}))

				// Verify we can read from the file
				file, ok := response.Body.(*os.File)
				Expect(ok).To(BeTrue(), "Failed to cast response body to *os.File")
				defer file.Close()

				content, err := os.ReadFile(file.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("fake image content"))
			})
		})

		Context("when path points to a directory", func() {
			It("should return bad request", func() {
				// Create a subdirectory
				subDir := filepath.Join(tempDir, config.AssetsDirName, userID, "emptydir")
				err := os.MkdirAll(subDir, 0o755)
				Expect(err).NotTo(HaveOccurred())

				response, err := service.GetAsset(ctx, "emptydir")

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

	})
})
