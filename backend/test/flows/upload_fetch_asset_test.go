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
				setup.LoginAndGetToken()

				testAssetContent := []byte("test image content for upload")

				tempFile, err := os.CreateTemp("", "test_upload_*.jpg")
				Expect(err).ToNot(HaveOccurred())
				defer os.Remove(tempFile.Name())
				defer tempFile.Close()

				_, err = tempFile.Write(testAssetContent)
				Expect(err).ToNot(HaveOccurred())
				_, err = tempFile.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())

				uploadResponse, httpResponse, err := setup.APIClient.UploadAssetsBatch(context.Background(), []*os.File{tempFile})
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				Expect(uploadResponse).ToNot(BeNil())
				Expect(uploadResponse.Count).To(Equal(1))

				uploadedFilename := uploadResponse.Files[0].SavedName
				Expect(uploadedFilename).To(HaveSuffix(".jpg"))

				assetResp, err := setup.APIClient.GetAsset(context.Background(), uploadedFilename)
				Expect(err).ToNot(HaveOccurred())
				defer assetResp.Body.Close()
				Expect(assetResp.StatusCode).To(Equal(http.StatusOK))

				retrievedContent, err := io.ReadAll(assetResp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(retrievedContent).To(Equal(testAssetContent))
			})
		})

		Context("when user tries to upload without authentication", func() {
			It("should receive 401 unauthorized", func() {
				tempFile, err := os.CreateTemp("", "test_upload_*.jpg")
				Expect(err).ToNot(HaveOccurred())
				defer os.Remove(tempFile.Name())
				defer tempFile.Close()

				_, err = tempFile.Write([]byte("test image content for upload"))
				Expect(err).ToNot(HaveOccurred())
				_, err = tempFile.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())

				_, httpResponse, err := setup.APIClient.UploadAssetsBatch(context.Background(), []*os.File{tempFile})
				Expect(err).To(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when user uploads multiple assets", func() {
			It("should successfully upload and retrieve multiple different assets", func() {
				setup.LoginAndGetToken()

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

				uploadResponse, httpResponse, err := setup.APIClient.UploadAssetsBatch(context.Background(), []*os.File{firstTempFile, secondTempFile})
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
				Expect(uploadResponse.Count).To(Equal(2))

				firstFilename := uploadResponse.Files[0].SavedName
				secondFilename := uploadResponse.Files[1].SavedName
				Expect(firstFilename).ToNot(Equal(secondFilename))

				firstAssetResp, err := setup.APIClient.GetAsset(context.Background(), firstFilename)
				Expect(err).ToNot(HaveOccurred())
				defer firstAssetResp.Body.Close()
				Expect(firstAssetResp.StatusCode).To(Equal(http.StatusOK))
				firstRetrievedContent, err := io.ReadAll(firstAssetResp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(firstRetrievedContent).To(Equal(firstAssetContent))

				secondAssetResp, err := setup.APIClient.GetAsset(context.Background(), secondFilename)
				Expect(err).ToNot(HaveOccurred())
				defer secondAssetResp.Body.Close()
				Expect(secondAssetResp.StatusCode).To(Equal(http.StatusOK))
				secondRetrievedContent, err := io.ReadAll(secondAssetResp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(secondRetrievedContent).To(Equal(secondAssetContent))
			})
		})
	})
})
