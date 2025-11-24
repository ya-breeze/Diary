package flows_test

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/generated/goclient"
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
				// Login and get token (also configures API client headers)
				setup.LoginAndGetToken()

				// Prepare item data
				date := time.Now().Format("2006-01-02")

				// Use generated client to PUT the item
				itemsReq := *goclient.NewItemsRequest(date, "Test Entry Title", "This is a test body for the diary entry.")
				itemsReq.SetTags([]string{"test", "ginkgo"})

				putResp, httpResp, err := setup.APIClient.ItemsAPI.PutItems(context.Background()).ItemsRequest(itemsReq).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				// Verify PUT response
				Expect(putResp.Date).To(Equal(date))
				Expect(putResp.Title).To(Equal("Test Entry Title"))
				Expect(putResp.Body).To(Equal("This is a test body for the diary entry."))
				Expect(putResp.Tags).To(Equal([]string{"test", "ginkgo"}))

				// Fetch the item via generated goclient GetItems (now returns list)
				fetched, httpResp, err := setup.APIClient.ItemsAPI.GetItems(context.Background()).Date(date).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(fetched).ToNot(BeNil())

				// Verify the list response format
				Expect(fetched.Items).To(HaveLen(1))
				Expect(fetched.TotalCount).To(Equal(int32(1)))

				// Verify the item content
				item := fetched.Items[0]
				Expect(item.Date).To(Equal(date))
				Expect(item.Title).To(Equal("Test Entry Title"))
				Expect(item.Body).To(Equal("This is a test body for the diary entry."))
				Expect(item.Tags).To(Equal([]string{"test", "ginkgo"}))
			})
		})
	})
})
