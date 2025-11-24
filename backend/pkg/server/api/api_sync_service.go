package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

type SyncAPIServiceImpl struct {
	logger *slog.Logger
	db     database.Storage
}

func NewSyncAPIService(logger *slog.Logger, db database.Storage) goserver.SyncAPIService {
	return &SyncAPIServiceImpl{
		logger: logger,
		db:     db,
	}
}

// GetChanges - get changes for synchronization
func (s *SyncAPIServiceImpl) GetChanges(
	ctx context.Context,
	since int32,
	limit int32,
) (goserver.ImplResponse, error) {
	start := time.Now()
	const op = "changes"

	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value(common.UserIDKey).(string)
	if !ok {
		s.logAuthError(op, since, limit, start)
		return goserver.Response(401, nil), nil
	}

	s.logSyncRequest(op, userID, since, limit)

	// Validate and normalize parameters
	limit = s.validateLimit(limit)

	// Get changes from database
	changes, err := s.fetchChanges(userID, since, limit)
	if err != nil {
		s.logSyncError(op, userID, since, limit, start, err)
		return goserver.Response(500, nil), nil
	}

	// Build response
	response := s.buildSyncResponse(changes, limit)

	s.logSyncSuccess(op, userID, since, limit, response, start)

	return goserver.Response(200, response), nil
}

func (s *SyncAPIServiceImpl) logAuthError(op string, since, limit int32, start time.Time) {
	s.logger.With(
		"syncOp", op,
		"since", since,
		"limit", limit,
		"duration", time.Since(start),
	).Error("User ID not found in context")
}

func (s *SyncAPIServiceImpl) logSyncRequest(op, userID string, since, limit int32) {
	s.logger.Info("Sync request received",
		"syncOp", op,
		"userID", userID,
		"since", since,
		"limit", limit,
	)
}

func (s *SyncAPIServiceImpl) validateLimit(limit int32) int32 {
	if limit <= 0 || limit > 1000 {
		return 100 // default limit
	}
	return limit
}

func (s *SyncAPIServiceImpl) fetchChanges(userID string, since, limit int32) ([]*models.ItemChange, error) {
	sinceUint := uint(since)
	if since < 0 {
		sinceUint = 0
	}
	return s.db.GetChangesSince(userID, sinceUint, int(limit))
}

func (s *SyncAPIServiceImpl) logSyncError(op, userID string, since, limit int32, start time.Time, err error) {
	s.logger.Error("Sync operation failed",
		"syncOp", op,
		"userID", userID,
		"since", since,
		"limit", limit,
		"status", 500,
		"error", err,
		"duration", time.Since(start),
	)
}

func (s *SyncAPIServiceImpl) buildSyncResponse(changes []*models.ItemChange, limit int32) goserver.SyncResponse {
	// Convert database changes to API response format
	responseChanges := make([]goserver.SyncChangeResponse, len(changes))
	for i, change := range changes {
		responseChanges[i] = change.ToSyncResponse()
	}

	// Determine if there are more changes available
	hasMore := len(changes) == int(limit)
	var nextID int32
	if hasMore && len(changes) > 0 {
		lastID := changes[len(changes)-1].ID
		if lastID <= uint(^uint32(0)>>1) { // Check if it fits in int32 (max positive value)
			nextID = int32(lastID) // #nosec G115 - checked above
		}
	}

	return goserver.SyncResponse{
		Changes: responseChanges,
		HasMore: hasMore,
		NextId:  nextID,
	}
}

func (s *SyncAPIServiceImpl) logSyncSuccess(
	op, userID string,
	since, limit int32,
	response goserver.SyncResponse,
	start time.Time,
) {
	s.logger.Info("Sync completed",
		"syncOp", op,
		"userID", userID,
		"since", since,
		"limit", limit,
		"items", len(response.Changes),
		"hasMore", response.HasMore,
		"nextId", response.NextId,
		"status", 200,
		"duration", time.Since(start),
	)
}
