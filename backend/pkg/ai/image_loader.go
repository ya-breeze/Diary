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

// MaxImages caps the total number of image/frame assets sent to Gemini per
// request (images + video keyframes combined). Gemini imposes an inline-blob
// size limit; keeping this value modest prevents OOM and keeps costs predictable.
const MaxImages = 10

func LoadImageAssets(body, dataPath, familyID string) []ImageAsset {
	names := utils.GetAssetsFromMarkdown(body)
	if len(names) == 0 {
		return nil
	}

	assetsDir := filepath.Join(dataPath, config.AssetsDirName, familyID)
	seen := make(map[string]struct{}, len(names))
	var assets []ImageAsset
	for _, name := range names {
		if len(assets) >= MaxImages {
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
