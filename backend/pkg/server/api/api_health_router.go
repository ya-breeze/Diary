package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
	"github.com/ya-breeze/diary.be/pkg/server/tasks"
)

// HealthIssueDTO is the JSON-serialisable shape of a health issue for the API.
type HealthIssueDTO struct {
	Check   string `json:"check"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Fixable bool   `json:"fixable"`
}

// HealthIssuesResponse is the response body for GET /v1/health/issues.
type HealthIssuesResponse struct {
	Issues      []HealthIssueDTO `json:"issues"`
	LastChecked *time.Time       `json:"lastChecked,omitempty"`
}

// HealthFixRequest is the request body for POST /v1/health/fix.
type HealthFixRequest struct {
	Checks []string `json:"checks"`
}

// HealthRouter serves the two health-check endpoints.
type HealthRouter struct {
	logger *slog.Logger
	task   *tasks.CheckerTask
}

func NewHealthRouter(logger *slog.Logger, task *tasks.CheckerTask) *HealthRouter {
	return &HealthRouter{logger: logger, task: task}
}

func (r *HealthRouter) Routes() goserver.Routes {
	return goserver.Routes{
		"getHealthIssues": {Method: http.MethodGet, Pattern: "/v1/health/issues", HandlerFunc: r.handleGet},
		"fixHealthIssues": {Method: http.MethodPost, Pattern: "/v1/health/fix", HandlerFunc: r.handleFix},
	}
}

func (r *HealthRouter) Use(router *mux.Router) {
	for name, route := range r.Routes() {
		router.Methods(route.Method).Path(route.Pattern).Name(name).Handler(route.HandlerFunc)
	}
}

func (r *HealthRouter) handleGet(w http.ResponseWriter, req *http.Request) {
	userID, _ := req.Context().Value(common.UserIDKey).(string)
	if userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result := r.task.GetIssues(userID)
	resp := toResponse(result)
	writeJSON(w, http.StatusOK, resp)
}

func (r *HealthRouter) handleFix(w http.ResponseWriter, req *http.Request) {
	userID, _ := req.Context().Value(common.UserIDKey).(string)
	if userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body HealthFixRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := r.task.RunFix(userID, body.Checks)
	if err != nil {
		r.logger.Error("Health fix failed", "userID", userID, "error", err)
		writeJSONError(w, http.StatusInternalServerError, "fix failed")
		return
	}

	writeJSON(w, http.StatusOK, toResponse(result))
}

func toResponse(result *tasks.UserResult) HealthIssuesResponse {
	resp := HealthIssuesResponse{Issues: []HealthIssueDTO{}}
	if result == nil {
		return resp
	}
	t := result.LastChecked
	resp.LastChecked = &t
	for _, i := range result.Issues {
		resp.Issues = append(resp.Issues, HealthIssueDTO{
			Check:   i.Check,
			Path:    i.Path,
			Message: i.Message,
			Fixable: i.Fixable,
		})
	}
	return resp
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
