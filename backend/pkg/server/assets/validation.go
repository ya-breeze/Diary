package assets

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/ya-breeze/diary.be/pkg/config"
)

// AllowedExtensions is the unified list of file extensions allowed for upload.
//
//nolint:gochecknoglobals
var AllowedExtensions = []string{
	".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".mp4", ".mov", ".avi", ".wmv", ".flv", ".mkv",
}

// ValidateExtension checks the filename extension against allowed list.
func ValidateExtension(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !slices.Contains(AllowedExtensions, ext) {
		return fmt.Errorf("invalid file type: %s", ext)
	}
	return nil
}

// EnforceBodySize wraps the request body with a MaxBytesReader to limit payload size.
// If maxBytes <= 0, no limit is enforced.
func EnforceBodySize(w http.ResponseWriter, r *http.Request, maxBytes int64) {
	if maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	}
}

// SaveFileAtomically saves the uploaded part to the destination directory with a generated UUID filename
// preserving original extension. It writes to a temporary file then renames to final path.
func SaveFileAtomically(
	dstDir string,
	header *multipart.FileHeader,
	src multipart.File,
	prefix string,
) (string, string, error) {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return "", "", err
	}

	finalName := uuid.New().String() + strings.ToLower(filepath.Ext(header.Filename))
	tmpPath := filepath.Join(dstDir, ".tmp_"+finalName)
	finalPath := filepath.Join(dstDir, finalName)

	f, err := os.Create(tmpPath)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = f.Close() }()

	if _, err = io.Copy(f, src); err != nil {
		_ = os.Remove(tmpPath)
		return "", "", err
	}
	// ensure data is flushed
	if err = f.Sync(); err != nil {
		_ = os.Remove(tmpPath)
		return "", "", err
	}

	if err = os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", "", err
	}

	return finalName, finalPath, nil
}

// BatchLimits contains computed absolute byte limits for enforcement.
type BatchLimits struct {
	MaxPerFileBytes    int64
	MaxBatchFiles      int
	MaxBatchTotalBytes int64
}

// ComputeBatchLimits converts MB config to byte limits.
func ComputeBatchLimits(cfg *config.Config) BatchLimits {
	return BatchLimits{
		MaxPerFileBytes:    int64(cfg.MaxPerFileSizeMB) * 1024 * 1024,
		MaxBatchFiles:      cfg.MaxBatchFiles,
		MaxBatchTotalBytes: int64(cfg.MaxBatchTotalSizeMB) * 1024 * 1024,
	}
}
