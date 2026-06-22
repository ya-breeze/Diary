package api_test

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/ya-breeze/diary.be/pkg/ai"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/api"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

func parseTestDate(s string) openapi_types.Date {
	t, _ := time.Parse("2006-01-02", s)
	return openapi_types.Date{Time: t}
}

// Helper function to create context with family ID for items tests
func createContextWithFamilyIDForItems(familyID uuid.UUID) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, common.FamilyIDKey, familyID)
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
	Expect(itemsResponse.Date.Time.Format("2006-01-02")).To(Equal(expectedDate))
	Expect(itemsResponse.Title).To(Equal(expectedTitle))
	Expect(itemsResponse.Body).ToNot(BeNil())
	Expect(*itemsResponse.Body).To(Equal(expectedBody))
	Expect(itemsResponse.Tags).ToNot(BeNil())
	Expect(*itemsResponse.Tags).To(Equal(expectedTags))
}

// Helper function to verify item was saved to database
func verifyItemInDatabase(
	storage database.Storage,
	familyID uuid.UUID, expectedDate, expectedTitle, expectedBody string,
	expectedTags []string,
) {
	savedItem, err := storage.GetItem(familyID, expectedDate)
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
		familyID uuid.UUID
		testDate string
		tempDir  string
	)

	// Create context outside of BeforeEach to avoid fatcontext linting issue
	familyID = uuid.New()
	testDate = "2024-01-15"
	ctx = createContextWithFamilyIDForItems(familyID)

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

		service = api.NewItemsAPIService(logger, storage, ai.NewDisabledSuggester(), 0.8, "")
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
				Expect(itemsListResponse.TotalCount).To(Equal(1))

				// Verify the empty item has the correct date and empty content
				item := itemsListResponse.Items[0]
				Expect(item.Date.Time.Format("2006-01-02")).To(Equal(testDate))
				Expect(item.Title).To(Equal(""))
				Expect(item.Body).ToNot(BeNil())
				Expect(*item.Body).To(Equal(""))
				Expect(item.Tags).ToNot(BeNil())
				Expect(*item.Tags).To(BeEmpty())
			})

			Context("when previous and next items exist", func() {
				BeforeEach(func() {
					// Create previous item
					prevItem := &models.Item{
	
						Date:   "2024-01-14",
						Title:  "Previous Item",
						Body:   "Previous content",
					}
					Expect(storage.PutItem(familyID, prevItem)).To(Succeed())

					// Create next item
					nextItem := &models.Item{
	
						Date:   "2024-01-16",
						Title:  "Next Item",
						Body:   "Next content",
					}
					Expect(storage.PutItem(familyID, nextItem)).To(Succeed())
				})

				It("should include navigation dates even for empty item", func() {
					response, err := service.GetItems(ctx, testDate, "", "")
					Expect(err).ToNot(HaveOccurred())
					Expect(response.Code).To(Equal(200))

					itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
					Expect(ok).To(BeTrue())
					Expect(itemsListResponse.Items).To(HaveLen(1))

					item := itemsListResponse.Items[0]
					Expect(item.Date.Time.Format("2006-01-02")).To(Equal(testDate))
					Expect(item.Title).To(Equal(""))
					Expect(item.Body).ToNot(BeNil())
					Expect(*item.Body).To(Equal(""))
					Expect(item.PreviousDate).ToNot(BeNil())
					Expect(item.PreviousDate.Time.Format("2006-01-02")).To(Equal("2024-01-14"))
					Expect(item.NextDate).ToNot(BeNil())
					Expect(item.NextDate.Time.Format("2006-01-02")).To(Equal("2024-01-16"))
				})
			})
		})

		Context("when item exists (backward compatibility with date filter)", func() {
			BeforeEach(func() {
				// Create a test item
				testItem := &models.Item{

					Date:   testDate,
					Title:  "Test Title",
					Body:   "Test Body Content",
					Tags:   models.StringList{"tag1", "tag2"},
				}
				Expect(storage.PutItem(familyID, testItem)).To(Succeed())
			})

			It("should return the item in list format with 200 status", func() {
				response, err := service.GetItems(ctx, testDate, "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))
				Expect(itemsListResponse.TotalCount).To(Equal(1))

				item := itemsListResponse.Items[0]
				Expect(item.Date.Time.Format("2006-01-02")).To(Equal(testDate))
				Expect(item.Title).To(Equal("Test Title"))
				Expect(item.Body).ToNot(BeNil())
				Expect(*item.Body).To(Equal("Test Body Content"))
				Expect(item.Tags).ToNot(BeNil())
				Expect(*item.Tags).To(Equal([]string{"tag1", "tag2"}))
			})
		})

		Context("when previous and next items exist", func() {
			BeforeEach(func() {
				// Create previous item
				prevItem := &models.Item{

					Date:   "2024-01-14",
					Title:  "Previous Item",
					Body:   "Previous content",
				}
				Expect(storage.PutItem(familyID, prevItem)).To(Succeed())

				// Create current item
				currentItem := &models.Item{

					Date:   testDate,
					Title:  "Current Item",
					Body:   "Current content",
				}
				Expect(storage.PutItem(familyID, currentItem)).To(Succeed())

				// Create next item
				nextItem := &models.Item{

					Date:   "2024-01-16",
					Title:  "Next Item",
					Body:   "Next content",
				}
				Expect(storage.PutItem(familyID, nextItem)).To(Succeed())
			})

			It("should include previous and next dates", func() {
				response, err := service.GetItems(ctx, testDate, "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))

				item := itemsListResponse.Items[0]
				Expect(item.Date.Time.Format("2006-01-02")).To(Equal(testDate))
				Expect(item.PreviousDate).ToNot(BeNil())
				Expect(item.PreviousDate.Time.Format("2006-01-02")).To(Equal("2024-01-14"))
				Expect(item.NextDate).ToNot(BeNil())
				Expect(item.NextDate.Time.Format("2006-01-02")).To(Equal("2024-01-16"))
			})
		})

		Context("when searching by text", func() {
			BeforeEach(func() {
				// Create test items with different content
				items := []*models.Item{
					{
	
						Date:   "2024-01-10",
						Title:  "Vacation Planning",
						Body:   "Planning my summer vacation to the beach",
						Tags:   models.StringList{"travel", "vacation"},
					},
					{
	
						Date:   "2024-01-11",
						Title:  "Work Meeting",
						Body:   "Had an important meeting about the project",
						Tags:   models.StringList{"work", "meeting"},
					},
					{
	
						Date:   "2024-01-12",
						Title:  "Beach Day",
						Body:   "Spent the day at the beach with family",
						Tags:   models.StringList{"family", "beach"},
					},
				}
				for _, item := range items {
					Expect(storage.PutItem(familyID, item)).To(Succeed())
				}
			})

			It("should return items matching search text in title", func() {
				response, err := service.GetItems(ctx, "", "vacation", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))
				Expect(itemsListResponse.TotalCount).To(Equal(1))
				Expect(itemsListResponse.Items[0].Title).To(Equal("Vacation Planning"))
			})

			It("should return items matching search text in body", func() {
				response, err := service.GetItems(ctx, "", "beach", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(2))
				Expect(itemsListResponse.TotalCount).To(Equal(2))
			})

			It("should return empty list when no matches found", func() {
				response, err := service.GetItems(ctx, "", "nonexistent", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(BeEmpty())
				Expect(itemsListResponse.TotalCount).To(Equal(0))
			})
		})

		Context("when searching by tags", func() {
			BeforeEach(func() {
				// Create test items with different tags
				items := []*models.Item{
					{
	
						Date:   "2024-01-10",
						Title:  "Work Project",
						Body:   "Working on the new project",
						Tags:   models.StringList{"work", "project"},
					},
					{
	
						Date:   "2024-01-11",
						Title:  "Family Time",
						Body:   "Spending time with family",
						Tags:   models.StringList{"family", "personal"},
					},
					{
	
						Date:   "2024-01-12",
						Title:  "Work Meeting",
						Body:   "Important work meeting",
						Tags:   models.StringList{"work", "meeting"},
					},
				}
				for _, item := range items {
					Expect(storage.PutItem(familyID, item)).To(Succeed())
				}
			})

			It("should return items matching single tag", func() {
				response, err := service.GetItems(ctx, "", "", "work")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(2))
				Expect(itemsListResponse.TotalCount).To(Equal(2))
			})

			It("should return items matching multiple tags", func() {
				response, err := service.GetItems(ctx, "", "", "family,personal")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))
				Expect(itemsListResponse.TotalCount).To(Equal(1))
				Expect(itemsListResponse.Items[0].Title).To(Equal("Family Time"))
			})

			It("should return empty list when no tag matches found", func() {
				response, err := service.GetItems(ctx, "", "", "nonexistent")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(BeEmpty())
				Expect(itemsListResponse.TotalCount).To(Equal(0))
			})
		})

		Context("when searching with combined filters", func() {
			BeforeEach(func() {
				// Create test items
				items := []*models.Item{
					{
	
						Date:   "2024-01-10",
						Title:  "Work Project Meeting",
						Body:   "Important project discussion",
						Tags:   models.StringList{"work", "project"},
					},
					{
	
						Date:   "2024-01-11",
						Title:  "Personal Project",
						Body:   "Working on personal coding project",
						Tags:   models.StringList{"personal", "coding"},
					},
				}
				for _, item := range items {
					Expect(storage.PutItem(familyID, item)).To(Succeed())
				}
			})

			It("should return items matching both text and tags", func() {
				response, err := service.GetItems(ctx, "", "project", "work")
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsListResponse.Items).To(HaveLen(1))
				Expect(itemsListResponse.TotalCount).To(Equal(1))
				Expect(itemsListResponse.Items[0].Title).To(Equal("Work Project Meeting"))
			})
		})
	})

	Describe("PutItems", func() {
		Context("when no user ID in context", func() {
			It("should return 401 unauthorized", func() {
				emptyCtx := context.Background()
				body := "Test Body"
				tags := []string{"tag1", "tag2"}
				request := goserver.ItemsRequest{
					Date:  parseTestDate(testDate),
					Title: "Test Title",
					Body:  &body,
					Tags:  &tags,
				}
				response, err := service.PutItems(emptyCtx, request)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(401))
			})
		})

		Context("when creating a new item", func() {
			It("should create and return the item with 200 status", func() {
				body := "New Test Body"
				tags := []string{"new", "test"}
				request := goserver.ItemsRequest{
					Date:  parseTestDate(testDate),
					Title: "New Test Title",
					Body:  &body,
					Tags:  &tags,
				}

				response, err := service.PutItems(ctx, request)
				Expect(err).ToNot(HaveOccurred())

				assertSuccessfulPutResponse(response, "New Test Title", "New Test Body", []string{"new", "test"}, testDate)
				verifyItemInDatabase(storage, familyID, testDate, "New Test Title", "New Test Body", []string{"new", "test"})
			})
		})

		Context("when updating an existing item", func() {
			BeforeEach(func() {
				// Create an initial item
				initialItem := &models.Item{

					Date:   testDate,
					Title:  "Original Title",
					Body:   "Original Body",
					Tags:   models.StringList{"original"},
				}
				Expect(storage.PutItem(familyID, initialItem)).To(Succeed())
			})

			It("should update and return the item with 200 status", func() {
				body := "Updated Body"
				tags := []string{"updated", "modified"}
				request := goserver.ItemsRequest{
					Date:  parseTestDate(testDate),
					Title: "Updated Title",
					Body:  &body,
					Tags:  &tags,
				}

				response, err := service.PutItems(ctx, request)
				Expect(err).ToNot(HaveOccurred())

				assertSuccessfulPutResponse(response, "Updated Title", "Updated Body", []string{"updated", "modified"}, testDate)
				verifyItemInDatabase(storage, familyID, testDate, "Updated Title", "Updated Body", []string{"updated", "modified"})
			})
		})

		Context("when saving item with navigation dates", func() {
			BeforeEach(func() {
				// Create previous item
				prevItem := &models.Item{

					Date:   "2024-01-14",
					Title:  "Previous Item",
					Body:   "Previous content",
				}
				Expect(storage.PutItem(familyID, prevItem)).To(Succeed())

				// Create next item
				nextItem := &models.Item{

					Date:   "2024-01-16",
					Title:  "Next Item",
					Body:   "Next content",
				}
				Expect(storage.PutItem(familyID, nextItem)).To(Succeed())
			})

			It("should include previous and next dates in response", func() {
				body := "Current content"
				tags := []string{"current"}
				request := goserver.ItemsRequest{
					Date:  parseTestDate(testDate),
					Title: "Current Item",
					Body:  &body,
					Tags:  &tags,
				}

				response, err := service.PutItems(ctx, request)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Code).To(Equal(200))

				itemsResponse, ok := response.Body.(goserver.ItemsResponse)
				Expect(ok).To(BeTrue())
				Expect(itemsResponse.Date.Time.Format("2006-01-02")).To(Equal(testDate))
				Expect(itemsResponse.PreviousDate).ToNot(BeNil())
				Expect(itemsResponse.PreviousDate.Time.Format("2006-01-02")).To(Equal("2024-01-14"))
				Expect(itemsResponse.NextDate).ToNot(BeNil())
				Expect(itemsResponse.NextDate.Time.Format("2006-01-02")).To(Equal("2024-01-16"))
			})
		})
	})
})

