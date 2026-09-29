package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// build/app is produced by the frontend's `bun run build`.
//
//go:embed all:build
var build embed.FS

// Handler serves the SPA: hashed assets are cached forever, every other
// unknown path falls back to index.html for client-side routing.
func Handler() http.Handler {
	app, err := fs.Sub(build, "build/app")
	if err != nil {
		panic(err)
	}
	if _, err := fs.Stat(app, "index.html"); err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Frontend not built. Run `bun run build` in frontend/ or use the Vite dev server.", http.StatusServiceUnavailable)
		})
	}

	files := http.FileServerFS(app)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" || name == "index.html" {
			serveIndex(w, r, app)
			return
		}
		if _, err := fs.Stat(app, name); err != nil {
			serveIndex(w, r, app)
			return
		}
		switch {
		case strings.HasPrefix(name, "assets/"):
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case name == "sw.js" || strings.HasSuffix(name, ".webmanifest"):
			w.Header().Set("Cache-Control", "no-cache")
		default:
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}
		files.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request, app fs.FS) {
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFileFS(w, r, app, "index.html")
}
