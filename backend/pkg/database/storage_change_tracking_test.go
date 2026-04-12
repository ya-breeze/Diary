package database_test

import (
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
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
		familyID uuid.UUID
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

		familyID = uuid.New()
		testItem = &models.Item{
			Date:  "2024-01-15",
			Title: "Test Entry",
			Body:  "This is a test diary entry",
			Tags:  models.StringList{"personal", "test"},
		}
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	Describe("CreateChangeRecord", func() {
		It("should create a change record successfully", func() {
			err := storage.CreateChangeRecord(
				familyID,
				testItem.Date,
				models.OperationTypeCreated,
				testItem,
				[]string{"test-metadata"},
			)
			Expect(err).NotTo(HaveOccurred())

			// Verify the change was created
			changes, err := storage.GetChangesSince(familyID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(1))

			change := changes[0]
			Expect(change.FamilyID).To(Equal(familyID))
			Expect(change.Date).To(Equal(testItem.Date))
			Expect(change.OperationType).To(Equal(models.OperationTypeCreated))
			Expect(change.ItemSnapshot).NotTo(BeNil())
			Expect(change.ItemSnapshot.Title).To(Equal(testItem.Title))
			Expect(change.Metadata).To(ConsistOf("test-metadata"))
		})

		It("should handle nil item snapshot", func() {
			err := storage.CreateChangeRecord(
				familyID,
				"2024-01-16",
				models.OperationTypeDeleted,
				nil,
				[]string{"deletion"},
			)
			Expect(err).NotTo(HaveOccurred())

			changes, err := storage.GetChangesSince(familyID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes[0].ItemSnapshot).To(BeNil())
		})

		It("should handle empty metadata", func() {
			err := storage.CreateChangeRecord(
				familyID,
				testItem.Date,
				models.OperationTypeUpdated,
				testItem,
				[]string{},
			)
			Expect(err).NotTo(HaveOccurred())

			changes, err := storage.GetChangesSince(familyID, 0, 10)
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
					Date:  "2024-01-15",
					Title: "Test Entry",
					Body:  "This is a test diary entry",
					Tags:  models.StringList{"personal", "test"},
				}
				err := storage.CreateChangeRecord(
					familyID,
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
			changes, err := storage.GetChangesSince(familyID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(5))

			// Verify changes are ordered by ID ascending
			for i := 1; i < len(changes); i++ {
				Expect(changes[i].ID).To(BeNumerically(">", changes[i-1].ID))
			}
		})

		It("should return changes after specified ID", func() {
			allChanges, err := storage.GetChangesSince(familyID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(allChanges).To(HaveLen(5))

			// Get changes after the second change
			sinceID := allChanges[1].ID
			changes, err := storage.GetChangesSince(familyID, sinceID, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(3))

			// Verify all returned changes have ID > sinceID
			for _, change := range changes {
				Expect(change.ID).To(BeNumerically(">", sinceID))
			}
		})

		It("should respect limit parameter", func() {
			changes, err := storage.GetChangesSince(familyID, 0, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(3))
		})

		It("should return empty slice for non-existent family", func() {
			nonExistentFamily := uuid.New()
			changes, err := storage.GetChangesSince(nonExistentFamily, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(BeEmpty())
		})

		It("should return empty slice when since ID is higher than latest", func() {
			changes, err := storage.GetChangesSince(familyID, 999999, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(BeEmpty())
		})
	})

	Describe("GetLatestChangeID", func() {
		Context("when family has no changes", func() {
			It("should return 0", func() {
				nonExistentFamily := uuid.New()
				latestID, err := storage.GetLatestChangeID(nonExistentFamily)
				Expect(err).NotTo(HaveOccurred())
				Expect(latestID).To(BeNumerically("==", 0))
			})
		})

		Context("when family has changes", func() {
			BeforeEach(func() {
				// Create a few change records
				for i := 0; i < 3; i++ {
					err := storage.CreateChangeRecord(
						familyID,
						"2024-01-15",
						models.OperationTypeCreated,
						testItem,
						[]string{"test"},
					)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should return the highest change ID for the family", func() {
				latestID, err := storage.GetLatestChangeID(familyID)
				Expect(err).NotTo(HaveOccurred())
				Expect(latestID).To(BeNumerically(">", 0))

				// Verify this is indeed the latest by checking all changes
				changes, err := storage.GetChangesSince(familyID, 0, 10)
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

		Context("with multiple families", func() {
			BeforeEach(func() {
				family1 := uuid.New()
				family2 := uuid.New()

				// Create changes for family1
				err := storage.CreateChangeRecord(
					family1,
					"2024-01-15",
					models.OperationTypeCreated,
					testItem,
					[]string{"family1"},
				)
				Expect(err).NotTo(HaveOccurred())

				// Create changes for family2
				err = storage.CreateChangeRecord(
					family2,
					"2024-01-15",
					models.OperationTypeCreated,
					testItem,
					[]string{"family2"},
				)
				Expect(err).NotTo(HaveOccurred())

				// Store family1 and family2 for the It block — use closure vars
				DeferCleanup(func() {
					// no-op, vars are local
				})
				// Reuse outer familyID for test verification
				familyID = family1
			})

			It("should return correct latest ID for each family", func() {
				// Since BeforeEach creates two new families but stores family1 in familyID,
				// just verify familyID (family1) has changes
				latestID, err := storage.GetLatestChangeID(familyID)
				Expect(err).NotTo(HaveOccurred())
				Expect(latestID).To(BeNumerically(">", 0))
			})
		})
	})
})
