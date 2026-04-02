package api

import (
	"context"
	"log/slog"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

type ItemsAPIServiceImpl struct {
	logger *slog.Logger
	db     database.Storage
}

func NewItemsAPIService(logger *slog.Logger, db database.Storage) goserver.ItemsAPIService {
	return &ItemsAPIServiceImpl{
		logger: logger,
		db:     db,
	}
}

// GetItems - get diary items
func (s *ItemsAPIServiceImpl) GetItems(
	ctx context.Context,
	date string,
	search string,
	tags string,
) (goserver.ImplResponse, error) {
	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value(common.UserIDKey).(string)
	if !ok {
		s.logger.Error("User ID not found in context")
		return goserver.Response(401, nil), nil
	}

	s.logger.Info("Getting items", "userID", userID, "date", date, "search", search, "tags", tags)

	// Parse search parameters
	searchParams := database.SearchParams{
		Date:       date,
		SearchText: search,
	}

	// Parse tags parameter (comma-separated)
	if tags != "" {
		searchParams.Tags = strings.Split(tags, ",")
		// Trim whitespace from each tag
		for i, tag := range searchParams.Tags {
			searchParams.Tags[i] = strings.TrimSpace(tag)
		}
	}

	// Get items using the new search method
	items, totalCount, err := s.db.GetItems(userID, searchParams)
	if err != nil {
		s.logger.Error("Failed to get items", "error", err, "userID", userID, "searchParams", searchParams)
		return goserver.Response(500, nil), nil
	}

	// Convert database items to API response items
	responseItems := make([]goserver.ItemsResponse, len(items))
	for i, item := range items {
		tags := []string(item.Tags)
		body := item.Body
		responseItems[i] = goserver.ItemsResponse{
			Date:  parseDate(item.Date),
			Title: item.Title,
			Body:  &body,
			Tags:  &tags,
		}
		// Add navigation dates for each item
		s.addNavigationDates(&responseItems[i], userID, item.Date)
	}

	// Special case: when filtering by a specific date and no items found,
	// return an empty item with navigation dates to enable Previous/Next buttons
	if date != "" && len(items) == 0 {
		emptyTags := []string{}
		emptyBody := ""
		emptyItem := goserver.ItemsResponse{
			Date:  parseDate(date),
			Title: "",
			Body:  &emptyBody,
			Tags:  &emptyTags,
		}
		s.addNavigationDates(&emptyItem, userID, date)
		responseItems = []goserver.ItemsResponse{emptyItem}
		totalCount = 1
	}

	// Create the list response
	response := goserver.ItemsListResponse{
		Items:      responseItems,
		TotalCount: totalCount,
	}

	return goserver.Response(200, response), nil
}

// PutItems - upsert diary item
func (s *ItemsAPIServiceImpl) PutItems(
	ctx context.Context,
	itemsRequest goserver.ItemsRequest,
) (goserver.ImplResponse, error) {
	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value(common.UserIDKey).(string)
	if !ok {
		s.logger.Error("User ID not found in context")
		return goserver.Response(401, nil), nil
	}

	dateStr := itemsRequest.Date.Time.Format("2006-01-02")
	s.logger.Info("Saving item", "userID", userID, "date", dateStr)

	// Filter tags: trim spaces and skip empty values
	var filteredTags []string
	if itemsRequest.Tags != nil {
		filteredTags = make([]string, 0, len(*itemsRequest.Tags))
		for _, t := range *itemsRequest.Tags {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			filteredTags = append(filteredTags, t)
		}
	}

	body := ""
	if itemsRequest.Body != nil {
		body = *itemsRequest.Body
	}

	// Convert request to database model
	item := &models.Item{
		UserID: userID,
		Date:   dateStr,
		Title:  itemsRequest.Title,
		Body:   body,
		Tags:   models.StringList(filteredTags),
	}

	// Save the item to database
	if err := s.db.PutItem(userID, item); err != nil {
		s.logger.Error("Failed to save item", "error", err, "item", item)
		return goserver.Response(500, nil), nil
	}

	// Return the saved item as response
	savedTags := []string(item.Tags)
	savedBody := item.Body
	response := goserver.ItemsResponse{
		Date:  parseDate(item.Date),
		Title: item.Title,
		Body:  &savedBody,
		Tags:  &savedTags,
	}

	// Add navigation dates
	s.addNavigationDates(&response, userID, item.Date)

	return goserver.Response(200, response), nil
}

// addNavigationDates adds previous and next dates to the response
func (s *ItemsAPIServiceImpl) addNavigationDates(response *goserver.ItemsResponse, userID, date string) {
	if previousDate, err := s.db.GetPreviousDate(userID, date); err == nil {
		d := openapi_types.Date{Time: parseDate(previousDate).Time}
		response.PreviousDate = &d
	}
	if nextDate, err := s.db.GetNextDate(userID, date); err == nil {
		d := openapi_types.Date{Time: parseDate(nextDate).Time}
		response.NextDate = &d
	}
}

// parseDate parses a "2006-01-02" date string into an openapi_types.Date.
func parseDate(s string) openapi_types.Date {
	t, _ := time.Parse("2006-01-02", s)
	return openapi_types.Date{Time: t}
}
