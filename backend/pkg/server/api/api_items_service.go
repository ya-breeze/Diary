package api

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
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
	dataPath  string

	// inflight coalesces background retags: at most one per (family,date) runs
	// at a time, so rapid saves don't spawn piles of concurrent model calls.
	retagMu       sync.Mutex
	retagInflight map[string]struct{}
}

func NewItemsAPIService(
	logger *slog.Logger, db database.Storage, suggester ai.Suggester, threshold float64, dataPath string,
) goserver.ItemsAPIService {
	return &ItemsAPIServiceImpl{
		logger:        logger,
		db:            db,
		suggester:     suggester,
		threshold:     threshold,
		dataPath:      dataPath,
		retagInflight: make(map[string]struct{}),
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

// maybeRetag asynchronously refreshes an entry's AI tags. It is a no-op unless
// AI tagging is enabled for the family. Under auto mode, confident suggestions
// are applied directly to an untagged day's confirmed tags; otherwise (or for
// low-confidence suggestions) they are staged as pending for the user to accept.
// Errors are logged, not surfaced — retagging is best-effort and must never
// block or fail a save.
func (s *ItemsAPIServiceImpl) maybeRetag(
	ctx context.Context, familyID uuid.UUID, date, title, body string, confirmed models.StringList,
) {
	family, ok := s.enabledFamily(familyID)
	if !ok {
		return
	}
	autoMode := family.AITaggingAuto
	useImages := family.AITaggingUseImages

	// Coalesce: skip if a retag for this exact day is already running.
	key := familyID.String() + "|" + date
	s.retagMu.Lock()
	if _, running := s.retagInflight[key]; running {
		s.retagMu.Unlock()
		return
	}
	s.retagInflight[key] = struct{}{}
	s.retagMu.Unlock()

	// Detach from the request lifecycle (preserving any context values) so the
	// retag is not cancelled when the HTTP response returns.
	ctx = context.WithoutCancel(ctx)
	go func() {
		defer func() {
			s.retagMu.Lock()
			delete(s.retagInflight, key)
			s.retagMu.Unlock()
		}()
		s.runRetag(ctx, familyID, date, title, body, confirmed, autoMode, useImages)
	}()
}

// runRetag performs the background suggestion + routing for maybeRetag.
func (s *ItemsAPIServiceImpl) runRetag(
	ctx context.Context, familyID uuid.UUID, date, title, body string,
	confirmed models.StringList, autoMode bool, useImages bool,
) {
	knownTags, err := s.db.GetDistinctTags(familyID)
	if err != nil {
		s.logger.Error("Retag: failed to load known tags", "error", err, "familyID", familyID)
		return
	}
	var images []ai.ImageAsset
	if useImages {
		images = ai.LoadImageAssets(body, s.dataPath, familyID.String())
	}
	suggestions, err := s.suggester.SuggestTags(ctx, title, body, images, knownTags)
	if err != nil {
		s.logger.Error("Retag: suggestion failed", "error", err, "familyID", familyID, "date", date)
		return
	}
	pending, confident := ai.Partition(suggestions, confirmed, s.threshold)

	// Auto mode: apply confident tags directly to an untagged day. Curated days
	// (already have tags) are never auto-written — only staged. AddConfirmedTags
	// is additive and atomic, so a concurrent edit can't lose tags.
	if autoMode && len(confirmed) == 0 && len(confident) > 0 {
		if err := s.db.AddConfirmedTags(familyID, date, confident); err != nil {
			s.logger.Error("Retag: failed to auto-apply tags", "error", err, "familyID", familyID, "date", date)
		}
		// Stage any low-confidence suggestions that weren't auto-applied.
		if uncertain := subtractStrings(pending, confident); len(uncertain) > 0 {
			if err := s.db.SetPendingTags(familyID, date, uncertain); err != nil {
				s.logger.Error("Retag: failed to store pending tags", "error", err, "familyID", familyID, "date", date)
			}
		}
		return
	}
	if err := s.db.SetPendingTags(familyID, date, pending); err != nil {
		s.logger.Error("Retag: failed to store pending tags", "error", err, "familyID", familyID, "date", date)
	}
}

// subtractStrings returns the elements of all that are not in exclude.
func subtractStrings(all, exclude []string) []string {
	if len(exclude) == 0 {
		return all
	}
	ex := make(map[string]struct{}, len(exclude))
	for _, s := range exclude {
		ex[s] = struct{}{}
	}
	var out []string
	for _, s := range all {
		if _, ok := ex[s]; !ok {
			out = append(out, s)
		}
	}
	return out
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
