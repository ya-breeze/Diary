// Hand-written adapter: StrictServerImpl bridges from the per-service interfaces
// (AuthAPIService, ItemsAPIService, …) to the oapi-codegen StrictServerInterface.
// The adapter translates ImplResponse values into the typed response objects that
// the strict handler machinery expects.
package goserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

// StrictServerImpl implements StrictServerInterface by delegating to individual
// service implementations.
type StrictServerImpl struct {
	assets AssetsAPIService
	auth   AuthAPIService
	health HealthAPIService
	items  ItemsAPIService
	sync   SyncAPIService
	user   UserAPIService
}

// newStrictServerImpl creates a StrictServerImpl from a CustomControllers value.
func newStrictServerImpl(c CustomControllers) *StrictServerImpl {
	return &StrictServerImpl{
		assets: c.AssetsAPIService,
		auth:   c.AuthAPIService,
		health: c.HealthAPIService,
		items:  c.ItemsAPIService,
		sync:   c.SyncAPIService,
		user:   c.UserAPIService,
	}
}

// --- GetAsset ---

func (s *StrictServerImpl) GetAsset(ctx context.Context, req GetAssetRequestObject) (GetAssetResponseObject, error) {
	resp, err := s.assets.GetAsset(ctx, req.Params.Path)
	if err != nil {
		return nil, err
	}
	switch resp.Code {
	case http.StatusOK:
		f, ok := resp.Body.(*os.File)
		if !ok {
			return nil, fmt.Errorf("GetAsset: expected *os.File body")
		}
		info, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("GetAsset: stat: %w", err)
		}
		return GetAsset200AsteriskResponse{
			Body:          f,
			ContentType:   http.DetectContentType(func() []byte { b := make([]byte, 512); n, _ := f.Read(b); f.Seek(0, io.SeekStart); return b[:n] }()),
			ContentLength: info.Size(),
		}, nil
	case http.StatusNotFound:
		return GetAsset404Response{}, nil
	default:
		return nil, fmt.Errorf("GetAsset: unexpected status %d", resp.Code)
	}
}

// --- UploadAssetsBatch ---

func (s *StrictServerImpl) UploadAssetsBatch(_ context.Context, _ UploadAssetsBatchRequestObject) (UploadAssetsBatchResponseObject, error) {
	// Batch upload is handled by the custom AssetsBatchRouter; this path is not reached.
	return UploadAssetsBatch501Response{}, nil
}

// --- Authorize ---

func (s *StrictServerImpl) Authorize(ctx context.Context, req AuthorizeRequestObject) (AuthorizeResponseObject, error) {
	if req.Body == nil {
		return Authorize401Response{}, nil
	}
	resp, err := s.auth.Authorize(ctx, *req.Body)
	if err != nil {
		return nil, err
	}
	switch resp.Code {
	case http.StatusOK:
		body, ok := resp.Body.(Authorize200JSONResponse)
		if !ok {
			return nil, fmt.Errorf("Authorize: unexpected body type %T", resp.Body)
		}
		return body, nil
	case http.StatusUnauthorized:
		return Authorize401Response{}, nil
	default:
		return nil, fmt.Errorf("Authorize: unexpected status %d", resp.Code)
	}
}

// --- FixHealthIssues ---

func (s *StrictServerImpl) FixHealthIssues(ctx context.Context, req FixHealthIssuesRequestObject) (FixHealthIssuesResponseObject, error) {
	if req.Body == nil {
		return FixHealthIssues400Response{}, nil
	}
	resp, err := s.health.FixHealthIssues(ctx, *req.Body)
	if err != nil {
		return nil, err
	}
	switch resp.Code {
	case http.StatusOK:
		body, ok := resp.Body.(HealthIssuesResponse)
		if !ok {
			return nil, fmt.Errorf("FixHealthIssues: unexpected body type %T", resp.Body)
		}
		return FixHealthIssues200JSONResponse(body), nil
	case http.StatusUnauthorized:
		return FixHealthIssues401Response{}, nil
	case http.StatusBadRequest:
		return FixHealthIssues400Response{}, nil
	case http.StatusInternalServerError:
		return FixHealthIssues500Response{}, nil
	default:
		return nil, fmt.Errorf("FixHealthIssues: unexpected status %d", resp.Code)
	}
}

// --- GetHealthIssues ---

