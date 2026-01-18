package routers

import (
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/go-chi/chi/v5"
)

func ErrorRouter(h *handlers.ErrorHandler) chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListRecentErrors)

	return r
}
