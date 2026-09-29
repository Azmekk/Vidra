// @title Vidra API
// @version 2.0
// @description REST API for Vidra video downloader and manager
// @contact.name Martin Yordanov
// @contact.url https://github.com/Azmekk/Vidra
// @BasePath /
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azmekk/Vidra/backend/gen/database"
	_ "github.com/Azmekk/Vidra/backend/gen/docs/swagger"
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/Azmekk/Vidra/backend/routers"
	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/services/encoding"
	"github.com/Azmekk/Vidra/backend/web"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	ctx := context.Background()
	cfg := services.LoadConfig()

	if err := os.MkdirAll(cfg.DownloadsDir, 0o755); err != nil {
		log.Fatalf("❌ Cannot create downloads directory: %v", err)
	}
	db, err := services.OpenDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	defer db.Close()

	queries := database.New(db)
	ws := services.NewWebSocketService()
	settings := services.NewSettingsService(queries)
	caps := encoding.Detect(ctx)

	store := services.NewVideoStore(queries, ws)
	if err := store.Warm(ctx, settings.MustGet(ctx).CacheSize); err != nil {
		log.Fatalf("❌ Cannot load videos: %v", err)
	}
	settings.OnChange(func(s services.Settings) {
		if err := store.Warm(context.Background(), s.CacheSize); err != nil {
			log.Printf("WARN: failed to resize video cache: %v\n", err)
		}
	})

	ytdlp := services.NewYtdlpService(settings)
	downloader := services.NewDownloaderService(store, queries, ws, settings, ytdlp, caps, cfg.DownloadsDir)
	downloader.RecoverInterrupted(ctx)

	videoHandler := handlers.NewVideoHandler(store, downloader, ytdlp, settings)

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.GetHead)

	r.Route("/api", func(r chi.Router) {
		r.Get("/ws", ws.HandleConnections)
		r.Mount("/videos", routers.VideoRouter(videoHandler))
		r.Mount("/files", routers.FileRouter(videoHandler))
		r.Mount("/encoding", routers.EncodingRouter(handlers.NewEncodingHandler(store, caps)))
		r.Mount("/errors", routers.ErrorRouter(handlers.NewErrorHandler(queries)))
		r.Mount("/yt-dlp", routers.YtDlpRouter(handlers.NewYtDlpHandler(ytdlp)))
		r.Mount("/system", routers.SystemRouter(handlers.NewSystemHandler(cfg.DownloadsDir)))
		r.Mount("/settings", routers.SettingsRouter(handlers.NewSettingsHandler(settings, caps)))
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Handle("/*", web.Handler())

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("🌐 Vidra is running on http://localhost:%s\n", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
