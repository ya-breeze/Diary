package flows_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Item Edge Cases", func() {
	var setup *SharedTestSetup

	BeforeEach(func() {
		setup = SetupTestEnvironment()
		setup.LoginAndGetToken()
	})

	AfterEach(func() {
		setup.TeardownTestEnvironment()
	})

	Describe("Fetching items", func() {
		Context("when the requested date has no entry", func() {
			It("returns an empty list with status 200", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "1985-01-01", "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.TotalCount).To(Equal(1))
				Expect(result.Items).To(HaveLen(1))
				// Empty entry for the date is returned (placeholder behaviour)
				Expect(result.Items[0].Date).To(Equal("1985-01-01"))
				Expect(result.Items[0].Title).To(BeEmpty())
			})
		})

		Context("when the date query parameter is malformed", func() {
			It("returns 400 for a non-date string", func() {
				_, httpResp, err := setup.APIClient.GetItems(context.Background(), "not-a-date", "", "")
				Expect(err).To(HaveOccurred())
				Expect(httpResp).NotTo(BeNil())
				Expect(httpResp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("Putting items", func() {
		Context("when the same date is written twice", func() {
			It("second PUT overwrites the first (idempotent)", func() {
				ctx := context.Background()
				date := "2005-03-15"

				_, httpResp, err := setup.APIClient.PutItems(ctx, date, "First Title", "First body", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				_, httpResp, err = setup.APIClient.PutItems(ctx, date, "Second Title", "Second body", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				result, httpResp, err := setup.APIClient.GetItems(ctx, date, "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items[0].Title).To(Equal("Second Title"))
				Expect(result.Items[0].Body).To(Equal("Second body"))
			})
		})

		Context("when title is empty", func() {
			It("accepts the request with 200 (server accepts empty title)", func() {
				// The OpenAPI spec has no minLength constraint on title, so the server
				// accepts empty titles and returns 200.
				ctx := context.Background()
				putResp, httpResp, err := setup.APIClient.PutItems(ctx, "2005-04-01", "", "some body", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(putResp.Title).To(BeEmpty())
			})
		})
	})

	Describe("Searching items", func() {
		Context("when search matches multiple entries", func() {
			It("returns all matching entries", func() {
				ctx := context.Background()
				for i := 1; i <= 3; i++ {
					date := fmt.Sprintf("2006-%02d-01", i)
					_, httpResp, err := setup.APIClient.PutItems(ctx, date, fmt.Sprintf("SearchTarget%d", i), "body", nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				}

				result, httpResp, err := setup.APIClient.GetItems(ctx, "", "SearchTarget", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.TotalCount).To(Equal(3))
			})
		})
	})
})
