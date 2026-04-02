package flows_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Home Navigation Fix", func() {
	var setup *SharedTestSetup

	BeforeEach(func() {
		setup = SetupTestEnvironment()
	})

	AfterEach(func() {
		setup.TeardownTestEnvironment()
	})

	Describe("Navigation dates for non-existing entries", func() {
		Context("when accessing a date that doesn't exist but has entries before and after", func() {
			It("should provide navigation dates for empty entries", func() {
				setup.LoginAndGetToken()
				ctx := context.Background()

				_, httpResp, err := setup.APIClient.PutItems(ctx, "2024-01-10", "First Entry", "Content of first entry", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				_, httpResp, err = setup.APIClient.PutItems(ctx, "2024-01-12", "Third Entry", "Content of third entry", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				_, httpResp, err = setup.APIClient.PutItems(ctx, "2024-01-13", "Fourth Entry", "Content of fourth entry", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				// Fetch the gap date (2024-01-11 doesn't exist but has entries before and after)
				fetched, httpResp, err := setup.APIClient.GetItems(ctx, "2024-01-11", "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(fetched.Items).To(HaveLen(1))

				item := fetched.Items[0]
				Expect(item.Date).To(Equal("2024-01-11"))
				Expect(item.PreviousDate).ToNot(BeNil())
				Expect(*item.PreviousDate).To(Equal("2024-01-10"))
				Expect(item.NextDate).ToNot(BeNil())
				Expect(*item.NextDate).To(Equal("2024-01-12"))
			})
		})

		Context("when accessing a date with existing entries before and after", func() {
			It("should return navigation dates for existing entries", func() {
				setup.LoginAndGetToken()
				ctx := context.Background()

				_, httpResp, err := setup.APIClient.PutItems(ctx, "2024-01-10", "First Entry", "Content 1", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				_, httpResp, err = setup.APIClient.PutItems(ctx, "2024-01-11", "Second Entry", "Content 2", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				_, httpResp, err = setup.APIClient.PutItems(ctx, "2024-01-12", "Third Entry", "Content 3", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				fetched, httpResp, err := setup.APIClient.GetItems(ctx, "2024-01-11", "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(fetched.Items).To(HaveLen(1))

				item := fetched.Items[0]
				Expect(item.Date).To(Equal("2024-01-11"))
				Expect(item.PreviousDate).ToNot(BeNil())
				Expect(*item.PreviousDate).To(Equal("2024-01-10"))
				Expect(item.NextDate).ToNot(BeNil())
				Expect(*item.NextDate).To(Equal("2024-01-12"))
			})
		})

		Context("when items are saved with the gap date", func() {
			It("should correctly fill the gap and preserve navigation", func() {
				setup.LoginAndGetToken()
				ctx := context.Background()

				_, httpResp, err := setup.APIClient.PutItems(ctx, "2024-01-10", "First Entry", "Content 1", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				_, httpResp, err = setup.APIClient.PutItems(ctx, "2024-01-12", "Third Entry", "Content 3", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				// Save the gap date
				_, httpResp, err = setup.APIClient.PutItems(ctx, "2024-01-11", "Middle Entry", "Middle content", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				fetched, httpResp, err := setup.APIClient.GetItems(ctx, "2024-01-11", "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(fetched.Items).To(HaveLen(1))

				item := fetched.Items[0]
				Expect(item.Title).To(Equal("Middle Entry"))
				Expect(item.PreviousDate).ToNot(BeNil())
				Expect(*item.PreviousDate).To(Equal("2024-01-10"))
				Expect(item.NextDate).ToNot(BeNil())
				Expect(*item.NextDate).To(Equal("2024-01-12"))
			})
		})
	})
})
