package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
)

type AssetsAPIServiceImpl struct {
	logger *slog.Logger
	cfg    *config.Config
}

func NewAssetsAPIService(logger *slog.Logger, cfg *config.Config) goserver.AssetsAPIService {
	return &AssetsAPIServiceImpl{
		logger: logger,
		cfg:    cfg,
	}
}

// GetAsset - return asset by path
func (s *AssetsAPIServiceImpl) GetAsset(ctx context.Context, path string) (goserver.ImplResponse, error) {
	familyID, ok := common.GetFamilyID(ctx)
	if !ok {
		s.logger.Error("Failed to get family ID from context")
		return goserver.Response(http.StatusUnauthorized, nil), nil
	}

	familyIDStr := familyID.String()

	cleanPath, response := s.validateAndCleanPath(path, familyIDStr)
	if response != nil {
		return *response, nil
	}

	assetPath, response := s.validateAssetPath(cleanPath, familyIDStr)
	if response != nil {
		return *response, nil
	}

	s.logger.Info("Serving asset", "path", assetPath, "familyID", familyIDStr)

	if response := s.validateFileAccess(assetPath, familyIDStr); response != nil {
		return *response, nil
	}

	file, err := os.Open(assetPath)
	if err != nil {
		s.logger.Error("Failed to open asset file", "error", err, "path", assetPath, "familyID", familyIDStr)
		return goserver.Response(http.StatusInternalServerError, nil), nil
	}

	return goserver.Response(http.StatusOK, file), nil
}

// validateAndCleanPath validates the path and returns a cleaned version
func (s *AssetsAPIServiceImpl) validateAndCleanPath(path, familyID string) (string, *goserver.ImplResponse) {
	if strings.Contains(path, "..") {
		s.logger.Warn("Invalid asset path requested (contains ..)", "path", path, "familyID", familyID)
		response := goserver.Response(http.StatusBadRequest, nil)
		return "", &response
	}

	cleanPath := filepath.Clean(path)

	if filepath.IsAbs(cleanPath) {
		s.logger.Warn("Invalid asset path requested (absolute path)", "path", path, "familyID", familyID)
		response := goserver.Response(http.StatusBadRequest, nil)
		return "", &response
	}

	return cleanPath, nil
}

// validateAssetPath constructs and validates the asset path is within family directory
func (s *AssetsAPIServiceImpl) validateAssetPath(cleanPath, familyID string) (string, *goserver.ImplResponse) {
	familyAssetBasePath := filepath.Join(s.cfg.DataPath, config.AssetsDirName, familyID)
	assetPath := filepath.Join(familyAssetBasePath, cleanPath)

	absBasePath, err := filepath.Abs(familyAssetBasePath)
	if err != nil {
		s.logger.Error("Failed to get absolute base path", "error", err, "basePath", familyAssetBasePath)
		response := goserver.Response(http.StatusInternalServerError, nil)
		return "", &response
	}

	absAssetPath, err := filepath.Abs(assetPath)
	if err != nil {
		s.logger.Error("Failed to get absolute asset path", "error", err, "assetPath", assetPath)
		response := goserver.Response(http.StatusInternalServerError, nil)
		return "", &response
	}

	if !strings.HasPrefix(absAssetPath, absBasePath+string(filepath.Separator)) && absAssetPath != absBasePath {
		s.logger.Warn("Asset path outside family directory", "path", cleanPath, "resolvedPath", absAssetPath, "familyID", familyID)
		response := goserver.Response(http.StatusBadRequest, nil)
		return "", &response
	}

	return assetPath, nil
}

// validateFileAccess checks if the file exists and is accessible
func (s *AssetsAPIServiceImpl) validateFileAccess(assetPath, familyID string) *goserver.ImplResponse {
	fileInfo, err := os.Stat(assetPath)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Debug("Asset not found", "path", assetPath, "familyID", familyID)
			response := goserver.Response(http.StatusNotFound, nil)
			return &response
		}
		s.logger.Error("Failed to stat asset file", "error", err, "path", assetPath, "familyID", familyID)
		response := goserver.Response(http.StatusInternalServerError, nil)
		return &response
	}

	if fileInfo.IsDir() {
		s.logger.Warn("Requested path is a directory", "path", assetPath, "familyID", familyID)
		response := goserver.Response(http.StatusBadRequest, nil)
		return &response
	}

	return nil
}

// UploadAssetsBatch - not implemented here; manual router handles /v1/assets/batch.
func (s *AssetsAPIServiceImpl) UploadAssetsBatch(
	ctx context.Context,
	assetsFiles []*os.File,
) (goserver.ImplResponse, error) {
	return goserver.Response(http.StatusNotImplemented, nil), nil
}