// fakeSuggester is an enabled suggester returning canned suggestions, for
// exercising the SuggestItemTags wiring without calling Gemini.
type fakeSuggester struct {
	suggestions []ai.TagSuggestion
}

func (f fakeSuggester) Enabled() bool { return true }

func (f fakeSuggester) SuggestTags(
	_ context.Context, _, _ string, _ []ai.ImageAsset, _ []string,
) ([]ai.TagSuggestion, error) {
	return f.suggestions, nil
}

func ptr(s string) *string { return &s }

var _ = Describe("ItemsAPIService SuggestItemTags", func() {
	var (
		logger   *slog.Logger
		storage  database.Storage
		tempDir  string
		familyID uuid.UUID
		ctx      context.Context
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		var err error
		tempDir, err = os.MkdirTemp("", "suggest_test")
		Expect(err).NotTo(HaveOccurred())
		storage = database.NewStorage(logger, &config.Config{DataPath: tempDir})
		Expect(storage.Open()).To(Succeed())

		fam, err := storage.CreateFamily("suggest-fam")
		Expect(err).NotTo(HaveOccurred())
		familyID = fam.ID
		ctx = createContextWithFamilyIDForItems(familyID)
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	req := goserver.SuggestTagsRequest{
		Date:  parseTestDate("2024-01-15"),
		Title: "Beach day",
		Body:  ptr("We swam all afternoon"),
	}

	It("returns 503 when the suggester is disabled", func() {
		svc := api.NewItemsAPIService(logger, storage, ai.NewDisabledSuggester(), 0.8, "")
		resp, err := svc.SuggestItemTags(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(503))
	})

	It("returns 503 when the family has not enabled AI tagging", func() {
		svc := api.NewItemsAPIService(logger, storage,
			fakeSuggester{suggestions: []ai.TagSuggestion{{Name: "beach", Confidence: 0.9}}}, 0.8, "")
		resp, err := svc.SuggestItemTags(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(503))
	})

	It("returns suggestions when enabled and opted in", func() {
		Expect(storage.SetFamilyAITaggingEnabled(familyID, true)).To(Succeed())
		svc := api.NewItemsAPIService(logger, storage, fakeSuggester{suggestions: []ai.TagSuggestion{
			{Name: "beach", Confidence: 0.9},
			{Name: "summer", Confidence: 0.5},
		}}, 0.8, "")
		resp, err := svc.SuggestItemTags(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(200))
		body, ok := resp.Body.(goserver.SuggestTagsResponse)
		Expect(ok).To(BeTrue())
		Expect(body.Tags).To(HaveLen(2))
		Expect(body.Tags[0].Name).To(Equal("beach"))
	})

	It("returns 401 without a family in context", func() {
		svc := api.NewItemsAPIService(logger, storage, ai.NewDisabledSuggester(), 0.8, "")
		resp, err := svc.SuggestItemTags(context.Background(), req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(401))
	})
})

var _ = Describe("ItemsAPIService GetTags", func() {
	var (
		logger   *slog.Logger
		storage  database.Storage
		service  goserver.ItemsAPIService
		tempDir  string
		familyID uuid.UUID
		ctx      context.Context
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		var err error
		tempDir, err = os.MkdirTemp("", "gettags_test")
		Expect(err).NotTo(HaveOccurred())
		storage = database.NewStorage(logger, &config.Config{DataPath: tempDir})
		Expect(storage.Open()).To(Succeed())
		fam, err := storage.CreateFamily("tags-fam")
		Expect(err).NotTo(HaveOccurred())
		familyID = fam.ID
		ctx = createContextWithFamilyIDForItems(familyID)
		service = api.NewItemsAPIService(logger, storage, ai.NewDisabledSuggester(), 0.8, "")
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	It("returns 401 without a family in context", func() {
		resp, err := service.GetTags(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(401))
	})

	It("returns an empty list when there are no tags", func() {
		resp, err := service.GetTags(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(200))
		body, ok := resp.Body.(goserver.TagsResponse)
		Expect(ok).To(BeTrue())
		Expect(body.Tags).To(BeEmpty())
	})

	It("returns deduplicated, sorted tags", func() {
		for _, it := range []*models.Item{
			{Date: "2024-02-01", Title: "x", Tags: models.StringList{"travel", "family"}},
			{Date: "2024-02-02", Title: "y", Tags: models.StringList{"family", "work"}},
		} {
			Expect(storage.PutItem(familyID, it)).To(Succeed())
		}
		resp, err := service.GetTags(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(200))
		body := resp.Body.(goserver.TagsResponse)
		Expect(body.Tags).To(Equal([]string{"family", "travel", "work"}))
	})
})

var _ = Describe("ItemsAPIService tag management", func() {
	var (
		logger   *slog.Logger
		storage  database.Storage
		service  goserver.ItemsAPIService
		tempDir  string
		familyID uuid.UUID
		ctx      context.Context
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		var err error
		tempDir, err = os.MkdirTemp("", "tagmgmt_test")
		Expect(err).NotTo(HaveOccurred())
		storage = database.NewStorage(logger, &config.Config{DataPath: tempDir})
		Expect(storage.Open()).To(Succeed())
		fam, err := storage.CreateFamily("tagmgmt-fam")
		Expect(err).NotTo(HaveOccurred())
		familyID = fam.ID
		ctx = createContextWithFamilyIDForItems(familyID)
		service = api.NewItemsAPIService(logger, storage, ai.NewDisabledSuggester(), 0.8, "")
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	Describe("GetTagStats", func() {
		It("returns 401 without a family in context", func() {
			resp, err := service.GetTagStats(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(401))
		})

		It("returns tags with counts sorted by count desc then name", func() {
			for _, it := range []*models.Item{
				{Date: "2024-03-01", Title: "a", Tags: models.StringList{"family", "travel"}},
				{Date: "2024-03-02", Title: "b", Tags: models.StringList{"family", "work"}},
				{Date: "2024-03-03", Title: "c", Tags: models.StringList{"family"}},
			} {
				Expect(storage.PutItem(familyID, it)).To(Succeed())
			}
			resp, err := service.GetTagStats(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			body := resp.Body.(goserver.TagStatsResponse)
			Expect(body.Tags).To(Equal([]goserver.TagStat{
				{Name: "family", Count: 3},
				{Name: "travel", Count: 1},
				{Name: "work", Count: 1},
			}))
		})

		It("returns an empty list when there are no tags", func() {
			resp, err := service.GetTagStats(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			body := resp.Body.(goserver.TagStatsResponse)
			Expect(body.Tags).To(BeEmpty())
		})
	})

	Describe("RenameTag", func() {
		It("returns 401 without a family in context", func() {
			resp, err := service.RenameTag(context.Background(), "a", goserver.RenameTagRequest{NewName: "b"})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(401))
		})

		It("rejects a blank new name with 400", func() {
			resp, err := service.RenameTag(ctx, "old", goserver.RenameTagRequest{NewName: "  "})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(400))
		})

		It("rejects an unchanged name with 400", func() {
			resp, err := service.RenameTag(ctx, "old", goserver.RenameTagRequest{NewName: "old"})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(400))
		})

		It("renames the tag across all entries", func() {
			Expect(storage.PutItem(familyID, &models.Item{
				Date: "2024-03-01", Title: "a", Tags: models.StringList{"vacaiton", "work"},
			})).To(Succeed())
			resp, err := service.RenameTag(ctx, "vacaiton", goserver.RenameTagRequest{NewName: "vacation"})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			saved, err := storage.GetItem(familyID, "2024-03-01")
			Expect(err).NotTo(HaveOccurred())
			Expect(saved.Tags).To(Equal(models.StringList{"vacation", "work"}))
		})
	})

	Describe("DeleteTag", func() {
		It("returns 401 without a family in context", func() {
			resp, err := service.DeleteTag(context.Background(), "a")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(401))
		})

		It("removes the tag from all entries and returns 204", func() {
			Expect(storage.PutItem(familyID, &models.Item{
				Date: "2024-03-01", Title: "a", Tags: models.StringList{"misc", "work"},
			})).To(Succeed())
			resp, err := service.DeleteTag(ctx, "misc")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(204))
			saved, err := storage.GetItem(familyID, "2024-03-01")
			Expect(err).NotTo(HaveOccurred())
			Expect(saved.Tags).To(Equal(models.StringList{"work"}))
		})
	})
})

