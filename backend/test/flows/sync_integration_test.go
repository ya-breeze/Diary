package flows_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			setup.LoginAndGetToken()
		})

		It("should track changes when creating, updating, and deleting items", func() {
			// Step 1: Get initial sync state (should be empty)
			syncResponse, httpResponse, err := setup.APIClient.GetChanges(context.Background(), 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(BeEmpty())
			Expect(syncResponse.HasMore).To(BeFalse())

			// Step 2: Create a diary item
			_, httpResponse, err = setup.APIClient.PutItems(context.Background(), "2024-01-15", "My First Entry",
				"This is my first diary entry for sync testing", []string{"personal", "sync-test"})
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))

			// Step 3: Check sync changes after creation
			syncResponse, httpResponse, err = setup.APIClient.GetChanges(context.Background(), 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(HaveLen(1))

			createChange := syncResponse.Changes[0]
			Expect(createChange.OperationType).To(Equal("created"))
			Expect(createChange.Date).To(Equal("2024-01-15"))
			Expect(createChange.ItemSnapshot).NotTo(BeNil())
			Expect(createChange.ItemSnapshot.Title).To(Equal("My First Entry"))
			Expect(createChange.ItemSnapshot.Body).To(Equal("This is my first diary entry for sync testing"))
			Expect(createChange.ItemSnapshot.Tags).To(ConsistOf("personal", "sync-test"))

			firstChangeID := createChange.Id

			// Step 4: Update the same item
			_, httpResponse, err = setup.APIClient.PutItems(context.Background(), "2024-01-15", "My Updated Entry",
				"This is my updated diary entry for sync testing", []string{"personal", "sync-test", "updated"})
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))

			// Step 5: Check sync changes after update
			syncResponse, httpResponse, err = setup.APIClient.GetChanges(context.Background(), 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(HaveLen(2))

			var updateChange *TestSyncChangeResponse
			for i, change := range syncResponse.Changes {
				if change.OperationType == "updated" {
					updateChange = &syncResponse.Changes[i]
					break
				}
			}
			Expect(updateChange).NotTo(BeNil())
			Expect(updateChange.Date).To(Equal("2024-01-15"))
			Expect(updateChange.ItemSnapshot).NotTo(BeNil())
			Expect(updateChange.ItemSnapshot.Title).To(Equal("My Updated Entry"))
			Expect(updateChange.ItemSnapshot.Tags).To(ConsistOf("personal", "sync-test", "updated"))

			// Step 6: Create another item on a different date
			_, httpResponse, err = setup.APIClient.PutItems(context.Background(), "2024-01-16", "Second Entry",
				"This is my second diary entry", []string{"work", "sync-test"})
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))

			// Step 7: Test incremental sync (get changes since first change)
			syncResponse, httpResponse, err = setup.APIClient.GetChanges(context.Background(), firstChangeID, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(HaveLen(2)) // Update + second create

			Expect(syncResponse.Changes[0].OperationType).To(Equal("updated"))
			Expect(syncResponse.Changes[1].OperationType).To(Equal("created"))
			Expect(syncResponse.Changes[1].Date).To(Equal("2024-01-16"))
		})

		It("should handle pagination correctly", func() {
			// Create 5 items to test pagination (5 updates = 5 changes)
			for i := 1; i <= 5; i++ {
				_, httpResponse, err := setup.APIClient.PutItems(context.Background(), "2024-01-15",
					fmt.Sprintf("Entry %d", i), fmt.Sprintf("This is entry number %d", i), []string{"pagination-test"})
				Expect(err).NotTo(HaveOccurred())
				Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			}

			// Test pagination with limit=2
			page1, httpResponse, err := setup.APIClient.GetChanges(context.Background(), 0, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(page1.Changes).To(HaveLen(2))
			Expect(page1.HasMore).To(BeTrue())
			Expect(page1.NextId).NotTo(BeNil())
			Expect(*page1.NextId).To(BeNumerically(">", 0))

			// Get next page
			page2, httpResponse, err := setup.APIClient.GetChanges(context.Background(), *page1.NextId, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(page2.Changes).To(HaveLen(2))
			Expect(page2.HasMore).To(BeTrue())

			// Get final page
			page3, httpResponse, err := setup.APIClient.GetChanges(context.Background(), *page2.NextId, 2)
			Expect(err).NotTo(HaveOccurred())
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
			unauthClient := newTestAPIClient(setup.ServerAddr)
			_, httpResp, err := unauthClient.GetChanges(context.Background(), 0, 0)
			Expect(err).To(HaveOccurred())
			if httpResp != nil {
				Expect(httpResp.StatusCode).To(Equal(http.StatusUnauthorized))
			}
		})

		It("should handle empty sync state correctly", func() {
			syncResponse, httpResponse, err := setup.APIClient.GetChanges(context.Background(), 0, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
			Expect(syncResponse.Changes).To(BeEmpty())
			Expect(syncResponse.HasMore).To(BeFalse())
			if syncResponse.NextId != nil {
				Expect(*syncResponse.NextId).To(BeNumerically("==", 0))
			}
		})
	})
})
