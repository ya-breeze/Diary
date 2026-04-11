package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
	"github.com/ya-breeze/diary.be/pkg/server/tasks"
)

type HealthAPIServiceImpl struct {
	task *tasks.CheckerTask
}

func NewHealthAPIServiceImpl(task *tasks.CheckerTask) goserver.HealthAPIService {
	return &HealthAPIServiceImpl{task: task}
}

func (s *HealthAPIServiceImpl) GetHealthIssues(ctx context.Context) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		return goserver.Response(http.StatusUnauthorized, nil), nil
	}

	result := s.task.GetIssues(familyID)
	return goserver.Response(http.StatusOK, toGoserverResponse(result)), nil
}

func (s *HealthAPIServiceImpl) FixHealthIssues(ctx context.Context, req goserver.HealthFixRequest) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		return goserver.Response(http.StatusUnauthorized, nil), nil
	}

	result, err := s.task.RunFix(familyID, req.Checks)
	if err != nil {
		return goserver.Response(http.StatusInternalServerError, nil), err
	}

	return goserver.Response(http.StatusOK, toGoserverResponse(result)), nil
}

func (s *HealthAPIServiceImpl) DeleteOrphan(ctx context.Context, filename string) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		return goserver.Response(http.StatusUnauthorized, nil), nil
	}

	result, err := s.task.DeleteOrphan(familyID, filename)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return goserver.Response(http.StatusNotFound, nil), nil
		}
		if isValidationError(err) {
			return goserver.Response(http.StatusBadRequest, nil), nil
		}
		return goserver.Response(http.StatusInternalServerError, nil), err
	}

	return goserver.Response(http.StatusOK, toGoserverResponse(result)), nil
}

func (s *HealthAPIServiceImpl) AttachOrphan(ctx context.Context, filename string, req goserver.AttachOrphanRequest) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		return goserver.Response(http.StatusUnauthorized, nil), nil
	}

	result, err := s.task.AttachOrphan(familyID, filename, req.Date)
	if err != nil {
		if isValidationError(err) {
			return goserver.Response(http.StatusBadRequest, nil), nil
		}
		return goserver.Response(http.StatusInternalServerError, nil), err
	}

	return goserver.Response(http.StatusOK, toGoserverResponse(result)), nil
}

func (s *HealthAPIServiceImpl) IgnoreOrphan(ctx context.Context, filename string) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		return goserver.Response(http.StatusUnauthorized, nil), nil
	}

	result, err := s.task.IgnoreOrphan(familyID, filename)
	if err != nil {
		if isValidationError(err) {
			return goserver.Response(http.StatusBadRequest, nil), nil
		}
		return goserver.Response(http.StatusInternalServerError, nil), err
	}

	return goserver.Response(http.StatusOK, toGoserverResponse(result)), nil
}

func (s *HealthAPIServiceImpl) UnignoreOrphan(ctx context.Context, filename string) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		return goserver.Response(http.StatusUnauthorized, nil), nil
	}

	result, err := s.task.UnignoreOrphan(familyID, filename)
	if err != nil {
		if isValidationError(err) {
			return goserver.Response(http.StatusBadRequest, nil), nil
		}
		return goserver.Response(http.StatusInternalServerError, nil), err
	}

	return goserver.Response(http.StatusOK, toGoserverResponse(result)), nil
}

func toGoserverResponse(result *tasks.UserResult) goserver.HealthIssuesResponse {
	resp := goserver.HealthIssuesResponse{Issues: []goserver.HealthIssue{}}
	if result == nil {
		return resp
	}
	t := result.LastChecked
	resp.LastChecked = &t
	for _, i := range result.Issues {
		resp.Issues = append(resp.Issues, goserver.HealthIssue{
			Check:   i.Check,
			Path:    i.Path,
			Message: i.Message,
			Fixable: i.Fixable,
		})
	}
	if len(result.IgnoredOrphans) > 0 {
		ignored := result.IgnoredOrphans
		resp.IgnoredOrphans = &ignored
	}
	return resp
}

// isValidationError reports whether the error is a caller-supplied input validation failure.
func isValidationError(err error) bool {
	return errors.Is(err, tasks.ErrInvalidInput)
}
