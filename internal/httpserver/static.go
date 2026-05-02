package httpserver

import (
	"blazeginx/internal/config"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func NewStaticHandler(cfg config.Static) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cfg.Enabled {
			http.NotFound(w, r)
			return
		}

		root, err := filepath.Abs(cfg.Root)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		requestPath := filepath.Clean("/" + r.URL.Path)
		filePath := filepath.Join(root, strings.TrimPrefix(requestPath, "/"))

		if !isPathInside(root, filePath) {
			http.NotFound(w, r)
			return
		}

		if isRegularFile(filePath) {
			http.ServeFile(w, r, filePath)
			return
		}

		if filepath.Ext(requestPath) != "" {
			http.NotFound(w, r)
			return
		}

		indexPath := filepath.Join(root, "index.html")
		if isRegularFile(indexPath) {
			http.ServeFile(w, r, indexPath)
			return
		}

		http.NotFound(w, r)
	})
}

func isPathInside(root string, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}
