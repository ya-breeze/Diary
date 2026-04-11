package api

import (
	"context"
	"log/slog"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/google/uuid"
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
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Family ID not found in context")
		return goserver.Response(401, nil), nil
	}

	s.logger.Info("Getting items", "familyID", familyID, "date", date, "search", search, "tags", tags)

	searchParams := database.SearchParams{
		Date:       date,
		SearchText: search,
	}

	if tags != "" {
		searchParams.Tags = strings.Split(tags, ",")
		for i, tag := range searchParams.Tags {
			searchParams.Tags[i] = strings.TrimSpace(tag)
		}
	}

	items, totalCount, err := s.db.GetItems(familyID, searchParams)
	if err != nil {
		s.logger.Error("Failed to get items", "error", err, "familyID", familyID, "searchParams", searchParams)
		return goserver.Response(500, nil), nil
	}

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
		s.addNavigationDates(&responseItems[i], familyID, item.Date)
	}

	if date != "" && len(items) == 0 {
		emptyTags := []string{}
		emptyBody := ""
		emptyItem := goserver.ItemsResponse{
			Date:  parseDate(date),
			Title: "",
			Body:  &emptyBody,
			Tags:  &emptyTags,
		}
		s.addNavigationDates(&emptyItem, familyID, date)
		responseItems = []goserver.ItemsResponse{emptyItem}
		totalCount = 1
	}

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
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Family ID not found in context")
		return goserver.Response(401, nil), nil
	}

	dateStr := itemsRequest.Date.Time.Format("2006-01-02")
	s.logger.Info("Saving item", "familyID", familyID, "date", dateStr)

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

	item := &models.Item{
		Date:  dateStr,
		Title: itemsRequest.Title,
		Body:  body,
		Tags:  models.StringList(filteredTags),
	}

	if err := s.db.PutItem(familyID, item); err != nil {
		s.logger.Error("Failed to save item", "error", err, "item", item)
		return goserver.Response(500, nil), nil
	}

	savedTags := []string(item.Tags)
	savedBody := item.Body
	response := goserver.ItemsResponse{
		Date:  parseDate(item.Date),
		Title: item.Title,
		Body:  &savedBody,
		Tags:  &savedTags,
	}

	s.addNavigationDates(&response, familyID, item.Date)

	return goserver.Response(200, response), nil
}

// addNavigationDates adds previous and next dates to the response
func (s *ItemsAPIServiceImpl) addNavigationDates(response *goserver.ItemsResponse, familyID uuid.UUID, date string) {
	if previousDate, err := s.db.GetPreviousDate(familyID, date); err == nil {
		d := openapi_types.Date{Time: parseDate(previousDate).Time}
		response.PreviousDate = &d
	}
	if nextDate, err := s.db.GetNextDate(familyID, date); err == nil {
		d := openapi_types.Date{Time: parseDate(nextDate).Time}
		response.NextDate = &d
	}
}

// parseDate parses a "2006-01-02" date string into an openapi_types.Date.
func parseDate(s string) openapi_types.Date {
	t, _ := time.Parse("2006-01-02", s)
	return openapi_types.Date{Time: t}
}
