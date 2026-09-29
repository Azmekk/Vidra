package routers

import (
	"github.com/Azmekk/Vidra/backend/handlers"
	"github.com/go-chi/chi/v5"
)

func YtDlpRouter(h *handlers.YtDlpHandler) chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetYtdlp)
	r.Post("/update", h.UpdateYtdlp)
	r.Put("/pin", h.PinYtdlp)
	r.Delete("/pin", h.UnpinYtdlp)

	return r
}
