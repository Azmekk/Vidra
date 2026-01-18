package routers

import (
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/go-chi/chi/v5"
)

func VideoRouter(h *handlers.VideoHandler) chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.CreateVideo)
	r.Get("/", h.ListVideos)
	r.Post("/metadata", h.GetMetadata)
	r.Get("/progress", h.ListAllProgress)
	r.Get("/{id}", h.GetVideo)
	r.Get("/{id}/progress", h.GetProgress)
	r.Delete("/{id}", h.DeleteVideo)
	// System routes (could be in a separate router, but putting here for simplicity as requested)
	r.Post("/update-ytdlp", h.UpdateYtdlp)
	return r
}
