package routers

import (
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/go-chi/chi/v5"
)

func VideoRouter(h *handlers.VideoHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListVideos)
	r.Post("/", h.CreateVideo)
	r.Post("/quick", h.QuickDownload)
	r.Post("/metadata", h.GetMetadata)
	r.Get("/progress", h.ListProgress)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetVideo)
		r.Put("/", h.UpdateVideo)
		r.Delete("/", h.DeleteVideo)
		r.Get("/thumbnail", h.GetThumbnail)
		r.Post("/files", h.CreateVersion)
		r.Patch("/files/{fileId}", h.UpdateVersion)
		r.Delete("/files/{fileId}", h.DeleteVersion)
		r.Post("/files/{fileId}/cancel", h.CancelVersion)
	})
	return r
}

func FileRouter(h *handlers.VideoHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/{fileId}", h.GetFile)
	return r
}
