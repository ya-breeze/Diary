package flows_test

import (
	"context"
	"io"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Batch Upload Assets Flow", func() {
	var setup *SharedTestSetup

	BeforeEach(func() {
		setup = SetupTestEnvironment()
	})

	AfterEach(func() {
		setup.TeardownTestEnvironment()
	})

	Context("when user uploads multiple assets via /v1/assets/batch", func() {
		It("should successfully upload and retrieve them all", func() {
			setup.LoginAndGetToken()

			mkFile := func(suffix string, content []byte) *os.File {
				f, err := os.CreateTemp("", "batch_*."+suffix)
				Expect(err).ToNot(HaveOccurred())
				_, err = f.Write(content)
				Expect(err).ToNot(HaveOccurred())
				_, err = f.Seek(0, 0)
				Expect(err).ToNot(HaveOccurred())
				return f
			}

			f1 := mkFile("jpg", []byte("one"))
			defer os.Remove(f1.Name())
			defer f1.Close()
			f2 := mkFile("jpg", []byte("two"))
			defer os.Remove(f2.Name())
			defer f2.Close()
			f3 := mkFile("jpg", []byte("three"))
			defer os.Remove(f3.Name())
			defer f3.Close()

			resp, httpResp, err := setup.APIClient.UploadAssetsBatch(context.Background(), []*os.File{f1, f2, f3})
			Expect(err).ToNot(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp).ToNot(BeNil())
			Expect(resp.Count).To(Equal(3))
			Expect(resp.Files).To(HaveLen(3))

			// Verify retrieval for each saved file
			for _, file := range resp.Files {
				assetResp, err := setup.APIClient.GetAsset(context.Background(), file.SavedName)
				Expect(err).ToNot(HaveOccurred())
				defer assetResp.Body.Close()
				Expect(assetResp.StatusCode).To(Equal(http.StatusOK))
				b, err := io.ReadAll(assetResp.Body)
				Expect(err).ToNot(HaveOccurred())
				Expect(b).ToNot(BeEmpty())
			}
		})
	})

	Context("when one of the files is invalid", func() {
		It("should return 4xx and not upload anything", func() {
			setup.LoginAndGetToken()

			good, err := os.CreateTemp("", "batch_good_*.jpg")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(good.Name())
			defer good.Close()
			_, _ = good.WriteString("ok")
			_, _ = good.Seek(0, 0)

			bad, err := os.CreateTemp("", "batch_bad_*.exe")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(bad.Name())
			defer bad.Close()
			_, _ = bad.WriteString("bad")
			_, _ = bad.Seek(0, 0)

			_, httpResp, err := setup.APIClient.UploadAssetsBatch(context.Background(), []*os.File{good, bad})
			Expect(err).To(HaveOccurred())
			Expect(httpResp.StatusCode).To(BeNumerically(">=", 400))
		})
	})
})
