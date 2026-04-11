package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
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

	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logAuthError(op, since, limit, start)
		return goserver.Response(401, nil), nil
	}

	s.logSyncRequest(op, familyID, since, limit)

	limit = s.validateLimit(limit)

	changes, err := s.fetchChanges(familyID, since, limit)
	if err != nil {
		s.logSyncError(op, familyID, since, limit, start, err)
		return goserver.Response(500, nil), nil
	}

	response := s.buildSyncResponse(changes, limit)

	s.logSyncSuccess(op, familyID, since, limit, response, start)

	return goserver.Response(200, response), nil
}

func (s *SyncAPIServiceImpl) logAuthError(op string, since, limit int32, start time.Time) {
	s.logger.With(
		"syncOp", op,
		"since", since,
		"limit", limit,
		"duration", time.Since(start),
	).Error("Family ID not found in context")
}

func (s *SyncAPIServiceImpl) logSyncRequest(op string, familyID uuid.UUID, since, limit int32) {
	s.logger.Info("Sync request received",
		"syncOp", op,
		"familyID", familyID,
		"since", since,
		"limit", limit,
	)
}

func (s *SyncAPIServiceImpl) validateLimit(limit int32) int32 {
	if limit <= 0 || limit > 1000 {
		return 100
	}
	return limit
}

func (s *SyncAPIServiceImpl) fetchChanges(familyID uuid.UUID, since, limit int32) ([]*models.ItemChange, error) {
	sinceUint := uint(since)
	if since < 0 {
		sinceUint = 0
	}
	return s.db.GetChangesSince(familyID, sinceUint, int(limit))
}

func (s *SyncAPIServiceImpl) logSyncError(op string, familyID uuid.UUID, since, limit int32, start time.Time, err error) {
	s.logger.Error("Sync operation failed",
		"syncOp", op,
		"familyID", familyID,
		"since", since,
		"limit", limit,
		"status", 500,
		"error", err,
		"duration", time.Since(start),
	)
}

func (s *SyncAPIServiceImpl) buildSyncResponse(changes []*models.ItemChange, limit int32) goserver.SyncResponse {
	responseChanges := make([]goserver.SyncChangeResponse, len(changes))
	for i, change := range changes {
		responseChanges[i] = change.ToSyncResponse()
	}

	hasMore := len(changes) == int(limit)
	var nextID *int32
	if hasMore && len(changes) > 0 {
		lastID := changes[len(changes)-1].ID
		if lastID <= uint(^uint32(0)>>1) {
			id := int32(lastID) // #nosec G115 - checked above
			nextID = &id
		}
	}

	return goserver.SyncResponse{
		Changes: responseChanges,
		HasMore: hasMore,
		NextId:  nextID,
	}
}

func (s *SyncAPIServiceImpl) logSyncSuccess(
	op string,
	familyID uuid.UUID,
	since, limit int32,
	response goserver.SyncResponse,
	start time.Time,
) {
	s.logger.Info("Sync completed",
		"syncOp", op,
		"familyID", familyID,
		"since", since,
		"limit", limit,
		"items", len(response.Changes),
		"hasMore", response.HasMore,
		"nextId", func() int32 {
			if response.NextId != nil {
				return *response.NextId
			}
			return 0
		}(),
		"status", 200,
		"duration", time.Since(start),
	)
}
