package ai

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

var supportedImageMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// maxImages caps the number of images sent to Gemini per request.
// Gemini imposes an inline-blob size limit; this prevents OOM on entries
// with many large images and keeps requests well under that limit.
const maxImages = 10

func LoadImageAssets(body, dataPath, familyID string) []ImageAsset {
	names := utils.GetAssetsFromMarkdown(body)
	if len(names) == 0 {
		return nil
	}

	assetsDir := filepath.Join(dataPath, config.AssetsDirName, familyID)
	seen := make(map[string]struct{}, len(names))
	var assets []ImageAsset
	for _, name := range names {
		if len(assets) >= maxImages {
			break
		}
		clean := filepath.Clean(name)
		if strings.Contains(clean, "..") {
			continue
		}
		if _, dup := seen[clean]; dup {
			continue
		}
		seen[clean] = struct{}{}

		path := filepath.Join(assetsDir, clean)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		mime := http.DetectContentType(data)
		if i := strings.IndexByte(mime, ';'); i >= 0 {
			mime = strings.TrimSpace(mime[:i])
		}
		if !supportedImageMIMEs[mime] {
			continue
		}
		assets = append(assets, ImageAsset{MIMEType: mime, Data: data})
	}
	return assets
}