var _ = Describe("ItemsAPIService DismissItemTag", func() {
	var (
		logger   *slog.Logger
		storage  database.Storage
		service  goserver.ItemsAPIService
		tempDir  string
		familyID uuid.UUID
		ctx      context.Context
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		var err error
		tempDir, err = os.MkdirTemp("", "dismiss_test")
		Expect(err).NotTo(HaveOccurred())
		storage = database.NewStorage(logger, &config.Config{DataPath: tempDir})
		Expect(storage.Open()).To(Succeed())
		fam, err := storage.CreateFamily("dismiss-fam")
		Expect(err).NotTo(HaveOccurred())
		familyID = fam.ID
		ctx = createContextWithFamilyIDForItems(familyID)
		service = api.NewItemsAPIService(logger, storage, ai.NewDisabledSuggester(), 0.8, "")
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	It("removes one pending tag and keeps the rest", func() {
		Expect(storage.PutItem(familyID, &models.Item{Date: "2024-01-01", Title: "trip"})).To(Succeed())
		Expect(storage.SetPendingTags(familyID, "2024-01-01", []string{"hiking", "mountains"})).To(Succeed())

		resp, err := service.DismissItemTag(ctx, goserver.DismissTagRequest{
			Date: parseTestDate("2024-01-01"), Tag: "hiking",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(200))
		body := resp.Body.(goserver.ItemsResponse)
		Expect(*body.PendingTags).To(Equal([]string{"mountains"}))

		item, _ := storage.GetItem(familyID, "2024-01-01")
		Expect([]string(item.PendingTags)).To(Equal([]string{"mountains"}))
		Expect(item.Tags).To(BeEmpty()) // not confirmed
	})

	It("returns 404 for a missing entry", func() {
		resp, err := service.DismissItemTag(ctx, goserver.DismissTagRequest{
			Date: parseTestDate("2099-01-01"), Tag: "x",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(404))
	})

	It("returns 401 without a family in context", func() {
		resp, err := service.DismissItemTag(context.Background(), goserver.DismissTagRequest{
			Date: parseTestDate("2024-01-01"), Tag: "x",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(401))
	})
})

var _ = Describe("ItemsAPIService AcceptItemTag", func() {
	var (
		logger   *slog.Logger
		storage  database.Storage
		service  goserver.ItemsAPIService
		tempDir  string
		familyID uuid.UUID
		ctx      context.Context
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		var err error
		tempDir, err = os.MkdirTemp("", "accept_test")
		Expect(err).NotTo(HaveOccurred())
		storage = database.NewStorage(logger, &config.Config{DataPath: tempDir})
		Expect(storage.Open()).To(Succeed())
		fam, err := storage.CreateFamily("accept-fam")
		Expect(err).NotTo(HaveOccurred())
		familyID = fam.ID
		ctx = createContextWithFamilyIDForItems(familyID)
		service = api.NewItemsAPIService(logger, storage, ai.NewDisabledSuggester(), 0.8, "")
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	It("moves a pending tag into confirmed (additive) and removes it from pending", func() {
		Expect(storage.PutItem(familyID, &models.Item{
			Date: "2024-01-01", Title: "trip", Tags: models.StringList{"existing"},
		})).To(Succeed())
		Expect(storage.SetPendingTags(familyID, "2024-01-01", []string{"hiking", "mountains"})).To(Succeed())

		resp, err := service.AcceptItemTag(ctx, goserver.DismissTagRequest{
			Date: parseTestDate("2024-01-01"), Tag: "hiking",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(200))
		body := resp.Body.(goserver.ItemsResponse)
		Expect(*body.Tags).To(ConsistOf("existing", "hiking")) // additive, existing kept
		Expect(*body.PendingTags).To(Equal([]string{"mountains"}))

		item, _ := storage.GetItem(familyID, "2024-01-01")
		Expect([]string(item.Tags)).To(ConsistOf("existing", "hiking"))
		Expect([]string(item.PendingTags)).To(Equal([]string{"mountains"}))
	})

	It("returns 404 for a missing entry", func() {
		resp, err := service.AcceptItemTag(ctx, goserver.DismissTagRequest{
			Date: parseTestDate("2099-01-01"), Tag: "x",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(404))
	})

	It("returns 401 without a family in context", func() {
		resp, err := service.AcceptItemTag(context.Background(), goserver.DismissTagRequest{
			Date: parseTestDate("2024-01-01"), Tag: "x",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(401))
	})
})

// Blank-tag filtering: items carrying legacy blank tags must never expose them
// through the API, even when scrubBlankTags hasn't run yet.
var _ = Describe("ItemsAPIService blank-tag filtering", func() {
	var (
		logger   *slog.Logger
		storage  database.Storage
		service  goserver.ItemsAPIService
		tempDir  string
		familyID uuid.UUID
		ctx      context.Context
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		var err error
		tempDir, err = os.MkdirTemp("", "blanktag_test")
		Expect(err).NotTo(HaveOccurred())
		storage = database.NewStorage(logger, &config.Config{DataPath: tempDir})
		Expect(storage.Open()).To(Succeed())
		fam, err := storage.CreateFamily("blank-tag-fam")
		Expect(err).NotTo(HaveOccurred())
		familyID = fam.ID
		ctx = createContextWithFamilyIDForItems(familyID)
		service = api.NewItemsAPIService(logger, storage, ai.NewDisabledSuggester(), 0.8, "")
	})

	AfterEach(func() {
		storage.Close()
		os.RemoveAll(tempDir)
	})

	It("strips blank tags from GetItems response without affecting valid tags", func() {
		// Write a clean item then corrupt its tags via raw SQL to simulate legacy
		// data that predates the write-path filter.
		Expect(storage.PutItem(familyID, &models.Item{Date: "2024-03-01", Title: "day"})).To(Succeed())
		Expect(storage.GetDB().Exec(
			`UPDATE items SET tags = '["real","","  "]', pending_tags = '[""," beach"]'
			 WHERE date = ? AND family_id = ?`, "2024-03-01", familyID,
		).Error).To(Succeed())

		resp, err := service.GetItems(ctx, "2024-03-01", "", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(200))
		body := resp.Body.(goserver.ItemsListResponse)
		Expect(body.Items).To(HaveLen(1))
		item := body.Items[0]
		Expect(*item.Tags).To(Equal([]string{"real"}))
		Expect(*item.PendingTags).To(Equal([]string{"beach"}))
	})

	It("returns non-nil empty slices when all stored tags are blank", func() {
		Expect(storage.PutItem(familyID, &models.Item{Date: "2024-03-02", Title: "day2"})).To(Succeed())
		Expect(storage.GetDB().Exec(
			`UPDATE items SET tags = '[""]', pending_tags = '["  "]'
			 WHERE date = ? AND family_id = ?`, "2024-03-02", familyID,
		).Error).To(Succeed())

		resp, err := service.GetItems(ctx, "2024-03-02", "", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.Code).To(Equal(200))
		body := resp.Body.(goserver.ItemsListResponse)
		item := body.Items[0]
		Expect(item.Tags).NotTo(BeNil())
		Expect(*item.Tags).To(BeEmpty())
		Expect(item.PendingTags).NotTo(BeNil())
		Expect(*item.PendingTags).To(BeEmpty())
	})
})
