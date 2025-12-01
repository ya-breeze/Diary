package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ya-breeze/diary.be/pkg/server/assets"
)

func (r *WebAppRouter) uploadHandler(w http.ResponseWriter, req *http.Request) {
	userID, code, err := r.GetUserIDFromSession(req)
	if err != nil {
		r.logger.Error("Failed to get user ID from session", "error", err)
		http.Error(w, err.Error(), code)
		return
	}

	if err = req.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	asset, header, err := req.FormFile("asset")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer asset.Close()

	// Save the file to the server
	userAssetPath := filepath.Join(r.cfg.AssetPath, userID)
	if err = os.MkdirAll(userAssetPath, 0o755); err != nil {
		r.logger.Error("Failed to create directory", "error", err, "path", userAssetPath)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate extension
	if extErr := assets.ValidateExtension(header.Filename); extErr != nil {
		http.Error(w, extErr.Error(), http.StatusBadRequest)
		return
	}

	// Save atomically using shared util
	name, _, err := assets.SaveFileAtomically(userAssetPath, header, asset, "")
	if err != nil {
		r.logger.Error("Failed to save file", "error", err)
		http.Error(w, "Could not save the file", http.StatusInternalServerError)
		return
	}

	// Respond with the saved file name
	fmt.Fprint(w, name)
}

func (r *WebAppRouter) uploadBatchHandler(w http.ResponseWriter, req *http.Request) {
	userID, code, err := r.GetUserIDFromSession(req)
	if err != nil {
		r.logger.Error("Failed to get user ID from session", "error", err)
		http.Error(w, err.Error(), code)
		return
	}

	limits := assets.ComputeBatchLimits(r.cfg)
	// Note: We don't use EnforceBodySize here because ParseMultipartForm
	// handles size limits internally via its maxMemory parameter.
	// Using MaxBytesReader before ParseMultipartForm can cause "unexpected EOF" errors.
	if err = req.ParseMultipartForm(limits.MaxBatchTotalBytes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	files := req.MultipartForm.File["assets"]
	statusCode, vErr := r.prevalidateFiles(files, limits)
	if vErr != nil {
		http.Error(w, vErr.Error(), statusCode)
		return
	}

	userAssetPath := filepath.Join(r.cfg.AssetPath, userID)
	if err = os.MkdirAll(userAssetPath, 0o755); err != nil {
		r.logger.Error("Failed to create directory", "error", err, "path", userAssetPath)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, code, err := r.processBatch(userAssetPath, files)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		r.logger.Error("failed to encode response", "error", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// prevalidateFiles performs basic checks similar to API handler
func (r *WebAppRouter) prevalidateFiles(files []*multipart.FileHeader, limits assets.BatchLimits) (int, error) {
	if len(files) == 0 {
		return http.StatusBadRequest, errors.New("no files provided")
	}
	if limits.MaxBatchFiles > 0 && len(files) > limits.MaxBatchFiles {
		return http.StatusRequestEntityTooLarge, errors.New("too many files")
	}
	var total int64
	for _, fh := range files {
		if err := validateFileHeaderBasic(fh, limits); err != nil {
			return http.StatusBadRequest, err
		}
		total += fh.Size
	}
	if limits.MaxBatchTotalBytes > 0 && total > limits.MaxBatchTotalBytes {
		return http.StatusRequestEntityTooLarge, errors.New("batch total size exceeded")
	}
	return 0, nil
}

func validateFileHeaderBasic(fh *multipart.FileHeader, limits assets.BatchLimits) error {
	if err := assets.ValidateExtension(fh.Filename); err != nil {
		return err
	}
	if limits.MaxPerFileBytes > 0 && fh.Size > limits.MaxPerFileBytes {
		return errors.New("file too large")
	}
	return nil
}

// processBatch saves files atomically; on any error rolls back
func (r *WebAppRouter) processBatch(
	userAssetPath string,
	files []*multipart.FileHeader,
) (respJSON, int, error) {
	resp := respJSON{Files: make([]string, 0, len(files))}
	createdPaths := make([]string, 0, len(files))
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			rollbackFiles(createdPaths)
			return respJSON{}, http.StatusBadRequest, fmt.Errorf("open: %w", err)
		}
		name, path, err := func() (string, string, error) {
			defer src.Close()
			return assets.SaveFileAtomically(userAssetPath, fh, src, "")
		}()
		if err != nil {
			rollbackFiles(createdPaths)
			return respJSON{}, http.StatusInternalServerError, fmt.Errorf("save: %w", err)
		}
		createdPaths = append(createdPaths, path)
		resp.Files = append(resp.Files, name)
	}
	resp.Count = len(resp.Files)
	return resp, http.StatusOK, nil
}

func rollbackFiles(paths []string) {
	for i := len(paths) - 1; i >= 0; i-- {
		_ = os.Remove(paths[i])
	}
}

type respJSON struct {
	Files []string `json:"files"`
	Count int      `json:"count"`
}