func (s *StrictServerImpl) GetHealthIssues(ctx context.Context, _ GetHealthIssuesRequestObject) (GetHealthIssuesResponseObject, error) {
	resp, err := s.health.GetHealthIssues(ctx)
	if err != nil {
		return nil, err
	}
	switch resp.Code {
	case http.StatusOK:
		body, ok := resp.Body.(HealthIssuesResponse)
		if !ok {
			return nil, fmt.Errorf("GetHealthIssues: unexpected body type %T", resp.Body)
		}
		return GetHealthIssues200JSONResponse(body), nil
	case http.StatusUnauthorized:
		return GetHealthIssues401Response{}, nil
	default:
		return nil, fmt.Errorf("GetHealthIssues: unexpected status %d", resp.Code)
	}
}

// --- GetItems ---

func (s *StrictServerImpl) GetItems(ctx context.Context, req GetItemsRequestObject) (GetItemsResponseObject, error) {
	date := ""
	if req.Params.Date != nil {
		date = req.Params.Date.Time.Format("2006-01-02")
	}
	search := ""
	if req.Params.Search != nil {
		search = *req.Params.Search
	}
	tags := ""
	if req.Params.Tags != nil {
		tags = *req.Params.Tags
	}
	resp, err := s.items.GetItems(ctx, date, search, tags)
	if err != nil {
		return nil, err
	}
	switch resp.Code {
	case http.StatusOK:
		body, ok := resp.Body.(ItemsListResponse)
		if !ok {
			return nil, fmt.Errorf("GetItems: unexpected body type %T", resp.Body)
		}
		return GetItems200JSONResponse(body), nil
	case http.StatusBadRequest:
		return GetItems400Response{}, nil
	case http.StatusNotFound:
		return GetItems404Response{}, nil
	default:
		return nil, fmt.Errorf("GetItems: unexpected status %d", resp.Code)
	}
}

// --- PutItems ---

func (s *StrictServerImpl) PutItems(ctx context.Context, req PutItemsRequestObject) (PutItemsResponseObject, error) {
	if req.Body == nil {
		return PutItems400Response{}, nil
	}
	resp, err := s.items.PutItems(ctx, *req.Body)
	if err != nil {
		return nil, err
	}
	switch resp.Code {
	case http.StatusOK:
		body, ok := resp.Body.(ItemsResponse)
		if !ok {
			return nil, fmt.Errorf("PutItems: unexpected body type %T", resp.Body)
		}
		return PutItems200JSONResponse(body), nil
	case http.StatusBadRequest:
		return PutItems400Response{}, nil
	case http.StatusUnauthorized:
		return PutItems401Response{}, nil
	default:
		return nil, fmt.Errorf("PutItems: unexpected status %d", resp.Code)
	}
}

// --- GetChanges ---

func (s *StrictServerImpl) GetChanges(ctx context.Context, req GetChangesRequestObject) (GetChangesResponseObject, error) {
	var since, limit int32
	if req.Params.Since != nil {
		since = *req.Params.Since
	}
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	resp, err := s.sync.GetChanges(ctx, since, limit)
	if err != nil {
		return nil, err
	}
	switch resp.Code {
	case http.StatusOK:
		body, ok := resp.Body.(SyncResponse)
		if !ok {
			return nil, fmt.Errorf("GetChanges: unexpected body type %T", resp.Body)
		}
		return GetChanges200JSONResponse(body), nil
	case http.StatusBadRequest:
		return GetChanges400Response{}, nil
	case http.StatusUnauthorized:
		return GetChanges401Response{}, nil
	default:
		return nil, fmt.Errorf("GetChanges: unexpected status %d", resp.Code)
	}
}

// --- GetUser ---

func (s *StrictServerImpl) GetUser(ctx context.Context, _ GetUserRequestObject) (GetUserResponseObject, error) {
	resp, err := s.user.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	switch resp.Code {
	case http.StatusOK:
		body, ok := resp.Body.(User)
		if !ok {
			return nil, fmt.Errorf("GetUser: unexpected body type %T", resp.Body)
		}
		return GetUser200JSONResponse(body), nil
	default:
		return nil, fmt.Errorf("GetUser: unexpected status %d", resp.Code)
	}
}

// UploadAssetsBatch501Response is a placeholder for the not-implemented batch upload via strict server.
type UploadAssetsBatch501Response struct{}

func (response UploadAssetsBatch501Response) VisitUploadAssetsBatchResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNotImplemented)
	return nil
}

