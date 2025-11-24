package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/assets"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

type AssetsBatchResponse struct {
	Files []AssetsBatchFile `json:"files"`
	Count int               `json:"count"`
}

type AssetsBatchFile struct {
	OriginalName string `json:"originalName"`
	SavedName    string `json:"savedName"`
	Size         int64  `json:"size"`
	ContentType  string `json:"contentType"`
}

type AssetsBatchRouter struct {
	logger *slog.Logger
	cfg    *config.Config
}

func NewAssetsBatchRouter(logger *slog.Logger, cfg *config.Config) *AssetsBatchRouter {
	return &AssetsBatchRouter{logger: logger, cfg: cfg}
}

// Implement goserver.Router
func (r *AssetsBatchRouter) Routes() goserver.Routes {
	return goserver.Routes{
		"uploadAssetsBatch": {Method: http.MethodPost, Pattern: "/v1/assets/batch", HandlerFunc: r.handleBatch},
	}
}

func (r *AssetsBatchRouter) handleBatch(w http.ResponseWriter, req *http.Request) {
	userID, _ := req.Context().Value(common.UserIDKey).(string)
	if userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limits := assets.ComputeBatchLimits(r.cfg)
	assets.EnforceBodySize(w, req, limits.MaxBatchTotalBytes)

	if err := req.ParseMultipartForm(limits.MaxBatchTotalBytes); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid multipart form: %v", err))
		return
	}

	files := req.MultipartForm.File["assets"]
	r.logger.Info("Batch upload request", "userID", userID, "files", len(files))
	if code, err := r.prevalidate(files, limits); err != nil {
		writeJSONError(w, code, err.Error())
		return
	}

	resp, code, err := r.saveAllFiles(userID, files, limits)
	if err != nil {
		writeJSONError(w, code, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		r.logger.Error("failed to encode response", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to encode response")
	}
}

func (r *AssetsBatchRouter) prevalidate(files []*multipart.FileHeader, limits assets.BatchLimits) (int, error) {
	if len(files) == 0 {
		return http.StatusBadRequest, errors.New("missing assets")
	}
	if limits.MaxBatchFiles > 0 && len(files) > limits.MaxBatchFiles {
		return http.StatusRequestEntityTooLarge, errors.New("too many files in batch")
	}
	var totalSize int64
	for _, fh := range files {
		if err := assets.ValidateExtension(fh.Filename); err != nil {
			return http.StatusBadRequest, err
		}
		if fh.Size > 0 {
			totalSize += fh.Size
		}
	}
	if limits.MaxBatchTotalBytes > 0 && totalSize > limits.MaxBatchTotalBytes {
		return http.StatusRequestEntityTooLarge, errors.New("batch total size exceeded")
	}
	return 0, nil
}

func (r *AssetsBatchRouter) saveAllFiles(
	userID string,
	files []*multipart.FileHeader,
	limits assets.BatchLimits,
) (
	AssetsBatchResponse,
	int,
	error,
) {
	userAssetPath := filepath.Join(r.cfg.AssetPath, userID)
	created := make([]string, 0, len(files))
	resp := AssetsBatchResponse{Files: make([]AssetsBatchFile, 0, len(files))}

	for _, fh := range files {
		if limits.MaxPerFileBytes > 0 && fh.Size > limits.MaxPerFileBytes {
			rollback(created)
			return AssetsBatchResponse{}, http.StatusRequestEntityTooLarge, errors.New("file too large")
		}
		src, err := fh.Open()
		if err != nil {
			rollback(created)
			return AssetsBatchResponse{}, http.StatusBadRequest, fmt.Errorf("failed to open part: %w", err)
		}
		name, path, err := func() (string, string, error) {
			defer src.Close()
			return assets.SaveFileAtomically(userAssetPath, fh, src, "")
		}()
		if err != nil {
			rollback(created)
			return AssetsBatchResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to save file: %w", err)
		}
		created = append(created, path)
		resp.Files = append(resp.Files, AssetsBatchFile{
			OriginalName: fh.Filename,
			SavedName:    name,
			Size:         fh.Size,
			ContentType:  contentType(fh),
		})
	}

	resp.Count = len(resp.Files)
	return resp, http.StatusOK, nil
}

func (r *AssetsBatchRouter) Use(router *mux.Router) {
	for name, route := range r.Routes() {
		router.Methods(route.Method).Path(route.Pattern).Name(name).Handler(route.HandlerFunc)
	}
}

func rollback(paths []string) {
	for i := len(paths) - 1; i >= 0; i-- {
		_ = os.Remove(paths[i])
	}
}

func contentType(fh *multipart.FileHeader) string {
	if ct := fh.Header.Get("Content-Type"); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// writeJSONError writes a minimal JSON error response {"error":"..."}
func writeJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		// we cannot write another response at this point; best effort log via fmt
		//nolint:forbidigo
		fmt.Println("failed to write JSON error:", err)
	}
}
