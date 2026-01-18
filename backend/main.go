package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/Azmekk/Vidra/backend/gen/docs/swagger"
)

func main() {
	log.Println("Hello, World!")

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Mount auth routes
	//r.Mount("/api/auth", routers.AuthRouter())

	// Start server
	port := ":8080"
	log.Println("🚀 Fornext Authentication Service")
	log.Printf("🌐 Server is running on http://localhost%s\n", port)
	log.Println("✨ Ready to handle requests!")

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
