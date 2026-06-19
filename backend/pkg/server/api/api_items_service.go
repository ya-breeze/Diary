package api

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/ya-breeze/diary.be/pkg/ai"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

type ItemsAPIServiceImpl struct {
	logger    *slog.Logger
	db        database.Storage
	suggester ai.Suggester
}

func NewItemsAPIService(
	logger *slog.Logger, db database.Storage, suggester ai.Suggester,
) goserver.ItemsAPIService {
	return &ItemsAPIServiceImpl{
		logger:    logger,
		db:        db,
		suggester: suggester,
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
		pendingTags := []string(item.PendingTags)
		body := item.Body
		responseItems[i] = goserver.ItemsResponse{
			Date:        parseDate(item.Date),
			Title:       item.Title,
			Body:        &body,
			Tags:        &tags,
			PendingTags: &pendingTags,
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

	// Detect whether the tag-relevant content changed, to decide on retagging.
	newHash := utils.ComputeTagsSourceHash(itemsRequest.Title, body)
	contentChanged := true
	if existing, err := s.db.GetItem(familyID, dateStr); err == nil {
		contentChanged = existing.TagsSourceHash != newHash
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

	// Edit-triggered retag: when content changed and AI tagging is enabled, refresh
	// the day's pending suggestions in the background (suggest-only in phase 1).
	if contentChanged {
		s.maybeRetag(familyID, dateStr, item.Title, item.Body, item.Tags)
	}

	savedTags := []string(item.Tags)
	savedPendingTags := []string(item.PendingTags)
	savedBody := item.Body
	response := goserver.ItemsResponse{
		Date:        parseDate(item.Date),
		Title:       item.Title,
		Body:        &savedBody,
		Tags:        &savedTags,
		PendingTags: &savedPendingTags,
	}

	s.addNavigationDates(&response, familyID, item.Date)

	return goserver.Response(200, response), nil
}

// SuggestItemTags - suggest tags for draft entry content without saving.
func (s *ItemsAPIServiceImpl) SuggestItemTags(
	ctx context.Context,
	req goserver.SuggestTagsRequest,
) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Family ID not found in context")
		return goserver.Response(401, nil), nil
	}

	if !s.aiTaggingEnabled(familyID) {
		return goserver.Response(503, nil), nil
	}

	body := ""
	if req.Body != nil {
		body = *req.Body
	}

	knownTags, err := s.db.GetDistinctTags(familyID)
	if err != nil {
		s.logger.Error("Failed to load known tags", "error", err, "familyID", familyID)
		return goserver.Response(500, nil), nil
	}

	suggestions, err := s.suggester.SuggestTags(ctx, req.Title, body, knownTags)
	if err != nil {
		s.logger.Error("Tag suggestion failed", "error", err, "familyID", familyID)
		return goserver.Response(500, nil), nil
	}

	return goserver.Response(200, goserver.SuggestTagsResponse{
		Tags: toAPITagSuggestions(suggestions),
	}), nil
}

// aiTaggingEnabled reports whether AI tagging may run for the family: the server
// must have an API key (suggester enabled) AND the family must have opted in.
func (s *ItemsAPIServiceImpl) aiTaggingEnabled(familyID uuid.UUID) bool {
	if s.suggester == nil || !s.suggester.Enabled() {
		return false
	}
	family, err := s.db.GetFamily(familyID)
	if err != nil {
		s.logger.Error("Failed to load family for AI gate", "error", err, "familyID", familyID)
		return false
	}
	return family.AITaggingEnabled
}

// maybeRetag asynchronously refreshes an entry's pending suggestions. It is a
// no-op unless AI tagging is enabled for the family. Errors are logged, not
// surfaced — retagging is best-effort and must never block or fail a save.
func (s *ItemsAPIServiceImpl) maybeRetag(
	familyID uuid.UUID, date, title, body string, confirmed models.StringList,
) {
	if !s.aiTaggingEnabled(familyID) {
		return
	}
	go func() {
		ctx := context.Background()
		knownTags, err := s.db.GetDistinctTags(familyID)
		if err != nil {
			s.logger.Error("Retag: failed to load known tags", "error", err, "familyID", familyID)
			return
		}
		suggestions, err := s.suggester.SuggestTags(ctx, title, body, knownTags)
		if err != nil {
			s.logger.Error("Retag: suggestion failed", "error", err, "familyID", familyID, "date", date)
			return
		}
		// Exclude anything already confirmed; storage prunes again as defense-in-depth.
		confirmedSet := make(map[string]struct{}, len(confirmed))
		for _, t := range confirmed {
			confirmedSet[strings.ToLower(t)] = struct{}{}
		}
		pending := make([]string, 0, len(suggestions))
		for _, sug := range suggestions {
			if _, ok := confirmedSet[strings.ToLower(sug.Name)]; ok {
				continue
			}
			pending = append(pending, sug.Name)
		}
		if err := s.db.SetPendingTags(familyID, date, pending); err != nil {
			s.logger.Error("Retag: failed to store pending tags", "error", err, "familyID", familyID, "date", date)
		}
	}()
}

// toAPITagSuggestions maps internal suggestions to the API response type.
func toAPITagSuggestions(in []ai.TagSuggestion) []goserver.TagSuggestion {
	out := make([]goserver.TagSuggestion, len(in))
	for i, s := range in {
		out[i] = goserver.TagSuggestion{Name: s.Name, Confidence: s.Confidence}
	}
	return out
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
