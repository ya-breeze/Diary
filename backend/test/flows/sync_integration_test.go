package flows_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/generated/goclient"
)

var _ = Describe("Sync Integration Flow", func() {
	var setup *SharedTestSetup

	BeforeEach(func() {
		setup = SetupTestEnvironment()
	})

	AfterEach(func() {
		setup.TeardownTestEnvironment()
	})

	Describe("Complete synchronization workflow", func() {
		BeforeEach(func() {
			// Login and get token (also configures API client headers)
			setup.LoginAndGetToken()
		})

		It("should track changes when creating, updating, and deleting items", func() {
			// Step 1: Get initial sync state (should be empty)
			syncResponse, httpResponse, err := setup.APIClient.SyncAPI.GetChanges(context.Background()).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(BeEmpty())
			Expect(syncResponse.HasMore).To(BeFalse())

			// Step 2: Create a diary item
			createRequest := goclient.ItemsRequest{
				Date:  "2024-01-15",
				Title: "My First Entry",
				Body:  "This is my first diary entry for sync testing",
				Tags:  []string{"personal", "sync-test"},
			}

			itemResponse, httpResponse, err := setup.APIClient.ItemsAPI.PutItems(context.Background()).
				ItemsRequest(createRequest).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(itemResponse.Title).To(Equal("My First Entry"))

			// Step 3: Check sync changes after creation
			syncResponse, httpResponse, err = setup.APIClient.SyncAPI.GetChanges(context.Background()).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(HaveLen(1))

			createChange := syncResponse.Changes[0]
			Expect(createChange.OperationType).To(Equal("created"))
			Expect(createChange.Date).To(Equal("2024-01-15"))
			Expect(createChange.ItemSnapshot.IsSet()).To(BeTrue())
			itemSnapshot, _ := createChange.GetItemSnapshotOk()
			Expect(itemSnapshot).NotTo(BeNil())
			Expect(itemSnapshot.Title).To(Equal("My First Entry"))
			Expect(itemSnapshot.Body).To(Equal("This is my first diary entry for sync testing"))
			Expect(itemSnapshot.Tags).To(ConsistOf("personal", "sync-test"))

			firstChangeID := createChange.Id

			// Step 4: Update the same item
			updateRequest := goclient.ItemsRequest{
				Date:  "2024-01-15",
				Title: "My Updated Entry",
				Body:  "This is my updated diary entry for sync testing",
				Tags:  []string{"personal", "sync-test", "updated"},
			}

			itemResponse, httpResponse, err = setup.APIClient.ItemsAPI.PutItems(context.Background()).
				ItemsRequest(updateRequest).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(itemResponse.Title).To(Equal("My Updated Entry"))

			// Step 5: Check sync changes after update
			syncResponse, httpResponse, err = setup.APIClient.SyncAPI.GetChanges(context.Background()).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(HaveLen(2))

			// Find the update change (should be the second one)
			var updateChange *goclient.SyncChangeResponse
			for _, change := range syncResponse.Changes {
				if change.OperationType == "updated" {
					updateChange = &change
					break
				}
			}
			Expect(updateChange).NotTo(BeNil())
			Expect(updateChange.Date).To(Equal("2024-01-15"))
			Expect(updateChange.ItemSnapshot.IsSet()).To(BeTrue())
			updateSnapshot, _ := updateChange.GetItemSnapshotOk()
			Expect(updateSnapshot).NotTo(BeNil())
			Expect(updateSnapshot.Title).To(Equal("My Updated Entry"))
			Expect(updateSnapshot.Tags).To(ConsistOf("personal", "sync-test", "updated"))

			// Step 6: Create another item on a different date
			secondRequest := goclient.ItemsRequest{
				Date:  "2024-01-16",
				Title: "Second Entry",
				Body:  "This is my second diary entry",
				Tags:  []string{"work", "sync-test"},
			}

			_, httpResponse, err = setup.APIClient.ItemsAPI.PutItems(context.Background()).ItemsRequest(secondRequest).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))

			// Step 7: Test incremental sync (get changes since first change)
			syncResponse, httpResponse, err = setup.APIClient.SyncAPI.GetChanges(context.Background()).
				Since(firstChangeID).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(HaveLen(2)) // Update + second create

			// Verify the changes are in correct order
			Expect(syncResponse.Changes[0].OperationType).To(Equal("updated"))
			Expect(syncResponse.Changes[1].OperationType).To(Equal("created"))
			Expect(syncResponse.Changes[1].Date).To(Equal("2024-01-16"))
		})

		It("should handle pagination correctly", func() {
			// Create multiple items to test pagination
			for i := 1; i <= 5; i++ {
				request := goclient.ItemsRequest{
					Date:  "2024-01-15",
					Title: "Entry " + string(rune('0'+i)),
					Body:  "This is entry number " + string(rune('0'+i)),
					Tags:  []string{"pagination-test"},
				}

				_, httpResponse, err := setup.APIClient.ItemsAPI.PutItems(context.Background()).ItemsRequest(request).Execute()
				Expect(err).NotTo(HaveOccurred())
				defer httpResponse.Body.Close()
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			}

			// Test pagination with limit=2
			page1, httpResponse, err := setup.APIClient.SyncAPI.GetChanges(context.Background()).Limit(2).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(page1.Changes).To(HaveLen(2))
			Expect(page1.HasMore).To(BeTrue())
			Expect(page1.NextId).NotTo(BeNil())
			Expect(*page1.NextId).To(BeNumerically(">", 0))

			// Get next page
			page2, httpResponse, err := setup.APIClient.SyncAPI.GetChanges(context.Background()).
				Since(*page1.NextId).Limit(2).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(page2.Changes).To(HaveLen(2))
			Expect(page2.HasMore).To(BeTrue())

			// Get final page
			page3, httpResponse, err := setup.APIClient.SyncAPI.GetChanges(context.Background()).
				Since(*page2.NextId).Limit(2).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(page3.Changes).To(HaveLen(1))
			Expect(page3.HasMore).To(BeFalse())

			// Verify no duplicate changes across pages
			allIDs := make(map[int32]bool)
			for _, change := range page1.Changes {
				allIDs[change.Id] = true
			}
			for _, change := range page2.Changes {
				Expect(allIDs[change.Id]).To(BeFalse(), "Found duplicate change ID: %d", change.Id)
				allIDs[change.Id] = true
			}
			for _, change := range page3.Changes {
				Expect(allIDs[change.Id]).To(BeFalse(), "Found duplicate change ID: %d", change.Id)
			}
		})

		It("should handle authentication properly", func() {
			// Test without authentication
			unauthenticatedClient := goclient.NewAPIClient(goclient.NewConfiguration())
			unauthenticatedClient.GetConfig().Servers = goclient.ServerConfigurations{
				{
					URL:         setup.ServerAddr,
					Description: "Test server",
				},
			}

			_, httpResponse, err := unauthenticatedClient.SyncAPI.GetChanges(context.Background()).Execute()
			Expect(err).To(HaveOccurred())
			if httpResponse != nil {
				defer httpResponse.Body.Close()
				Expect(httpResponse.StatusCode).To(Equal(http.StatusUnauthorized))
			}
		})

		It("should handle empty sync state correctly", func() {
			// Test sync when no changes exist
			syncResponse, httpResponse, err := setup.APIClient.SyncAPI.GetChanges(context.Background()).Execute()
			Expect(err).NotTo(HaveOccurred())
			defer httpResponse.Body.Close()
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(BeEmpty())
			Expect(syncResponse.HasMore).To(BeFalse())
			if syncResponse.NextId != nil {
				Expect(*syncResponse.NextId).To(BeNumerically("==", 0))
			}
		})
	})
})
