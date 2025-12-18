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

// Helper function to create context with user ID for items tests
func createContextWithUserIDForItems(userID string) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, common.UserIDKey, userID)
}

// Helper function to assert successful PUT response and database save
func assertSuccessfulPutResponse(
	response goserver.ImplResponse,
	expectedTitle, expectedBody string,
	expectedTags []string,
	expectedDate string,
) {
	Expect(response.Code).To(Equal(200))

	itemsResponse, ok := response.Body.(goserver.ItemsResponse)
	Expect(ok).To(BeTrue())
	Expect(itemsResponse.Date).To(Equal(expectedDate))
	Expect(itemsResponse.Title).To(Equal(expectedTitle))
	Expect(itemsResponse.Body).To(Equal(expectedBody))
	Expect(itemsResponse.Tags).To(Equal(expectedTags))
}

// Helper function to verify item was saved to database
func verifyItemInDatabase(
	storage database.Storage,
	userID, expectedDate, expectedTitle, expectedBody string,
	expectedTags []string,
) {
	savedItem, err := storage.GetItem(userID, expectedDate)
	Expect(err).ToNot(HaveOccurred())
	Expect(savedItem.Title).To(Equal(expectedTitle))
	Expect(savedItem.Body).To(Equal(expectedBody))
	Expect(savedItem.Tags).To(Equal(models.StringList(expectedTags)))
}

