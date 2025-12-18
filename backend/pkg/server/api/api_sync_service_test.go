package api_test

import (
	"context"
	"log/slog"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/api"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

var _ = Describe("SyncAPIService", func() {
	var (
		service goserver.SyncAPIService
		logger  *slog.Logger
		storage database.Storage
		ctx     context.Context
		userID  string
		tempDir string
	)

	// Create context outside of BeforeEach to avoid fatcontext linting issue
	userID = "test-user-id"
	ctx = context.WithValue(context.Background(), common.UserIDKey, userID)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

		var err error
		tempDir, err = os.MkdirTemp("", "sync_test")
		Expect(err).NotTo(HaveOccurred())

		cfg := &config.Config{
			DataPath: tempDir,
		}
		storage = database.NewStorage(logger, cfg)
		Expect(storage.Open()).To(Succeed())

		service = api.NewSyncAPIService(logger, storage)
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	Describe("GetChanges", func() {
		Context("when user ID is missing from context", func() {
			It("should return unauthorized", func() {
				emptyCtx := context.Background()
				response, err := service.GetChanges(emptyCtx, 0, 100)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(401))
			})
		})

		Context("when user has no changes", func() {
			It("should return empty changes list", func() {
				response, err := service.GetChanges(ctx, 0, 100)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				syncResponse, ok := response.Body.(goserver.SyncResponse)
				Expect(ok).To(BeTrue())
				Expect(syncResponse.Changes).To(BeEmpty())
				Expect(syncResponse.HasMore).To(BeFalse())
				Expect(syncResponse.NextId).To(BeNumerically("==", 0))
			})
		})

		Context("when user has changes", func() {
			BeforeEach(func() {
				// Create test changes
				testItem := &models.Item{
					UserID: userID,
					Date:   "2024-01-15",
					Title:  "Test Entry",
					Body:   "This is a test diary entry",
					Tags:   models.StringList{"personal", "test"},
				}

				for i := 0; i < 5; i++ {
					err := storage.CreateChangeRecord(
						userID,
						testItem.Date,
						models.OperationTypeCreated,
						testItem,
						[]string{"test-metadata"},
					)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should return all changes when since=0", func() {
				response, err := service.GetChanges(ctx, 0, 100)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				syncResponse, ok := response.Body.(goserver.SyncResponse)
				Expect(ok).To(BeTrue())
				Expect(syncResponse.Changes).To(HaveLen(5))
				Expect(syncResponse.HasMore).To(BeFalse())

				// Verify change content
				change := syncResponse.Changes[0]
				Expect(change.UserId).To(Equal(userID))
				Expect(change.Date).To(Equal("2024-01-15"))
				Expect(change.OperationType).To(Equal("created"))
				Expect(change.ItemSnapshot).NotTo(BeNil())
				Expect(change.ItemSnapshot.Title).To(Equal("Test Entry"))
				Expect(change.Metadata).To(ConsistOf("test-metadata"))
			})

			It("should respect limit parameter", func() {
				response, err := service.GetChanges(ctx, 0, 3)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				syncResponse, ok := response.Body.(goserver.SyncResponse)
				Expect(ok).To(BeTrue())
				Expect(syncResponse.Changes).To(HaveLen(3))
				Expect(syncResponse.HasMore).To(BeTrue())
				Expect(syncResponse.NextId).To(BeNumerically(">", 0))
			})

			It("should return changes after specified ID", func() {
				// First, get all changes to find a middle ID
				allResponse, err := service.GetChanges(ctx, 0, 100)
				Expect(err).NotTo(HaveOccurred())

				allSyncResponse, ok := allResponse.Body.(goserver.SyncResponse)
				Expect(ok).To(BeTrue())
				Expect(allSyncResponse.Changes).To(HaveLen(5))

				// Get changes after the second change
				sinceID := allSyncResponse.Changes[1].Id
				response, err := service.GetChanges(ctx, sinceID, 100)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				syncResponse, ok := response.Body.(goserver.SyncResponse)
				Expect(ok).To(BeTrue())
				Expect(syncResponse.Changes).To(HaveLen(3))

				// Verify all returned changes have ID > sinceID
				for _, change := range syncResponse.Changes {
					Expect(change.Id).To(BeNumerically(">", sinceID))
				}
			})

			It("should handle pagination correctly", func() {
				// Get first page
				response1, err := service.GetChanges(ctx, 0, 2)
				Expect(err).NotTo(HaveOccurred())

				syncResponse1, ok := response1.Body.(goserver.SyncResponse)
				Expect(ok).To(BeTrue())
				Expect(syncResponse1.Changes).To(HaveLen(2))
				Expect(syncResponse1.HasMore).To(BeTrue())

				// Get second page
				response2, err := service.GetChanges(ctx, syncResponse1.NextId, 2)
				Expect(err).NotTo(HaveOccurred())

				syncResponse2, ok := response2.Body.(goserver.SyncResponse)
				Expect(ok).To(BeTrue())
				Expect(syncResponse2.Changes).To(HaveLen(2))
				Expect(syncResponse2.HasMore).To(BeTrue())

				// Get final page
				response3, err := service.GetChanges(ctx, syncResponse2.NextId, 2)
				Expect(err).NotTo(HaveOccurred())

				syncResponse3, ok := response3.Body.(goserver.SyncResponse)
				Expect(ok).To(BeTrue())
				Expect(syncResponse3.Changes).To(HaveLen(1))
				Expect(syncResponse3.HasMore).To(BeFalse())

				// Verify no duplicate changes
				allIDs := make(map[int32]bool)
				for _, change := range syncResponse1.Changes {
					allIDs[change.Id] = true
				}
				for _, change := range syncResponse2.Changes {
					Expect(allIDs[change.Id]).To(BeFalse())
					allIDs[change.Id] = true
				}
				for _, change := range syncResponse3.Changes {
					Expect(allIDs[change.Id]).To(BeFalse())
				}
			})
		})

		Context("with invalid parameters", func() {
			It("should use default limit when limit is 0", func() {
				response, err := service.GetChanges(ctx, 0, 0)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(200))
			})

			It("should use default limit when limit is negative", func() {
				response, err := service.GetChanges(ctx, 0, -10)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(200))
			})

			It("should use default limit when limit exceeds maximum", func() {
				response, err := service.GetChanges(ctx, 0, 2000)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(200))
			})
		})

		Context("with deleted items", func() {
			BeforeEach(func() {
				// Create a change for a deleted item
				testItem := &models.Item{
					UserID: userID,
					Date:   "2024-01-15",
					Title:  "Deleted Entry",
					Body:   "This entry was deleted",
					Tags:   models.StringList{"deleted"},
				}

				err := storage.CreateChangeRecord(
					userID,
					testItem.Date,
					models.OperationTypeDeleted,
					testItem,
					[]string{"deletion"},
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should include deleted items in sync response", func() {
				response, err := service.GetChanges(ctx, 0, 100)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				syncResponse, ok := response.Body.(goserver.SyncResponse)
				Expect(ok).To(BeTrue())
				Expect(syncResponse.Changes).To(HaveLen(1))

				change := syncResponse.Changes[0]
				Expect(change.OperationType).To(Equal("deleted"))
				Expect(change.ItemSnapshot).NotTo(BeNil()) // Deleted items still have snapshot
				Expect(change.ItemSnapshot.Title).To(Equal("Deleted Entry"))
			})
		})
	})
})
