package app

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type frontendHandler struct {
	distDir string
	index   string
}

func newFrontendHandler(distDir string) *frontendHandler {
	if distDir == "" {
		return nil
	}

	indexPath := filepath.Join(distDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return nil
	}

	return &frontendHandler{
		distDir: distDir,
		index:   indexPath,
	}
}

func newUploadsHandler(uploadsDir string) http.Handler {
	if strings.TrimSpace(uploadsDir) == "" {
		return nil
	}

	return http.StripPrefix("/api/uploads/", http.FileServer(http.Dir(uploadsDir)))
}

func (h *frontendHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	if isBackendRoute(r.URL.Path) {
		http.NotFound(w, r)
		return
	}

	requestPath := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if requestPath == ".." || strings.HasPrefix(requestPath, "../") {
		http.NotFound(w, r)
		return
	}
	candidate := filepath.Join(h.distDir, requestPath)

	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		http.ServeFile(w, r, candidate)
		return
	}

	http.ServeFile(w, r, h.index)
}

func isBackendRoute(requestPath string) bool {
	return strings.HasPrefix(requestPath, "/api/") ||
		requestPath == "/healthz" ||
		requestPath == "/readyz" ||
		requestPath == "/openapi.yaml" ||
		strings.HasPrefix(requestPath, "/swagger")
}
