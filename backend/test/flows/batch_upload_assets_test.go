package flows_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"

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

			// Create temp files
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

			resp, httpResp, err := setup.APIClient.AssetsAPI.UploadAssetsBatch(context.Background()).Assets([]*os.File{f1, f2, f3}).Execute()
			Expect(err).ToNot(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp).ToNot(BeNil())
			Expect(int(resp.GetCount())).To(Equal(3))
			Expect(resp.Files).To(HaveLen(3))

			// Verify retrieval for each saved file name
			for _, file := range resp.Files {
				name := strings.TrimSpace(file.GetSavedName())
				rd, httpResp, err := setup.APIClient.AssetsAPI.GetAsset(context.Background()).Path(name).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				b, err := io.ReadAll(rd)
				Expect(err).ToNot(HaveOccurred())
				Expect(b).ToNot(BeEmpty())
				_ = rd.Close()
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

			_, httpResp, err := setup.APIClient.AssetsAPI.UploadAssetsBatch(context.Background()).Assets([]*os.File{good, bad}).Execute()
			Expect(err).To(HaveOccurred())
			Expect(httpResp.StatusCode).To(BeNumerically(">=", 400))
		})
	})
})
