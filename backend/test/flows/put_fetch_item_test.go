package flows_test

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Put and Fetch Item Flow", func() {
	var setup *SharedTestSetup

	BeforeEach(func() {
		setup = SetupTestEnvironment()
	})

	AfterEach(func() {
		setup.TeardownTestEnvironment()
	})

	Describe("Item put and retrieval", func() {
		Context("when user puts an item and then fetches it", func() {
			It("should successfully save and then retrieve the same item", func() {
				setup.LoginAndGetToken()

				date := time.Now().Format("2006-01-02")

				putResp, httpResp, err := setup.APIClient.PutItems(
					context.Background(), date, "Test Entry Title",
					"This is a test body for the diary entry.", []string{"test", "ginkgo"},
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(putResp.Date).To(Equal(date))
				Expect(putResp.Title).To(Equal("Test Entry Title"))
				Expect(putResp.Body).To(Equal("This is a test body for the diary entry."))
				Expect(putResp.Tags).To(Equal([]string{"test", "ginkgo"}))

				fetched, httpResp, err := setup.APIClient.GetItems(context.Background(), date, "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(fetched).ToNot(BeNil())
				Expect(fetched.Items).To(HaveLen(1))
				Expect(fetched.TotalCount).To(Equal(1))

				item := fetched.Items[0]
				Expect(item.Date).To(Equal(date))
				Expect(item.Title).To(Equal("Test Entry Title"))
				Expect(item.Body).To(Equal("This is a test body for the diary entry."))
				Expect(item.Tags).To(Equal([]string{"test", "ginkgo"}))
			})
		})
	})
})
