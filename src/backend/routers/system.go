package routers

import (
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/go-chi/chi/v5"
)

func SystemRouter(h *handlers.SystemHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/info", h.GetSystemInfo)
	return r
}
