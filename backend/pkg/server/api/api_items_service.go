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
	threshold float64
}

func NewItemsAPIService(
	logger *slog.Logger, db database.Storage, suggester ai.Suggester, threshold float64,
) goserver.ItemsAPIService {
	return &ItemsAPIServiceImpl{
		logger:    logger,
		db:        db,
		suggester: suggester,
		threshold: threshold,
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
		s.maybeRetag(ctx, familyID, dateStr, item.Title, item.Body, item.Tags)
	}

	response := newItemResponse(item)
	s.addNavigationDates(&response, familyID, item.Date)

	return goserver.Response(200, response), nil
}

// newItemResponse maps a stored item to the API response shape. Tag lists are
// normalized to non-nil slices so they serialize as [] rather than null.
func newItemResponse(item *models.Item) goserver.ItemsResponse {
	tags := nonNil([]string(item.Tags))
	pendingTags := nonNil([]string(item.PendingTags))
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

// maybeRetag asynchronously refreshes an entry's AI tags. It is a no-op unless
// AI tagging is enabled for the family. Under auto mode, confident suggestions
// are applied directly to an untagged day's confirmed tags; otherwise (or for
// low-confidence suggestions) they are staged as pending for the user to accept.
// Errors are logged, not surfaced — retagging is best-effort and must never
// block or fail a save.
func (s *ItemsAPIServiceImpl) maybeRetag(
	ctx context.Context, familyID uuid.UUID, date, title, body string, confirmed models.StringList,
) {
	if s.suggester == nil || !s.suggester.Enabled() {
		return
	}
	family, err := s.db.GetFamily(familyID)
	if err != nil || !family.AITaggingEnabled {
		return
	}
	autoMode := family.AITaggingAuto
	// Detach from the request lifecycle (preserving any context values) so the
	// retag is not cancelled when the HTTP response returns.
	ctx = context.WithoutCancel(ctx)
	go s.runRetag(ctx, familyID, date, title, body, confirmed, autoMode)
}

// runRetag performs the background suggestion + routing for maybeRetag.
func (s *ItemsAPIServiceImpl) runRetag(
	ctx context.Context, familyID uuid.UUID, date, title, body string,
	confirmed models.StringList, autoMode bool,
) {
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
	pending, confident := partitionSuggestions(suggestions, confirmed, s.threshold)

	// Auto mode: apply confident tags directly to an untagged day. Curated days
	// (already have tags) are never auto-written — only staged.
	if autoMode && len(confirmed) == 0 && len(confident) > 0 {
		s.autoApply(familyID, date, confident)
		return
	}
	if err := s.db.SetPendingTags(familyID, date, pending); err != nil {
		s.logger.Error("Retag: failed to store pending tags", "error", err, "familyID", familyID, "date", date)
	}
}

// partitionSuggestions returns suggested names (excluding already-confirmed
// tags) and the subset meeting the confidence threshold.
func partitionSuggestions(
	suggestions []ai.TagSuggestion, confirmed models.StringList, threshold float64,
) ([]string, []string) {
	confirmedSet := make(map[string]struct{}, len(confirmed))
	for _, t := range confirmed {
		confirmedSet[strings.ToLower(t)] = struct{}{}
	}
	var pending, confident []string
	for _, sug := range suggestions {
		if _, ok := confirmedSet[strings.ToLower(sug.Name)]; ok {
			continue
		}
		pending = append(pending, sug.Name)
		if sug.Confidence >= threshold {
			confident = append(confident, sug.Name)
		}
	}
	return pending, confident
}

// autoApply writes confident tags to an untagged day's confirmed tags.
func (s *ItemsAPIServiceImpl) autoApply(familyID uuid.UUID, date string, confident []string) {
	item, err := s.db.GetItem(familyID, date)
	if err != nil {
		s.logger.Error("Retag: failed to load item for auto-apply", "error", err, "familyID", familyID, "date", date)
		return
	}
	item.Tags = models.StringList(confident)
	if err := s.db.PutItem(familyID, item); err != nil {
		s.logger.Error("Retag: failed to auto-apply tags", "error", err, "familyID", familyID, "date", date)
	}
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
