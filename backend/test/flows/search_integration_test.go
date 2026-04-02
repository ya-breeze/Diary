package flows_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
			setup.LoginAndGetToken()

			testItems := []struct {
				date, title, body string
				tags              []string
			}{
				{"2024-01-10", "Vacation Planning", "Planning my summer vacation to the beach. Need to book hotel and flights.", []string{"travel", "vacation", "planning"}},
				{"2024-01-11", "Work Meeting", "Had an important meeting about the new project. Discussed timeline and budget.", []string{"work", "meeting", "project"}},
				{"2024-01-12", "Beach Day", "Spent the day at the beach with family. Great weather and fun activities.", []string{"family", "beach", "leisure"}},
				{"2024-01-13", "Project Review", "Reviewed the project progress and made adjustments to the timeline.", []string{"work", "project", "review"}},
			}

			for _, item := range testItems {
				_, httpResp, err := setup.APIClient.PutItems(context.Background(), item.date, item.title, item.body, item.tags)
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
			}
		})

		Context("when searching by text", func() {
			It("should return items matching search text in title", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "vacation", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(1))
				Expect(result.TotalCount).To(Equal(1))
				Expect(result.Items[0].Title).To(Equal("Vacation Planning"))
			})

			It("should return items matching search text in body", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "beach", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(2))
				Expect(result.TotalCount).To(Equal(2))
				Expect(result.Items[0].Date).To(Equal("2024-01-12"))
				Expect(result.Items[1].Date).To(Equal("2024-01-10"))
			})

			It("should return items matching search text case-insensitively", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "PROJECT", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(2))
				Expect(result.TotalCount).To(Equal(2))
			})

			It("should return empty results when no matches found", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "nonexistent", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(BeEmpty())
				Expect(result.TotalCount).To(Equal(0))
			})
		})

		Context("when searching by tags", func() {
			It("should return items matching single tag", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "", "work")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(2))
				Expect(result.TotalCount).To(Equal(2))
			})

			It("should return items matching multiple tags", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "", "family,leisure")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(1))
				Expect(result.TotalCount).To(Equal(1))
				Expect(result.Items[0].Title).To(Equal("Beach Day"))
			})

			It("should handle tags with spaces correctly", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "", "work, project")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(2))
				Expect(result.TotalCount).To(Equal(2))
			})

			It("should return empty results when no tag matches found", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "", "nonexistent")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(BeEmpty())
				Expect(result.TotalCount).To(Equal(0))
			})
		})

		Context("when searching with combined filters", func() {
			It("should return items matching both text and tags", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "project", "work")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(2))
				Expect(result.TotalCount).To(Equal(2))
			})

			It("should return empty results when filters don't match together", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "vacation", "work")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(BeEmpty())
				Expect(result.TotalCount).To(Equal(0))
			})
		})

		Context("when using date filter", func() {
			It("should return specific item when date is provided", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "2024-01-11", "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(1))
				Expect(result.TotalCount).To(Equal(1))
				Expect(result.Items[0].Title).To(Equal("Work Meeting"))
				Expect(result.Items[0].Date).To(Equal("2024-01-11"))
			})

			It("should return empty item with navigation dates when date has no item", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "2024-01-20", "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(1))
				Expect(result.TotalCount).To(Equal(1))

				emptyItem := result.Items[0]
				Expect(emptyItem.Date).To(Equal("2024-01-20"))
				Expect(emptyItem.Title).To(Equal(""))
				Expect(emptyItem.Body).To(Equal(""))
				Expect(emptyItem.PreviousDate).ToNot(BeNil())
				Expect(*emptyItem.PreviousDate).To(Equal("2024-01-13"))
				Expect(emptyItem.NextDate).To(BeNil())
			})
		})

		Context("when no filters are provided", func() {
			It("should return all items", func() {
				result, httpResp, err := setup.APIClient.GetItems(context.Background(), "", "", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
				Expect(result.Items).To(HaveLen(4))
				Expect(result.TotalCount).To(Equal(4))
				Expect(result.Items[0].Date).To(Equal("2024-01-13"))
				Expect(result.Items[1].Date).To(Equal("2024-01-12"))
				Expect(result.Items[2].Date).To(Equal("2024-01-11"))
				Expect(result.Items[3].Date).To(Equal("2024-01-10"))
			})
		})
	})
})
