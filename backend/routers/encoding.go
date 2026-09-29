package routers

import (
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/go-chi/chi/v5"
)

func EncodingRouter(h *handlers.EncodingHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/capabilities", h.GetCapabilities)
	r.Post("/recommend", h.Recommend)
	return r
}
