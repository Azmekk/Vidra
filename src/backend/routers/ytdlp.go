package routers

import (
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/go-chi/chi/v5"
)

func YtDlpRouter(h *handlers.YtDlpHandler) chi.Router {
	r := chi.NewRouter()

	r.Post("/update", h.UpdateYtdlp)

	return r
}
