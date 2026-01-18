package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/Azmekk/Vidra/backend/routers"
	"github.com/Azmekk/Vidra/backend/services"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/Azmekk/Vidra/backend/gen/docs/swagger"
)

func main() {
	ctx := context.Background()
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://postgres:postgres@localhost:5432/vidra?sslmode=disable"
	}

	conn, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	queries := database.New(conn)
	downloader := services.NewDownloaderService(queries)
	videoHandler := handlers.NewVideoHandler(queries, downloader)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Mount routes
	r.Mount("/api/videos", routers.VideoRouter(videoHandler))

	// Start server
	port := ":8080"
	log.Printf("🌐 Server is running on http://localhost%s\n", port)
	log.Println("✨ Ready to handle requests!")

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
