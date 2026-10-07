package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
)

//go:embed all:dist
var files embed.FS

type EmbedInfo struct {
	Count int      `json:"count"`
	Files []string `json:"files"`
	Err   string   `json:"err,omitempty"`
}

func Info() EmbedInfo {
	sub, err := fs.Sub(files, "dist")
	if err != nil {
		return EmbedInfo{Err: err.Error()}
	}
	var info EmbedInfo
	_ = fs.WalkDir(sub, ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			info.Files = append(info.Files, p)
		}
		return nil
	})
	info.Count = len(info.Files)
	return info
}

func Handler() (http.Handler, error) {
	sub, err := fs.Sub(files, "dist")
	if err != nil {
		slog.Error("web: fs.Sub failed", "err", err)
		return nil, err
	}

	info := Info()
	slog.Info("web: embedded", "count", info.Count, "files", info.Files)

	if _, err := fs.Stat(sub, "index.html"); err != nil {
		slog.Warn("web: no index.html, serving placeholder", "err", err)
		return placeholder(), nil
	}

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

func placeholder() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html><body style="font-family:system-ui;padding:2rem"><h1>Pulse</h1><p>SPA not bundled in this image. API is at <code>/api/*</code>.</p></body></html>`)
	})
}

func DebugHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Info())
	})
}
