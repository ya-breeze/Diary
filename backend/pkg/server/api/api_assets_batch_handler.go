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
		r.logger.Error("Batch upload unauthorized - no user ID in context")
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limits := assets.ComputeBatchLimits(r.cfg)
	r.logger.Info("Batch upload request received",
		"userID", userID,
		"contentLength", req.ContentLength,
		"contentType", req.Header.Get("Content-Type"),
		"contentEncoding", req.Header.Get("Content-Encoding"),
		"transferEncoding", req.TransferEncoding,
		"maxBatchTotalBytes", limits.MaxBatchTotalBytes,
	)
	// Note: We don't use EnforceBodySize here because ParseMultipartForm
	// handles size limits internally via its maxMemory parameter.
	// Using MaxBytesReader before ParseMultipartForm can cause "unexpected EOF" errors.

	if err := req.ParseMultipartForm(limits.MaxBatchTotalBytes); err != nil {
		r.logger.Error("Failed to parse multipart form",
			"userID", userID,
			"error", err,
			"maxSize", limits.MaxBatchTotalBytes)
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid multipart form: %v", err))
		return
	}

	files := req.MultipartForm.File["assets"]
	r.logger.Info("Batch upload request", "userID", userID, "files", len(files))
	if code, err := r.prevalidate(files, limits); err != nil {
		r.logger.Error("Batch upload validation failed",
			"userID", userID,
			"fileCount", len(files),
			"statusCode", code,
			"error", err)
		writeJSONError(w, code, err.Error())
		return
	}

	resp, code, err := r.saveAllFiles(userID, files, limits)
	if err != nil {
		r.logger.Error("Failed to save batch files",
			"userID", userID,
			"fileCount", len(files),
			"statusCode", code,
			"error", err)
		writeJSONError(w, code, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		r.logger.Error("failed to encode response", "error", err, "userID", userID)
		writeJSONError(w, http.StatusInternalServerError, "failed to encode response")
	}

	r.logger.Info("Batch upload completed successfully",
		"userID", userID,
		"filesUploaded", resp.Count)
}

func (r *AssetsBatchRouter) prevalidate(files []*multipart.FileHeader, limits assets.BatchLimits) (int, error) {
	if len(files) == 0 {
		r.logger.Warn("Batch upload validation failed: no files provided")
		return http.StatusBadRequest, errors.New("missing assets")
	}
	if limits.MaxBatchFiles > 0 && len(files) > limits.MaxBatchFiles {
		r.logger.Warn("Batch upload validation failed: too many files",
			"fileCount", len(files),
			"maxFiles", limits.MaxBatchFiles)
		return http.StatusRequestEntityTooLarge, errors.New("too many files in batch")
	}
	var totalSize int64
	for i, fh := range files {
		if err := assets.ValidateExtension(fh.Filename); err != nil {
			r.logger.Warn("Batch upload validation failed: invalid file extension",
				"fileIndex", i,
				"filename", fh.Filename,
				"error", err)
			return http.StatusBadRequest, err
		}
		if fh.Size > 0 {
			totalSize += fh.Size
		}
	}
	if limits.MaxBatchTotalBytes > 0 && totalSize > limits.MaxBatchTotalBytes {
		r.logger.Warn("Batch upload validation failed: total size exceeded",
			"totalSize", totalSize,
			"maxSize", limits.MaxBatchTotalBytes)
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

	for i, fh := range files {
		if limits.MaxPerFileBytes > 0 && fh.Size > limits.MaxPerFileBytes {
			r.logger.Error("File too large in batch upload",
				"userID", userID,
				"fileIndex", i,
				"filename", fh.Filename,
				"fileSize", fh.Size,
				"maxSize", limits.MaxPerFileBytes)
			rollback(created)
			return AssetsBatchResponse{}, http.StatusRequestEntityTooLarge, errors.New("file too large")
		}
		src, err := fh.Open()
		if err != nil {
			r.logger.Error("Failed to open file part in batch upload",
				"userID", userID,
				"fileIndex", i,
				"filename", fh.Filename,
				"error", err)
			rollback(created)
			return AssetsBatchResponse{}, http.StatusBadRequest, fmt.Errorf("failed to open part: %w", err)
		}
		name, path, err := func() (string, string, error) {
			defer src.Close()
			return assets.SaveFileAtomically(userAssetPath, fh, src, "")
		}()
		if err != nil {
			r.logger.Error("Failed to save file atomically in batch upload",
				"userID", userID,
				"fileIndex", i,
				"filename", fh.Filename,
				"targetPath", userAssetPath,
				"error", err)
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
