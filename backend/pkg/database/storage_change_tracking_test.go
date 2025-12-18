package database_test

import (
	"log/slog"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
)

var _ = Describe("Storage Change Tracking", func() {
	var (
		storage  database.Storage
		logger   *slog.Logger
		userID   string
		testItem *models.Item
		tempDir  string
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		var err error
		tempDir, err = os.MkdirTemp("", "changes_test")
		Expect(err).NotTo(HaveOccurred())

		cfg := &config.Config{
			DataPath: tempDir,
		}
		storage = database.NewStorage(logger, cfg)
		Expect(storage.Open()).To(Succeed())

		userID = "test-user-id"
		testItem = &models.Item{
			UserID: userID,
			Date:   "2024-01-15",
			Title:  "Test Entry",
			Body:   "This is a test diary entry",
			Tags:   models.StringList{"personal", "test"},
		}
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	Describe("CreateChangeRecord", func() {
		It("should create a change record successfully", func() {
			err := storage.CreateChangeRecord(
				userID,
				testItem.Date,
				models.OperationTypeCreated,
				testItem,
				[]string{"test-metadata"},
			)
			Expect(err).NotTo(HaveOccurred())

			// Verify the change was created
			changes, err := storage.GetChangesSince(userID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(1))

			change := changes[0]
			Expect(change.UserID).To(Equal(userID))
			Expect(change.Date).To(Equal(testItem.Date))
			Expect(change.OperationType).To(Equal(models.OperationTypeCreated))
			Expect(change.ItemSnapshot).NotTo(BeNil())
			Expect(change.ItemSnapshot.Title).To(Equal(testItem.Title))
			Expect(change.Metadata).To(ConsistOf("test-metadata"))
		})

		It("should handle nil item snapshot", func() {
			err := storage.CreateChangeRecord(
				userID,
				"2024-01-16",
				models.OperationTypeDeleted,
				nil,
				[]string{"deletion"},
			)
			Expect(err).NotTo(HaveOccurred())

			changes, err := storage.GetChangesSince(userID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes[0].ItemSnapshot).To(BeNil())
		})

		It("should handle empty metadata", func() {
			err := storage.CreateChangeRecord(
				userID,
				testItem.Date,
				models.OperationTypeUpdated,
				testItem,
				[]string{},
			)
			Expect(err).NotTo(HaveOccurred())

			changes, err := storage.GetChangesSince(userID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes[0].Metadata).To(BeEmpty())
		})
	})

	Describe("GetChangesSince", func() {
		BeforeEach(func() {
			// Create multiple change records
			for i := 0; i < 5; i++ {
				item := &models.Item{
					UserID: userID,
					Date:   "2024-01-15",
					Title:  "Test Entry",
					Body:   "This is a test diary entry",
					Tags:   models.StringList{"personal", "test"},
				}
				err := storage.CreateChangeRecord(
					userID,
					item.Date,
					models.OperationTypeCreated,
					item,
					[]string{"batch-test"},
				)
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(1 * time.Millisecond) // Ensure different timestamps
			}
		})

		It("should return all changes when since=0", func() {
			changes, err := storage.GetChangesSince(userID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(5))

			// Verify changes are ordered by ID ascending
			for i := 1; i < len(changes); i++ {
				Expect(changes[i].ID).To(BeNumerically(">", changes[i-1].ID))
			}
		})

		It("should return changes after specified ID", func() {
			allChanges, err := storage.GetChangesSince(userID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(allChanges).To(HaveLen(5))

			// Get changes after the second change
			sinceID := allChanges[1].ID
			changes, err := storage.GetChangesSince(userID, sinceID, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(3))

			// Verify all returned changes have ID > sinceID
			for _, change := range changes {
				Expect(change.ID).To(BeNumerically(">", sinceID))
			}
		})

		It("should respect limit parameter", func() {
			changes, err := storage.GetChangesSince(userID, 0, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(3))
		})

		It("should return empty slice for non-existent user", func() {
			changes, err := storage.GetChangesSince("non-existent-user", 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(BeEmpty())
		})

		It("should return empty slice when since ID is higher than latest", func() {
			changes, err := storage.GetChangesSince(userID, 999999, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(BeEmpty())
		})
	})

	Describe("GetLatestChangeID", func() {
		Context("when user has no changes", func() {
			It("should return 0", func() {
				latestID, err := storage.GetLatestChangeID("non-existent-user")
				Expect(err).NotTo(HaveOccurred())
				Expect(latestID).To(BeNumerically("==", 0))
			})
		})

		Context("when user has changes", func() {
			BeforeEach(func() {
				// Create a few change records
				for i := 0; i < 3; i++ {
					err := storage.CreateChangeRecord(
						userID,
						"2024-01-15",
						models.OperationTypeCreated,
						testItem,
						[]string{"test"},
					)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should return the highest change ID for the user", func() {
				latestID, err := storage.GetLatestChangeID(userID)
				Expect(err).NotTo(HaveOccurred())
				Expect(latestID).To(BeNumerically(">", 0))

				// Verify this is indeed the latest by checking all changes
				changes, err := storage.GetChangesSince(userID, 0, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(changes).NotTo(BeEmpty())

				maxID := changes[0].ID
				for _, change := range changes {
					if change.ID > maxID {
						maxID = change.ID
					}
				}
				Expect(latestID).To(Equal(maxID))
			})
		})

		Context("with multiple users", func() {
			BeforeEach(func() {
				// Create changes for user1
				err := storage.CreateChangeRecord(
					"user1",
					"2024-01-15",
					models.OperationTypeCreated,
					testItem,
					[]string{"user1"},
				)
				Expect(err).NotTo(HaveOccurred())

				// Create changes for user2
				err = storage.CreateChangeRecord(
					"user2",
					"2024-01-15",
					models.OperationTypeCreated,
					testItem,
					[]string{"user2"},
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return correct latest ID for each user", func() {
				latestID1, err := storage.GetLatestChangeID("user1")
				Expect(err).NotTo(HaveOccurred())

				latestID2, err := storage.GetLatestChangeID("user2")
				Expect(err).NotTo(HaveOccurred())

				Expect(latestID1).To(BeNumerically(">", 0))
				Expect(latestID2).To(BeNumerically(">", 0))
				Expect(latestID1).NotTo(Equal(latestID2))
			})
		})
	})
})
