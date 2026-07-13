package api

import (
	"context"
	"errors"
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
)

type ItemsAPIServiceImpl struct {
	logger    *slog.Logger
	db        database.Storage
	suggester ai.Suggester
	dataPath  string
}

func NewItemsAPIService(
	logger *slog.Logger, db database.Storage, suggester ai.Suggester, dataPath string,
) goserver.ItemsAPIService {
	return &ItemsAPIServiceImpl{
		logger:    logger,
		db:        db,
		suggester: suggester,
		dataPath:  dataPath,
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
		responseItems[i] = newItemResponse(item)
		s.addNavigationDates(&responseItems[i], familyID, item.Date)
	}

	if date != "" && len(items) == 0 {
		emptyItem := newItemResponse(&models.Item{Date: date})
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

// GetTags - list the family's distinct existing tags
func (s *ItemsAPIServiceImpl) GetTags(ctx context.Context) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Family ID not found in context")
		return goserver.Response(401, nil), nil
	}

	tags, err := s.db.GetDistinctTags(familyID)
	if err != nil {
		s.logger.Error("Failed to get distinct tags", "error", err, "familyID", familyID)
		return goserver.Response(500, nil), nil
	}
	if tags == nil {
		tags = []string{}
	}

	return goserver.Response(200, goserver.TagsResponse{Tags: tags}), nil
}

// GetTagStats - list the family's distinct tags with per-tag usage counts.
func (s *ItemsAPIServiceImpl) GetTagStats(ctx context.Context) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Family ID not found in context")
		return goserver.Response(401, nil), nil
	}

	stats, err := s.db.GetTagStats(familyID)
	if err != nil {
		s.logger.Error("Failed to get tag stats", "error", err, "familyID", familyID)
		return goserver.Response(500, nil), nil
	}

	tags := make([]goserver.TagStat, 0, len(stats))
	for _, st := range stats {
		tags = append(tags, goserver.TagStat{Name: st.Name, Count: st.Count})
	}
	return goserver.Response(200, goserver.TagStatsResponse{Tags: tags}), nil
}

// RenameTag renames a tag across all of the family's entries. A blank new name
// or one equal to the existing name is rejected; collisions merge.
func (s *ItemsAPIServiceImpl) RenameTag(
	ctx context.Context, name string, req goserver.RenameTagRequest,
) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Family ID not found in context")
		return goserver.Response(401, nil), nil
	}

	oldName := strings.TrimSpace(name)
	newName := strings.TrimSpace(req.NewName)
	if oldName == "" || newName == "" || oldName == newName {
		return goserver.Response(400, nil), nil
	}

	if err := s.db.RenameTag(familyID, oldName, newName); err != nil {
		s.logger.Error("Failed to rename tag", "error", err, "familyID", familyID, "old", oldName, "new", newName)
		return goserver.Response(500, nil), nil
	}
	return goserver.Response(200, nil), nil
}

// DeleteTag removes a tag from all of the family's entries.
func (s *ItemsAPIServiceImpl) DeleteTag(ctx context.Context, name string) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Family ID not found in context")
		return goserver.Response(401, nil), nil
	}

	tagName := strings.TrimSpace(name)
	if tagName == "" {
		return goserver.Response(204, nil), nil
	}

	if err := s.db.DeleteTag(familyID, tagName); err != nil {
		s.logger.Error("Failed to delete tag", "error", err, "familyID", familyID, "tag", tagName)
		return goserver.Response(500, nil), nil
	}
	return goserver.Response(204, nil), nil
}

