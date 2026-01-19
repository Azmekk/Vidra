// @title Vidra API
// @version 1.0
// @description REST API for Vidra video downloader and manager
// @termsOfService https://github.com/Azmekk/Vidra
// @contact.name Martin Yordanov
// @contact.url https://github.com/Azmekk/Vidra
// @contact.email martin.yordanov@vexbyte.com
// @BasePath /
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/Azmekk/Vidra/backend/routers"
	"github.com/Azmekk/Vidra/backend/services"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/Azmekk/Vidra/backend/gen/docs/swagger"
)

func main() {
	ctx := context.Background()
	pool, port := services.Bootstrap(ctx)
	defer pool.Close()

	queries := database.New(pool)
	wsService := services.NewWebSocketService()
	go wsService.Run()

	downloader := services.NewDownloaderService(queries, wsService)
	videoHandler := handlers.NewVideoHandler(queries, downloader, wsService)
	errorHandler := handlers.NewErrorHandler(queries)
	ytdlpHandler := handlers.NewYtDlpHandler(queries, downloader)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// WebSocket endpoint
	r.Get("/api/ws", wsService.HandleConnections)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Mount routes
	r.Mount("/api/videos", routers.VideoRouter(videoHandler))
	r.Mount("/api/errors", routers.ErrorRouter(errorHandler))
	r.Mount("/api/yt-dlp", routers.YtDlpRouter(ytdlpHandler))

	// Serve downloads folder locally if VIDRA_DEV_ENVIRONMENT=true
	if os.Getenv("VIDRA_DEV_ENVIRONMENT") == "true" {
		workDir, _ := os.Getwd()
		filesDir := http.Dir(filepath.Join(workDir, "downloads"))
		r.Handle("/downloads/*", http.StripPrefix("/downloads/", http.FileServer(filesDir)))
		log.Println("📂 Serving /downloads locally (DEV mode)")
	}

	// Start server
	addr := ":" + port
	log.Printf("🌐 Server is running on http://localhost%s\n", addr)
	log.Println("✨ Ready to handle requests!")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
