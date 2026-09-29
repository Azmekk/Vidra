package routers

import (
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/go-chi/chi/v5"
)

func BackupRouter(h *handlers.BackupHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/targets", h.List)
	r.Post("/targets", h.Create)
	r.Put("/targets/{id}", h.Update)
	r.Delete("/targets/{id}", h.Delete)
	r.Post("/targets/{id}/test", h.Test)
	r.Post("/targets/{id}/run", h.Run)
	return r
}
