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
