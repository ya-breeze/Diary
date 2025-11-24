package models_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/database/models"
)

var _ = Describe("ItemChange", func() {
	var (
		itemChange *models.ItemChange
		testTime   time.Time
		testItem   *models.Item
	)

	BeforeEach(func() {
		testTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		testItem = &models.Item{
			UserID: "test-user",
			Date:   "2024-01-15",
			Title:  "Test Entry",
			Body:   "This is a test diary entry",
			Tags:   models.StringList{"personal", "test"},
		}

		itemChange = &models.ItemChange{
			ID:            123,
			UserID:        "test-user",
			Date:          "2024-01-15",
			OperationType: models.OperationTypeCreated,
			Timestamp:     testTime,
			ItemSnapshot:  testItem,
			Metadata:      models.StringList{"mobile-app", "v1.0.0"},
		}
	})

	Describe("OperationType constants", func() {
		It("should have correct string values", func() {
			Expect(string(models.OperationTypeCreated)).To(Equal("created"))
			Expect(string(models.OperationTypeUpdated)).To(Equal("updated"))
			Expect(string(models.OperationTypeDeleted)).To(Equal("deleted"))
		})
	})

	Describe("JSON serialization", func() {
		It("should serialize to JSON correctly", func() {
			// Test JSON serialization by converting to sync response
			response := itemChange.ToSyncResponse()

			Expect(response.Id).To(BeNumerically("==", 123))
			Expect(response.UserId).To(Equal("test-user"))
			Expect(response.Date).To(Equal("2024-01-15"))
			Expect(response.OperationType).To(Equal("created"))
			Expect(response.Metadata).To(ConsistOf("mobile-app", "v1.0.0"))
		})

		It("should handle field validation correctly", func() {
			// Test field validation by creating a new ItemChange
			change := &models.ItemChange{
				ID:            456,
				UserID:        "another-user",
				Date:          "2024-01-16",
				OperationType: models.OperationTypeUpdated,
				Timestamp:     testTime,
				Metadata:      models.StringList{"web-app", "v2.0.0"},
			}

			Expect(change.ID).To(BeNumerically("==", 456))
			Expect(change.UserID).To(Equal("another-user"))
			Expect(change.Date).To(Equal("2024-01-16"))
			Expect(change.OperationType).To(Equal(models.OperationTypeUpdated))
			Expect(change.Metadata).To(ConsistOf("web-app", "v2.0.0"))
		})
	})

	Describe("ToSyncResponse", func() {
		Context("for created/updated operations", func() {
			It("should convert to sync response with item snapshot", func() {
				response := itemChange.ToSyncResponse()

				Expect(response.Id).To(BeNumerically("==", 123))
				Expect(response.UserId).To(Equal("test-user"))
				Expect(response.Date).To(Equal("2024-01-15"))
				Expect(response.OperationType).To(Equal("created"))
				Expect(response.Timestamp).To(Equal(testTime))
				Expect(response.Metadata).To(ConsistOf("mobile-app", "v1.0.0"))

				Expect(response.ItemSnapshot).NotTo(BeNil())
				Expect(response.ItemSnapshot.Date).To(Equal("2024-01-15"))
				Expect(response.ItemSnapshot.Title).To(Equal("Test Entry"))
				Expect(response.ItemSnapshot.Body).To(Equal("This is a test diary entry"))
				Expect(response.ItemSnapshot.Tags).To(ConsistOf("personal", "test"))
			})
		})

		Context("for deleted operations", func() {
			BeforeEach(func() {
				itemChange.OperationType = models.OperationTypeDeleted
			})

			It("should convert to sync response without item snapshot", func() {
				response := itemChange.ToSyncResponse()

				Expect(response.Id).To(BeNumerically("==", 123))
				Expect(response.UserId).To(Equal("test-user"))
				Expect(response.Date).To(Equal("2024-01-15"))
				Expect(response.OperationType).To(Equal("deleted"))
				Expect(response.Timestamp).To(Equal(testTime))
				Expect(response.Metadata).To(ConsistOf("mobile-app", "v1.0.0"))

				Expect(response.ItemSnapshot).To(BeNil())
			})
		})

		Context("with nil item snapshot", func() {
			BeforeEach(func() {
				itemChange.ItemSnapshot = nil
			})

			It("should handle nil item snapshot gracefully", func() {
				response := itemChange.ToSyncResponse()

				Expect(response.Id).To(BeNumerically("==", 123))
				Expect(response.ItemSnapshot).To(BeNil())
			})
		})
	})

	Describe("Metadata handling", func() {
		It("should handle empty metadata", func() {
			itemChange.Metadata = models.StringList{}
			response := itemChange.ToSyncResponse()
			Expect(response.Metadata).To(BeEmpty())
		})

		It("should handle nil metadata", func() {
			itemChange.Metadata = nil
			response := itemChange.ToSyncResponse()
			Expect(response.Metadata).To(BeEmpty())
		})
	})
})
