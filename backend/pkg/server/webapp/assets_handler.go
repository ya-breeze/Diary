package webapp

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ya-breeze/diary.be/pkg/config"
)

func (r *WebAppRouter) assetsHandler(w http.ResponseWriter, req *http.Request) {
	userID, code, err := r.GetUserIDFromSession(req)
	if err != nil {
		r.logger.Error("Failed to get user ID from session", "error", err)
		http.Error(w, err.Error(), code)
		return
	}

	userAsset := filepath.Join(r.cfg.DataPath, config.AssetsDirName, userID, strings.TrimPrefix(req.URL.Path, "/web/assets/"))
	r.logger.Info("Serving asset", "path", userAsset)
	http.ServeFile(w, req, userAsset)
}
