package main

import (
	"context"
	"log"
	"net/http"

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
	conn, port := services.Bootstrap(ctx)
	defer conn.Close(ctx)

	queries := database.New(conn)
	downloader := services.NewDownloaderService(queries)
	videoHandler := handlers.NewVideoHandler(queries, downloader)
	errorHandler := handlers.NewErrorHandler(queries)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Mount routes
	r.Mount("/api/videos", routers.VideoRouter(videoHandler))
	r.Mount("/api/system/errors", routers.ErrorRouter(errorHandler))

	// Start server
	addr := ":" + port
	log.Printf("🌐 Server is running on http://localhost%s\n", addr)
	log.Println("✨ Ready to handle requests!")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
