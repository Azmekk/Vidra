// @title Vidra API
// @version 2.0
// @description REST API for Vidra video downloader and manager
// @contact.name Martin Yordanov
// @contact.url https://github.com/Azmekk/Vidra
// @BasePath /
package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Azmekk/Vidra/backend/gen/database"
	_ "github.com/Azmekk/Vidra/backend/gen/docs/swagger"
	"github.com/Azmekk/Vidra/backend/handlers"
	vmw "github.com/Azmekk/Vidra/backend/middleware"
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

	if os.Getenv("DATABASE_URL") != "" {
		if _, err := os.Stat(cfg.DBPath); errors.Is(err, fs.ErrNotExist) {
			log.Fatal("❌ DATABASE_URL is set, but Vidra 2 stores its data in SQLite. Migrate your Vidra 1.x data first: https://github.com/Azmekk/Vidra/blob/master/docs/MIGRATING.md")
		}
	}

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
			slog.Warn("failed to resize video cache", "error", err)
		}
	})

	ytdlp := services.NewYtdlpService(settings, filepath.Dir(cfg.DBPath))
	ytdlp.Start(ctx)
	downloader := services.NewDownloaderService(store, queries, ws, settings, ytdlp, caps, cfg.DownloadsDir)
	downloader.RecoverInterrupted(ctx)

	backups := services.NewBackupService(queries, db, ws, cfg.DownloadsDir)
	downloader.OnFileCompleted(backups.FileCompleted)
	downloader.OnFilesDeleted(backups.FilesDeleted)
	backups.Start(ctx)

	videoHandler := handlers.NewVideoHandler(store, downloader, ytdlp, settings)

	auth := services.NewAuthService(queries)
	if err := auth.Init(ctx); err != nil {
		log.Fatalf("❌ Cannot initialise auth: %v", err)
	}
	go auth.PruneSessions(ctx)
	authMiddleware := &vmw.Auth{Service: auth, InsecureCookies: cfg.InsecureCookies}
	authHandler := handlers.NewAuthHandler(authMiddleware)

	r := chi.NewRouter()
	r.Use(vmw.ClientIP)
	r.Use(vmw.LogFailures)
	r.Use(middleware.Recoverer)
	r.Use(middleware.GetHead)

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Get("/status", authHandler.GetStatus)
			r.Post("/logout", authHandler.Logout)
			r.With(vmw.RateLimitByIP(10, time.Minute)).Post("/login", authHandler.Login)
			r.With(vmw.RateLimitByIP(10, time.Minute)).Post("/setup", authHandler.Setup)
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Put("/password", authHandler.ChangePassword)
				r.Get("/tokens", authHandler.ListAPITokens)
				r.Post("/tokens", authHandler.CreateAPIToken)
				r.Delete("/tokens/{id}", authHandler.DeleteAPIToken)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/ws", ws.HandleConnections)
			r.Mount("/videos", routers.VideoRouter(videoHandler))
			r.Mount("/files", routers.FileRouter(videoHandler))
			r.Mount("/encoding", routers.EncodingRouter(handlers.NewEncodingHandler(store, caps)))
			r.Mount("/errors", routers.ErrorRouter(handlers.NewErrorHandler(queries)))
			r.Mount("/yt-dlp", routers.YtDlpRouter(handlers.NewYtDlpHandler(ytdlp)))
			r.Mount("/system", routers.SystemRouter(handlers.NewSystemHandler(cfg.DownloadsDir)))
			r.Mount("/settings", routers.SettingsRouter(handlers.NewSettingsHandler(settings, caps)))
			r.Mount("/backups", routers.BackupRouter(handlers.NewBackupHandler(backups)))
		})
	})
	r.With(authMiddleware.RequireAuth).Get("/swagger/*", httpSwagger.WrapHandler)
	r.Handle("/*", web.Handler())

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	slog.Info("Vidra is running", "port", cfg.Port, "db", cfg.DBPath, "downloads", cfg.DownloadsDir, "insecure_cookies", cfg.InsecureCookies)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
