package api

import (
	"context"
	"net/http"

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
	userID, _ := ctx.Value(common.UserIDKey).(string)
	if userID == "" {
		return goserver.Response(http.StatusUnauthorized, nil), nil
	}

	result := s.task.GetIssues(userID)
	return goserver.Response(http.StatusOK, toGoserverResponse(result)), nil
}

func (s *HealthAPIServiceImpl) FixHealthIssues(ctx context.Context, req goserver.HealthFixRequest) (goserver.ImplResponse, error) {
	userID, _ := ctx.Value(common.UserIDKey).(string)
	if userID == "" {
		return goserver.Response(http.StatusUnauthorized, nil), nil
	}

	result, err := s.task.RunFix(userID, req.Checks)
	if err != nil {
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
	return resp
}