// DismissItemTag removes a single pending suggestion from a day without
// confirming it.
func (s *ItemsAPIServiceImpl) DismissItemTag(
	ctx context.Context, req goserver.DismissTagRequest,
) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Family ID not found in context")
		return goserver.Response(401, nil), nil
	}

	dateStr := req.Date.Time.Format("2006-01-02")
	item, err := s.db.GetItem(familyID, dateStr)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return goserver.Response(404, nil), nil
		}
		s.logger.Error("Dismiss: failed to load item", "error", err, "familyID", familyID, "date", dateStr)
		return goserver.Response(500, nil), nil
	}

	remaining := make([]string, 0, len(item.PendingTags))
	for _, t := range item.PendingTags {
		if !strings.EqualFold(t, req.Tag) {
			remaining = append(remaining, t)
		}
	}
	if err := s.db.SetPendingTags(familyID, dateStr, remaining); err != nil {
		s.logger.Error("Dismiss: failed to update pending tags", "error", err, "familyID", familyID, "date", dateStr)
		return goserver.Response(500, nil), nil
	}

	item.PendingTags = models.StringList(remaining)
	resp := newItemResponse(item)
	s.addNavigationDates(&resp, familyID, item.Date)
	return goserver.Response(200, resp), nil
}

// AcceptItemTag confirms a suggested tag for a day: adds it to confirmed tags
// (additively) and removes it from pending.
func (s *ItemsAPIServiceImpl) AcceptItemTag(
	ctx context.Context, req goserver.DismissTagRequest,
) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Family ID not found in context")
		return goserver.Response(401, nil), nil
	}

	dateStr := req.Date.Time.Format("2006-01-02")
	if err := s.db.AddConfirmedTags(familyID, dateStr, []string{req.Tag}); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return goserver.Response(404, nil), nil
		}
		s.logger.Error("Accept: failed to add tag", "error", err, "familyID", familyID, "date", dateStr)
		return goserver.Response(500, nil), nil
	}

	item, err := s.db.GetItem(familyID, dateStr)
	if err != nil {
		return goserver.Response(500, nil), nil
	}
	resp := newItemResponse(item)
	s.addNavigationDates(&resp, familyID, item.Date)
	return goserver.Response(200, resp), nil
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

	filteredTags := filterTags(itemsRequest.Tags)

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

	response := newItemResponse(item)
	s.addNavigationDates(&response, familyID, item.Date)

	return goserver.Response(200, response), nil
}

// newItemResponse maps a stored item to the API response shape. Tag lists are
// normalized to non-nil slices so they serialize as [] rather than null.
func newItemResponse(item *models.Item) goserver.ItemsResponse {
	tags := nonNil(filterTags((*[]string)(&item.Tags)))
	pendingTags := nonNil(filterTags((*[]string)(&item.PendingTags)))
	body := item.Body
	return goserver.ItemsResponse{
		Date:        parseDate(item.Date),
		Title:       item.Title,
		Body:        &body,
		Tags:        &tags,
		PendingTags: &pendingTags,
	}
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// filterTags trims and drops empty entries from a request's tag list.
func filterTags(in *[]string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, 0, len(*in))
	for _, t := range *in {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
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

	family, ok2 := s.enabledFamily(familyID)
	if !ok2 {
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

	var images []ai.ImageAsset
	if family.AITaggingUseImages {
		images = ai.LoadImageAssets(body, s.dataPath, familyID.String())
	}
	if family.AITaggingUseVideo {
		images = append(images, ai.LoadVideoKeyframes(body, s.dataPath, familyID.String(), s.logger, ai.MaxImages-len(images))...)
	}

	suggestions, err := s.suggester.SuggestTags(ctx, req.Title, body, images, knownTags)
	if err != nil {
		s.logger.Error("Tag suggestion failed", "error", err, "familyID", familyID)
		return goserver.Response(500, nil), nil
	}

	return goserver.Response(200, goserver.SuggestTagsResponse{
		Tags: toAPITagSuggestions(suggestions),
	}), nil
}

// enabledFamily returns the family and true when AI tagging is available for
// it (suggester configured + family opted in). Returns nil, false otherwise.
func (s *ItemsAPIServiceImpl) enabledFamily(familyID uuid.UUID) (*models.Family, bool) {
	if s.suggester == nil || !s.suggester.Enabled() {
		return nil, false
	}
	family, err := s.db.GetFamily(familyID)
	if err != nil {
		s.logger.Error("Failed to load family for AI gate", "error", err, "familyID", familyID)
		return nil, false
	}
	if !family.AITaggingEnabled {
		return nil, false
	}
	return family, true
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