var _ = Describe("ItemsAPIService", func() {
	var (
		service  goserver.ItemsAPIService
		logger   *slog.Logger
		storage  database.Storage
		ctx      context.Context
		userID   string
		testDate string
		tempDir  string
	)

	// Create context outside of BeforeEach to avoid fatcontext linting issue
	userID = "test-user-id"
	testDate = "2024-01-15"
	ctx = createContextWithUserIDForItems(userID)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		var err error
		tempDir, err = os.MkdirTemp("", "items_test")
		Expect(err).NotTo(HaveOccurred())

		cfg := &config.Config{
			DataPath: tempDir,
		}
		storage = database.NewStorage(logger, cfg)
		Expect(storage.Open()).To(Succeed())

		service = api.NewItemsAPIService(logger, storage)
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	Describe("GetItems", func() {
		Context("when no user ID in context", func() {
			It("should return 401 unauthorized", func() {
				emptyCtx := context.Background()
				response, err := service.GetItems(emptyCtx, testDate, "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(401))
			})
		})

		Context("when item does not exist (backward compatibility with date filter)", func() {
			It("should return empty item with navigation dates for the requested date", func() {
				response, err := service.GetItems(ctx, testDate, "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))
				Expect(itemsListResponse.TotalCount).To(Equal(int32(1)))

				// Verify the empty item has the correct date and empty content
				item := itemsListResponse.Items[0]
				Expect(item.Date).To(Equal(testDate))
				Expect(item.Title).To(Equal(""))
				Expect(item.Body).To(Equal(""))
				Expect(item.Tags).To(Equal([]string{}))
			})

			Context("when previous and next items exist", func() {
				BeforeEach(func() {
					// Create previous item
					prevItem := &models.Item{
						UserID: userID,
						Date:   "2024-01-14",
						Title:  "Previous Item",
						Body:   "Previous content",
					}
					Expect(storage.PutItem(userID, prevItem)).To(Succeed())

					// Create next item
					nextItem := &models.Item{
						UserID: userID,
						Date:   "2024-01-16",
						Title:  "Next Item",
						Body:   "Next content",
					}
					Expect(storage.PutItem(userID, nextItem)).To(Succeed())
				})

				It("should include navigation dates even for empty item", func() {
					response, err := service.GetItems(ctx, testDate, "", "")
					Expect(err).ToNot(HaveOccurred())
					Expect(response.Code).To(Equal(200))

					itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
					Expect(ok).To(BeTrue())
					Expect(itemsListResponse.Items).To(HaveLen(1))

					item := itemsListResponse.Items[0]
					Expect(item.Date).To(Equal(testDate))
					Expect(item.Title).To(Equal(""))
					Expect(item.Body).To(Equal(""))
					Expect(item.PreviousDate).ToNot(BeNil())
					Expect(*item.PreviousDate).To(Equal("2024-01-14"))
					Expect(item.NextDate).ToNot(BeNil())
					Expect(*item.NextDate).To(Equal("2024-01-16"))
				})
			})
		})

		Context("when item exists (backward compatibility with date filter)", func() {
			BeforeEach(func() {
				// Create a test item
				testItem := &models.Item{
					UserID: userID,
					Date:   testDate,
					Title:  "Test Title",
					Body:   "Test Body Content",
					Tags:   models.StringList{"tag1", "tag2"},
				}
				Expect(storage.PutItem(userID, testItem)).To(Succeed())
			})

			It("should return the item in list format with 200 status", func() {
				response, err := service.GetItems(ctx, testDate, "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))
				Expect(itemsListResponse.TotalCount).To(Equal(int32(1)))

				item := itemsListResponse.Items[0]
				Expect(item.Date).To(Equal(testDate))
				Expect(item.Title).To(Equal("Test Title"))
				Expect(item.Body).To(Equal("Test Body Content"))
				Expect(item.Tags).To(Equal([]string{"tag1", "tag2"}))
			})
		})

		Context("when previous and next items exist", func() {
			BeforeEach(func() {
				// Create previous item
				prevItem := &models.Item{
					UserID: userID,
					Date:   "2024-01-14",
					Title:  "Previous Item",
					Body:   "Previous content",
				}
				Expect(storage.PutItem(userID, prevItem)).To(Succeed())

				// Create current item
				currentItem := &models.Item{
					UserID: userID,
					Date:   testDate,
					Title:  "Current Item",
					Body:   "Current content",
				}
				Expect(storage.PutItem(userID, currentItem)).To(Succeed())

				// Create next item
				nextItem := &models.Item{
					UserID: userID,
					Date:   "2024-01-16",
					Title:  "Next Item",
					Body:   "Next content",
				}
				Expect(storage.PutItem(userID, nextItem)).To(Succeed())
			})

			It("should include previous and next dates", func() {
				response, err := service.GetItems(ctx, testDate, "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))

				item := itemsListResponse.Items[0]
				Expect(item.Date).To(Equal(testDate))
				Expect(item.PreviousDate).ToNot(BeNil())
				Expect(*item.PreviousDate).To(Equal("2024-01-14"))
				Expect(item.NextDate).ToNot(BeNil())
				Expect(*item.NextDate).To(Equal("2024-01-16"))
			})
		})

		Context("when searching by text", func() {
			BeforeEach(func() {
				// Create test items with different content
				items := []*models.Item{
					{
						UserID: userID,
						Date:   "2024-01-10",
						Title:  "Vacation Planning",
						Body:   "Planning my summer vacation to the beach",
						Tags:   models.StringList{"travel", "vacation"},
					},
					{
						UserID: userID,
						Date:   "2024-01-11",
						Title:  "Work Meeting",
						Body:   "Had an important meeting about the project",
						Tags:   models.StringList{"work", "meeting"},
					},
					{
						UserID: userID,
						Date:   "2024-01-12",
						Title:  "Beach Day",
						Body:   "Spent the day at the beach with family",
						Tags:   models.StringList{"family", "beach"},
					},
				}
				for _, item := range items {
					Expect(storage.PutItem(userID, item)).To(Succeed())
				}
			})

			It("should return items matching search text in title", func() {
				response, err := service.GetItems(ctx, "", "vacation", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))
				Expect(itemsListResponse.TotalCount).To(Equal(int32(1)))
				Expect(itemsListResponse.Items[0].Title).To(Equal("Vacation Planning"))
			})

			It("should return items matching search text in body", func() {
				response, err := service.GetItems(ctx, "", "beach", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(2))
				Expect(itemsListResponse.TotalCount).To(Equal(int32(2)))
			})

			It("should return empty list when no matches found", func() {
				response, err := service.GetItems(ctx, "", "nonexistent", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(BeEmpty())
				Expect(itemsListResponse.TotalCount).To(Equal(int32(0)))
			})
		})

		Context("when searching by tags", func() {
			BeforeEach(func() {
				// Create test items with different tags
				items := []*models.Item{
					{
						UserID: userID,
						Date:   "2024-01-10",
						Title:  "Work Project",
						Body:   "Working on the new project",
						Tags:   models.StringList{"work", "project"},
					},
					{
						UserID: userID,
						Date:   "2024-01-11",
						Title:  "Family Time",
						Body:   "Spending time with family",
						Tags:   models.StringList{"family", "personal"},
					},
					{
						UserID: userID,
						Date:   "2024-01-12",
						Title:  "Work Meeting",
						Body:   "Important work meeting",
						Tags:   models.StringList{"work", "meeting"},
					},
				}
				for _, item := range items {
					Expect(storage.PutItem(userID, item)).To(Succeed())
				}
			})

			It("should return items matching single tag", func() {
				response, err := service.GetItems(ctx, "", "", "work")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(2))
				Expect(itemsListResponse.TotalCount).To(Equal(int32(2)))
			})

			It("should return items matching multiple tags", func() {
				response, err := service.GetItems(ctx, "", "", "family,personal")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))
				Expect(itemsListResponse.TotalCount).To(Equal(int32(1)))
				Expect(itemsListResponse.Items[0].Title).To(Equal("Family Time"))
			})

			It("should return empty list when no tag matches found", func() {
				response, err := service.GetItems(ctx, "", "", "nonexistent")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(BeEmpty())
				Expect(itemsListResponse.TotalCount).To(Equal(int32(0)))
			})
		})

		Context("when searching with combined filters", func() {
			BeforeEach(func() {
				// Create test items
				items := []*models.Item{
					{
						UserID: userID,
						Date:   "2024-01-10",
						Title:  "Work Project Meeting",
						Body:   "Important project discussion",
						Tags:   models.StringList{"work", "project"},
					},
					{
						UserID: userID,
						Date:   "2024-01-11",
						Title:  "Personal Project",
						Body:   "Working on personal coding project",
						Tags:   models.StringList{"personal", "coding"},
					},
				}
				for _, item := range items {
					Expect(storage.PutItem(userID, item)).To(Succeed())
				}
			})

			It("should return items matching both text and tags", func() {
				response, err := service.GetItems(ctx, "", "project", "work")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))
				Expect(itemsListResponse.TotalCount).To(Equal(int32(1)))
				Expect(itemsListResponse.Items[0].Title).To(Equal("Work Project Meeting"))
			})
		})
	})

	Describe("PutItems", func() {
		Context("when no user ID in context", func() {
			It("should return 401 unauthorized", func() {
				emptyCtx := context.Background()
				request := goserver.ItemsRequest{
					Date:  testDate,
					Title: "Test Title",
					Body:  "Test Body",
					Tags:  []string{"tag1", "tag2"},
				}
				response, err := service.PutItems(emptyCtx, request)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(401))
			})
		})

		Context("when creating a new item", func() {
			It("should create and return the item with 200 status", func() {
				request := goserver.ItemsRequest{
					Date:  testDate,
					Title: "New Test Title",
					Body:  "New Test Body",
					Tags:  []string{"new", "test"},
				}

				response, err := service.PutItems(ctx, request)
				Expect(err).ToNot(HaveOccurred())

				assertSuccessfulPutResponse(response, "New Test Title", "New Test Body", []string{"new", "test"}, testDate)
				verifyItemInDatabase(storage, userID, testDate, "New Test Title", "New Test Body", []string{"new", "test"})
			})
		})

		Context("when updating an existing item", func() {
			BeforeEach(func() {
				// Create an initial item
				initialItem := &models.Item{
					UserID: userID,
					Date:   testDate,
					Title:  "Original Title",
					Body:   "Original Body",
					Tags:   models.StringList{"original"},
				}
				Expect(storage.PutItem(userID, initialItem)).To(Succeed())
			})

			It("should update and return the item with 200 status", func() {
				request := goserver.ItemsRequest{
					Date:  testDate,
					Title: "Updated Title",
					Body:  "Updated Body",
					Tags:  []string{"updated", "modified"},
				}

				response, err := service.PutItems(ctx, request)
				Expect(err).ToNot(HaveOccurred())

				assertSuccessfulPutResponse(response, "Updated Title", "Updated Body", []string{"updated", "modified"}, testDate)
				verifyItemInDatabase(storage, userID, testDate, "Updated Title", "Updated Body", []string{"updated", "modified"})
			})
		})

		Context("when saving item with navigation dates", func() {
			BeforeEach(func() {
				// Create previous item
				prevItem := &models.Item{
					UserID: userID,
					Date:   "2024-01-14",
					Title:  "Previous Item",
					Body:   "Previous content",
				}
				Expect(storage.PutItem(userID, prevItem)).To(Succeed())

				// Create next item
				nextItem := &models.Item{
					UserID: userID,
					Date:   "2024-01-16",
					Title:  "Next Item",
					Body:   "Next content",
				}
				Expect(storage.PutItem(userID, nextItem)).To(Succeed())
			})

			It("should include previous and next dates in response", func() {
				request := goserver.ItemsRequest{
					Date:  testDate,
					Title: "Current Item",
					Body:  "Current content",
					Tags:  []string{"current"},
				}

				response, err := service.PutItems(ctx, request)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsResponse, ok := response.Body.(goserver.ItemsResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsResponse.Date).To(Equal(testDate))
				Expect(itemsResponse.PreviousDate).ToNot(BeNil())
				Expect(*itemsResponse.PreviousDate).To(Equal("2024-01-14"))
				Expect(itemsResponse.NextDate).ToNot(BeNil())
				Expect(*itemsResponse.NextDate).To(Equal("2024-01-16"))
			})
		})
	})
})
