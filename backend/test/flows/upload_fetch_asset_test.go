package flows_test

import (
	"context"
	"io"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upload and Fetch Asset Flow", func() {
	var setup *SharedTestSetup

	BeforeEach(func() {
		setup = SetupTestEnvironment()
	})

	AfterEach(func() {
		setup.TeardownTestEnvironment()
	})

	Describe("Asset Upload and Retrieval Flow", func() {
		Context("when user uploads an asset and then fetches it", func() {
			It("should successfully upload and then retrieve the same asset", func() {
				// Step 1: Login via API
				setup.LoginAndGetToken()

				// Step 2: Create a test asset file to upload
				testAssetContent := []byte("test image content for upload")

				// Create a temporary file for upload
				tempFile, err := os.CreateTemp("", "test_upload_*.jpg")
				Expect(err).ToNot(HaveOccurred())
				defer os.Remove(tempFile.Name())
				defer tempFile.Close()

				_, err = tempFile.Write(testAssetContent)
				Expect(err).ToNot(HaveOccurred())

				// Reset file pointer to beginning for reading
				_, err = tempFile.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())

				// Step 3: Upload the asset via batch API
				uploadResponse, httpResponse, err := setup.APIClient.AssetsAPI.UploadAssetsBatch(context.Background()).Assets([]*os.File{tempFile}).Execute()

				// We expect this to succeed with 200
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				Expect(uploadResponse).ToNot(BeNil())
				Expect(uploadResponse.Count).To(Equal(int32(1)))

				// The response should contain the filename of the uploaded asset
				uploadedFilename := uploadResponse.Files[0].GetSavedName()
				Expect(uploadedFilename).To(HaveSuffix(".jpg"))

				// Step 4: Fetch the uploaded asset using the returned filename
				assetFile, httpResponse, err := setup.APIClient.AssetsAPI.GetAsset(context.Background()).Path(uploadedFilename).Execute()

				// We expect this to succeed with 200
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				Expect(assetFile).ToNot(BeNil())

				// Step 5: Verify the content matches what we uploaded
				defer assetFile.Close()
				retrievedContent, err := io.ReadAll(assetFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(retrievedContent).To(Equal(testAssetContent))
			})
		})

		Context("when user tries to upload without authentication", func() {
			It("should receive 401 unauthorized", func() {
				// Create a test asset file to upload
				testAssetContent := []byte("test image content for upload")

				// Create a temporary file for upload
				tempFile, err := os.CreateTemp("", "test_upload_*.jpg")
				Expect(err).ToNot(HaveOccurred())
				defer os.Remove(tempFile.Name())
				defer tempFile.Close()

				_, err = tempFile.Write(testAssetContent)
				Expect(err).ToNot(HaveOccurred())

				// Reset file pointer to beginning for reading
				_, err = tempFile.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())

				// Try to upload without authentication
				_, httpResponse, err := setup.APIClient.AssetsAPI.UploadAssetsBatch(context.Background()).Assets([]*os.File{tempFile}).Execute()

				// We expect this to fail with 401
				Expect(err).To(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when user uploads multiple assets", func() {
			It("should successfully upload and retrieve multiple different assets", func() {
				// Step 1: Login via API
				setup.LoginAndGetToken()

				// Step 2: Create two test asset files
				firstAssetContent := []byte("first test image content")
				firstTempFile, err := os.CreateTemp("", "test_upload_1_*.jpg")
				Expect(err).ToNot(HaveOccurred())
				defer os.Remove(firstTempFile.Name())
				defer firstTempFile.Close()

				_, err = firstTempFile.Write(firstAssetContent)
				Expect(err).ToNot(HaveOccurred())
				_, err = firstTempFile.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())

				secondAssetContent := []byte("second test image content with different data")
				secondTempFile, err := os.CreateTemp("", "test_upload_2_*.jpg")
				Expect(err).ToNot(HaveOccurred())
				defer os.Remove(secondTempFile.Name())
				defer secondTempFile.Close()

				_, err = secondTempFile.Write(secondAssetContent)
				Expect(err).ToNot(HaveOccurred())
				_, err = secondTempFile.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())

				// Step 3: Upload both assets in one batch call
				uploadResponse, httpResponse, err := setup.APIClient.AssetsAPI.UploadAssetsBatch(context.Background()).Assets([]*os.File{firstTempFile, secondTempFile}).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				Expect(uploadResponse.Count).To(Equal(int32(2)))

				firstFilename := uploadResponse.Files[0].GetSavedName()
				secondFilename := uploadResponse.Files[1].GetSavedName()

				// Step 4: Verify filenames are different
				Expect(firstFilename).ToNot(Equal(secondFilename))

				// Step 5: Fetch and verify first asset
				firstAssetFile, httpResponse, err := setup.APIClient.AssetsAPI.GetAsset(context.Background()).Path(firstFilename).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				defer firstAssetFile.Close()

				firstRetrievedContent, err := io.ReadAll(firstAssetFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(firstRetrievedContent).To(Equal(firstAssetContent))

				// Step 6: Fetch and verify second asset
				secondAssetFile, httpResponse, err := setup.APIClient.AssetsAPI.GetAsset(context.Background()).Path(secondFilename).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				defer secondAssetFile.Close()

				secondRetrievedContent, err := io.ReadAll(secondAssetFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(secondRetrievedContent).To(Equal(secondAssetContent))
			})
		})
	})
})
