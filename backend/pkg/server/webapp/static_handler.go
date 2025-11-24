package webapp

import (
	"net/http"
	"path/filepath"
	"strings"
)

func (r *WebAppRouter) staticHandler(w http.ResponseWriter, req *http.Request) {
	// Extract the file path from the URL
	staticPath := strings.TrimPrefix(req.URL.Path, "/web/static/")

	// Construct the full file path
	filePath := filepath.Join("webapp", "static", staticPath)

	// Security check: ensure the path doesn't escape the static directory
	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, "webapp/static/") {
		r.logger.Warn("Attempted path traversal", "path", staticPath, "cleanPath", cleanPath)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	r.logger.Info("Serving static file", "path", cleanPath)

	// Set appropriate content type based on file extension
	ext := filepath.Ext(cleanPath)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	// Serve the file
	http.ServeFile(w, req, cleanPath)
}
