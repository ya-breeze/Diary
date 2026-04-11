package webapp

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ya-breeze/diary.be/pkg/config"
)

func (r *WebAppRouter) assetsHandler(w http.ResponseWriter, req *http.Request) {
	familyID, code, err := r.GetFamilyIDFromCookie(req)
	if err != nil {
		r.logger.Error("Failed to get family ID from cookie", "error", err)
		http.Error(w, err.Error(), code)
		return
	}

	assetPath := filepath.Join(r.cfg.DataPath, config.AssetsDirName, familyID.String(), strings.TrimPrefix(req.URL.Path, "/web/assets/"))
	r.logger.Info("Serving asset", "path", assetPath)
	http.ServeFile(w, req, assetPath)
}
