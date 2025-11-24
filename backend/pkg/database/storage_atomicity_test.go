package database_test

import (
	"log/slog"
	"os"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
)

var _ = Describe("Storage Atomicity and Transactions", func() {
	var (
		storage database.Storage
		logger  *slog.Logger
		userID  string
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		cfg := &config.Config{
			DBPath: ":memory:",
		}
		storage = database.NewStorage(logger, cfg)
		Expect(storage.Open()).To(Succeed())

		userID = "test-user-id"
	})

	AfterEach(func() {
		storage.Close()
	})

	Describe("PutItem atomicity", func() {
		It("should create both item and change record atomically", func() {
			testItem := &models.Item{
				UserID: userID,
				Date:   "2024-01-15",
				Title:  "Atomic Test Entry",
				Body:   "This tests atomic operations",
				Tags:   models.StringList{"atomic", "test"},
			}

			// Put the item
			err := storage.PutItem(userID, testItem)
			Expect(err).NotTo(HaveOccurred())

			// Verify item was created
			retrievedItem, err := storage.GetItem(userID, testItem.Date)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedItem.Title).To(Equal("Atomic Test Entry"))

			// Verify change record was created
			changes, err := storage.GetChangesSince(userID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(1))
			Expect(changes[0].OperationType).To(Equal(models.OperationTypeCreated))
			Expect(changes[0].Date).To(Equal(testItem.Date))
		})

		It("should update both item and create change record atomically", func() {
			testItem := &models.Item{
				UserID: userID,
				Date:   "2024-01-15",
				Title:  "Original Title",
				Body:   "Original body",
				Tags:   models.StringList{"original"},
			}

			// Create initial item
			err := storage.PutItem(userID, testItem)
			Expect(err).NotTo(HaveOccurred())

			// Update the item
			testItem.Title = "Updated Title"
			testItem.Body = "Updated body"
			testItem.Tags = models.StringList{"updated"}

			err = storage.PutItem(userID, testItem)
			Expect(err).NotTo(HaveOccurred())

			// Verify item was updated
			retrievedItem, err := storage.GetItem(userID, testItem.Date)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedItem.Title).To(Equal("Updated Title"))
			Expect(retrievedItem.Body).To(Equal("Updated body"))

			// Verify both create and update change records exist
			changes, err := storage.GetChangesSince(userID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(2))
			Expect(changes[0].OperationType).To(Equal(models.OperationTypeCreated))
			Expect(changes[1].OperationType).To(Equal(models.OperationTypeUpdated))
		})
	})

	Describe("DeleteItem atomicity", func() {
		BeforeEach(func() {
			// Create an item to delete
			testItem := &models.Item{
				UserID: userID,
				Date:   "2024-01-15",
				Title:  "Item to Delete",
				Body:   "This item will be deleted",
				Tags:   models.StringList{"delete-test"},
			}
			err := storage.PutItem(userID, testItem)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete item and create change record atomically", func() {
			// Delete the item
			err := storage.DeleteItem(userID, "2024-01-15")
			Expect(err).NotTo(HaveOccurred())

			// Verify item was deleted
			_, err = storage.GetItem(userID, "2024-01-15")
			Expect(err).To(Equal(database.ErrNotFound))

			// Verify change records exist (create + delete)
			changes, err := storage.GetChangesSince(userID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(2))
			Expect(changes[0].OperationType).To(Equal(models.OperationTypeCreated))
			Expect(changes[1].OperationType).To(Equal(models.OperationTypeDeleted))

			// Verify delete change has item snapshot
			deleteChange := changes[1]
			Expect(deleteChange.ItemSnapshot).NotTo(BeNil())
			Expect(deleteChange.ItemSnapshot.Title).To(Equal("Item to Delete"))
		})

		It("should handle deletion of non-existent item", func() {
			err := storage.DeleteItem(userID, "non-existent-date")
			Expect(err).To(Equal(database.ErrNotFound))

			// Verify no additional change records were created
			changes, err := storage.GetChangesSince(userID, 0, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(1)) // Only the create from BeforeEach
		})
	})

	Describe("Concurrent operations", func() {
		It("should handle concurrent PutItem operations safely", func() {
			const numGoroutines = 10
			var wg sync.WaitGroup
			errors := make(chan error, numGoroutines)

			// Launch concurrent PutItem operations
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					testItem := &models.Item{
						UserID: userID,
						Date:   "2024-01-15", // Same date for all
						Title:  "Concurrent Test",
						Body:   "Concurrent operation test",
						Tags:   models.StringList{"concurrent"},
					}
					err := storage.PutItem(userID, testItem)
					if err != nil {
						errors <- err
					}
				}()
			}

			wg.Wait()
			close(errors)

			// Check for errors
			for err := range errors {
				Expect(err).NotTo(HaveOccurred())
			}

			// Verify final state - should have one item
			item, err := storage.GetItem(userID, "2024-01-15")
			Expect(err).NotTo(HaveOccurred())
			Expect(item.Title).To(Equal("Concurrent Test"))

			// Verify change records - should have 1 create + (numGoroutines-1) updates
			changes, err := storage.GetChangesSince(userID, 0, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(numGoroutines))

			// First should be create, rest should be updates
			Expect(changes[0].OperationType).To(Equal(models.OperationTypeCreated))
			for i := 1; i < len(changes); i++ {
				Expect(changes[i].OperationType).To(Equal(models.OperationTypeUpdated))
			}
		})

		It("should handle concurrent operations on different dates safely", func() {
			const numGoroutines = 5
			var wg sync.WaitGroup
			errors := make(chan error, numGoroutines)

			// Launch concurrent PutItem operations on different dates
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					dateStr := "2024-01-" + string(rune('1'+index))
					testItem := &models.Item{
						UserID: userID,
						Date:   dateStr, // Different dates
						Title:  "Concurrent Test " + dateStr,
						Body:   "Concurrent operation test",
						Tags:   models.StringList{"concurrent"},
					}
					err := storage.PutItem(userID, testItem)
					if err != nil {
						errors <- err
					}
				}(i)
			}

			wg.Wait()
			close(errors)

			// Check for errors
			for err := range errors {
				Expect(err).NotTo(HaveOccurred())
			}

			// Verify all items were created
			for i := 0; i < numGoroutines; i++ {
				date := "2024-01-" + string(rune('1'+i))
				item, err := storage.GetItem(userID, date)
				Expect(err).NotTo(HaveOccurred())
				Expect(item.Date).To(Equal(date))
			}

			// Verify change records
			changes, err := storage.GetChangesSince(userID, 0, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(numGoroutines))

			// All should be creates since they're on different dates
			for _, change := range changes {
				Expect(change.OperationType).To(Equal(models.OperationTypeCreated))
			}
		})
	})

	Describe("Data consistency", func() {
		It("should maintain consistency between items and change records", func() {
			// Create multiple items
			dates := []string{"2024-01-15", "2024-01-16", "2024-01-17"}
			for _, date := range dates {
				testItem := &models.Item{
					UserID: userID,
					Date:   date,
					Title:  "Entry for " + date,
					Body:   "Body for " + date,
					Tags:   models.StringList{"consistency-test"},
				}
				err := storage.PutItem(userID, testItem)
				Expect(err).NotTo(HaveOccurred())
			}

			// Update one item
			updateItem := &models.Item{
				UserID: userID,
				Date:   "2024-01-16",
				Title:  "Updated Entry for 2024-01-16",
				Body:   "Updated body for 2024-01-16",
				Tags:   models.StringList{"consistency-test", "updated"},
			}
			err := storage.PutItem(userID, updateItem)
			Expect(err).NotTo(HaveOccurred())

			// Delete one item
			err = storage.DeleteItem(userID, "2024-01-17")
			Expect(err).NotTo(HaveOccurred())

			// Verify final state
			items, _, err := storage.GetItems(userID, database.SearchParams{})
			Expect(err).NotTo(HaveOccurred())
			Expect(items).To(HaveLen(2)) // Two remaining items

			// Verify change records match operations
			changes, err := storage.GetChangesSince(userID, 0, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes).To(HaveLen(5)) // 3 creates + 1 update + 1 delete

			operationCounts := make(map[models.OperationType]int)
			for _, change := range changes {
				operationCounts[change.OperationType]++
			}

			Expect(operationCounts[models.OperationTypeCreated]).To(Equal(3))
			Expect(operationCounts[models.OperationTypeUpdated]).To(Equal(1))
			Expect(operationCounts[models.OperationTypeDeleted]).To(Equal(1))
		})
	})
})
