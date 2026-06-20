package api

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

type FamilyAPIServiceImpl struct {
	logger *slog.Logger
	db     database.Storage
}

func NewFamilyAPIService(logger *slog.Logger, db database.Storage) goserver.FamilyAPIService {
	return &FamilyAPIServiceImpl{logger: logger, db: db}
}

func (s *FamilyAPIServiceImpl) GetFamily(ctx context.Context) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		return goserver.Response(500, nil), nil
	}

	family, err := s.db.GetFamily(familyID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return goserver.Response(404, nil), nil
		}
		return goserver.Response(500, nil), nil
	}

	return goserver.Response(200, family.FromDB()), nil
}

func (s *FamilyAPIServiceImpl) UpdateFamilySettings(
	ctx context.Context, req goserver.FamilySettingsRequest,
) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		return goserver.Response(401, nil), nil
	}

	// Load current settings so omitted optional flags keep their existing value.
	current, err := s.db.GetFamily(familyID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return goserver.Response(404, nil), nil
		}
		return goserver.Response(500, nil), nil
	}

	backfill := current.AITaggingBackfill
	if req.AiTaggingBackfill != nil {
		backfill = *req.AiTaggingBackfill
	}
	auto := current.AITaggingAuto
	if req.AiTaggingAuto != nil {
		auto = *req.AiTaggingAuto
	}
	// Auto mode and backfill are meaningless without the master switch on.
	if !req.AiTaggingEnabled {
		backfill = false
		auto = false
	}

	if err = s.db.SetFamilyAISettings(familyID, req.AiTaggingEnabled, backfill, auto); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return goserver.Response(404, nil), nil
		}
		s.logger.Error("Failed to update family settings", "error", err, "familyID", familyID)
		return goserver.Response(500, nil), nil
	}

	family, err := s.db.GetFamily(familyID)
	if err != nil {
		return goserver.Response(500, nil), nil
	}
	return goserver.Response(200, family.FromDB()), nil
}
