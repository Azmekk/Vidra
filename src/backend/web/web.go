package web

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// build/app is produced by the frontend's `bun run build`.
//
//go:embed all:build
var build embed.FS

// Handler serves the SPA: hashed assets are cached forever, other files are
// revalidated against a content ETag, and every unknown path falls back to
// index.html for client-side routing.
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

	etags := make(map[string]string)
	err = fs.WalkDir(app, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasPrefix(name, "assets/") {
			return err
		}
		data, err := fs.ReadFile(app, name)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		etags[name] = `"` + hex.EncodeToString(sum[:8]) + `"`
		return nil
	})
	if err != nil {
		panic(err)
	}

	files := http.FileServerFS(app)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(app, name); err != nil {
			name = "index.html"
		}
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("ETag", etags[name])
		http.ServeFileFS(w, r, app, name)
	})
}
