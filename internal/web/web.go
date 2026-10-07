package web

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
)

//go:embed all:dist
var files embed.FS

func Handler() (http.Handler, error) {
	sub, err := fs.Sub(files, "dist")
	if err != nil {
		return nil, err
	}

	count := 0
	_ = fs.WalkDir(sub, ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			count++
		}
		return nil
	})
	slog.Info("spa files embedded", "count", count)

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(sub, path); err == nil {
			if strings.HasPrefix(path, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, r)
			return
		}
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	}), nil
}
