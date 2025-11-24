package api

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

type UserAPIServiceImpl struct {
	logger *slog.Logger
	db     database.Storage
}

func NewUserAPIService(logger *slog.Logger, db database.Storage) goserver.UserAPIService {
	return &UserAPIServiceImpl{
		logger: logger,
		db:     db,
	}
}

// GetUser - return user object
func (s *UserAPIServiceImpl) GetUser(ctx context.Context) (goserver.ImplResponse, error) {
	userID, ok := ctx.Value(common.UserIDKey).(string)
	if !ok {
		return goserver.Response(500, nil), nil
	}

	user, err := s.db.GetUser(userID)
	if err != nil && errors.Is(err, database.ErrNotFound) {
		return goserver.Response(500, nil), nil
	}
	if user == nil {
		return goserver.Response(404, nil), nil
	}

	return goserver.Response(200, user.FromDB()), nil
}
