package flows_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/generated/goclient"
)

var _ = Describe("Search Integration Flow", func() {
	var setup *SharedTestSetup

	BeforeEach(func() {
		setup = SetupTestEnvironment()
	})

	AfterEach(func() {
		setup.TeardownTestEnvironment()
	})

	Describe("Search functionality end-to-end", func() {
		BeforeEach(func() {
			// Login and get token (also configures API client headers)
			setup.LoginAndGetToken()

			// Create test data for search
			testItems := []struct {
				date  string
				title string
				body  string
				tags  []string
			}{
				{
					date:  "2024-01-10",
					title: "Vacation Planning",
					body:  "Planning my summer vacation to the beach. Need to book hotel and flights.",
					tags:  []string{"travel", "vacation", "planning"},
				},
				{
					date:  "2024-01-11",
					title: "Work Meeting",
					body:  "Had an important meeting about the new project. Discussed timeline and budget.",
					tags:  []string{"work", "meeting", "project"},
				},
				{
					date:  "2024-01-12",
					title: "Beach Day",
					body:  "Spent the day at the beach with family. Great weather and fun activities.",
					tags:  []string{"family", "beach", "leisure"},
				},
				{
					date:  "2024-01-13",
					title: "Project Review",
					body:  "Reviewed the project progress and made adjustments to the timeline.",
					tags:  []string{"work", "project", "review"},
				},
			}

			// Create all test items
			for _, item := range testItems {
				itemsReq := *goclient.NewItemsRequest(item.date, item.title, item.body)
				itemsReq.SetTags(item.tags)

				_, httpResp, err := setup.APIClient.ItemsAPI.PutItems(context.Background()).ItemsRequest(itemsReq).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
			}
		})

		Context("when searching by text", func() {
			It("should return items matching search text in title", func() {
				// Search for "vacation" in title
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.GetItems(context.Background()).Search("vacation").Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(HaveLen(1))
				Expect(searchResult.TotalCount).To(Equal(int32(1)))
				Expect(searchResult.Items[0].Title).To(Equal("Vacation Planning"))
			})

			It("should return items matching search text in body", func() {
				// Search for "beach" which appears in both title and body
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.GetItems(context.Background()).Search("beach").Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(HaveLen(2))
				Expect(searchResult.TotalCount).To(Equal(int32(2)))

				// Results should be ordered by date descending
				Expect(searchResult.Items[0].Date).To(Equal("2024-01-12")) // Beach Day
				Expect(searchResult.Items[1].Date).To(Equal("2024-01-10")) // Vacation Planning
			})

			It("should return items matching search text case-insensitively", func() {
				// Search for "PROJECT" in uppercase
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.GetItems(context.Background()).Search("PROJECT").Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(HaveLen(2))
				Expect(searchResult.TotalCount).To(Equal(int32(2)))
			})

			It("should return empty results when no matches found", func() {
				// Search for non-existent text
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.
					GetItems(context.Background()).
					Search("nonexistent").
					Execute()

				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(BeEmpty())
				Expect(searchResult.TotalCount).To(Equal(int32(0)))
			})
		})

		Context("when searching by tags", func() {
			It("should return items matching single tag", func() {
				// Search for "work" tag
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.GetItems(context.Background()).Tags("work").Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(HaveLen(2))
				Expect(searchResult.TotalCount).To(Equal(int32(2)))
			})

			It("should return items matching multiple tags", func() {
				// Search for multiple tags
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.
					GetItems(context.Background()).
					Tags("family,leisure").
					Execute()

				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(HaveLen(1))
				Expect(searchResult.TotalCount).To(Equal(int32(1)))
				Expect(searchResult.Items[0].Title).To(Equal("Beach Day"))
			})

			It("should handle tags with spaces correctly", func() {
				// Search for tags with spaces around commas
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.
					GetItems(context.Background()).
					Tags("work, project").
					Execute()

				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(HaveLen(2))
				Expect(searchResult.TotalCount).To(Equal(int32(2)))
			})

			It("should return empty results when no tag matches found", func() {
				// Search for non-existent tag
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.GetItems(context.Background()).Tags("nonexistent").Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(BeEmpty())
				Expect(searchResult.TotalCount).To(Equal(int32(0)))
			})
		})

		Context("when searching with combined filters", func() {
			It("should return items matching both text and tags", func() {
				// Search for "project" text with "work" tag
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.
					GetItems(context.Background()).
					Search("project").
					Tags("work").
					Execute()

				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(HaveLen(2))
				Expect(searchResult.TotalCount).To(Equal(int32(2)))
			})

			It("should return empty results when filters don't match together", func() {
				// Search for "vacation" text with "work" tag (should not match)
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.
					GetItems(context.Background()).
					Search("vacation").
					Tags("work").
					Execute()

				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(BeEmpty())
				Expect(searchResult.TotalCount).To(Equal(int32(0)))
			})
		})

		Context("when using date filter for backward compatibility", func() {
			It("should return specific item when date is provided", func() {
				// Search for specific date
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.GetItems(context.Background()).Date("2024-01-11").Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(HaveLen(1))
				Expect(searchResult.TotalCount).To(Equal(int32(1)))
				Expect(searchResult.Items[0].Title).To(Equal("Work Meeting"))
				Expect(searchResult.Items[0].Date).To(Equal("2024-01-11"))
			})

			It("should return empty item with navigation dates when date has no item", func() {
				// Search for date with no item
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.GetItems(context.Background()).Date("2024-01-20").Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				// Should return one empty item with the requested date
				Expect(searchResult.Items).To(HaveLen(1))
				Expect(searchResult.TotalCount).To(Equal(int32(1)))

				emptyItem := searchResult.Items[0]
				Expect(emptyItem.Date).To(Equal("2024-01-20"))
				Expect(emptyItem.Title).To(Equal(""))
				Expect(emptyItem.Body).To(Equal(""))

				// Navigation dates should be populated based on existing entries
				// In this test, there are entries on 2024-01-10, 2024-01-11, 2024-01-12, 2024-01-13
				// So for 2024-01-20, previousDate should be 2024-01-13 and nextDate should not be set
				Expect(emptyItem.PreviousDate.IsSet()).To(BeTrue())
				prevDate := emptyItem.PreviousDate.Get()
				Expect(*prevDate).To(Equal("2024-01-13"))
				Expect(emptyItem.NextDate.IsSet()).To(BeFalse())
			})
		})

		Context("when no filters are provided", func() {
			It("should return all items", func() {
				// Search without any filters
				searchResult, httpResp, err := setup.APIClient.ItemsAPI.GetItems(context.Background()).Execute()
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

				Expect(searchResult.Items).To(HaveLen(4))
				Expect(searchResult.TotalCount).To(Equal(int32(4)))

				// Results should be ordered by date descending
				Expect(searchResult.Items[0].Date).To(Equal("2024-01-13"))
				Expect(searchResult.Items[1].Date).To(Equal("2024-01-12"))
				Expect(searchResult.Items[2].Date).To(Equal("2024-01-11"))
				Expect(searchResult.Items[3].Date).To(Equal("2024-01-10"))
			})
		})
	})
})
