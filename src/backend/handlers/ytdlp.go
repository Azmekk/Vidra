package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/utils"
)

type YtDlpHandler struct {
	Ytdlp *services.YtdlpService
}

func NewYtDlpHandler(ytdlp *services.YtdlpService) *YtDlpHandler {
	return &YtDlpHandler{Ytdlp: ytdlp}
}

// GetYtdlp godoc
// @Summary Get the installed yt-dlp version, the pin and recent releases
// @ID getYtdlp
// @Tags ytdlp
// @Produce json
// @Success 200 {object} services.YtdlpStatus
// @Router /api/yt-dlp [get]
func (h *YtDlpHandler) GetYtdlp(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithJSON(w, http.StatusOK, h.Ytdlp.Status(r.Context()))
}

// UpdateYtdlp godoc
// @Summary Install the latest yt-dlp, or the pinned version
// @ID updateYtdlp
// @Tags ytdlp
// @Produce json
// @Success 200 {object} services.YtdlpStatus
// @Failure 502 {object} utils.ErrorResponse
// @Router /api/yt-dlp/update [post]
func (h *YtDlpHandler) UpdateYtdlp(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, h.Ytdlp.Sync(r.Context(), true))
}

type PinYtdlpRequest struct {
	Version string `json:"version"`
}

// PinYtdlp godoc
// @Summary Install a specific yt-dlp version and keep it
// @ID pinYtdlp
// @Tags ytdlp
// @Accept json
// @Produce json
// @Param request body PinYtdlpRequest true "Version to pin"
// @Success 200 {object} services.YtdlpStatus
// @Failure 400 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Router /api/yt-dlp/pin [put]
func (h *YtDlpHandler) PinYtdlp(w http.ResponseWriter, r *http.Request) {
	var req PinYtdlpRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	h.respond(w, r, h.Ytdlp.Pin(r.Context(), strings.TrimSpace(req.Version)))
}

// UnpinYtdlp godoc
// @Summary Follow the latest yt-dlp release again
// @ID unpinYtdlp
// @Tags ytdlp
// @Produce json
// @Success 200 {object} services.YtdlpStatus
// @Failure 502 {object} utils.ErrorResponse
// @Router /api/yt-dlp/pin [delete]
func (h *YtDlpHandler) UnpinYtdlp(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, h.Ytdlp.Unpin(r.Context()))
}

func (h *YtDlpHandler) respond(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidYtdlpVersion):
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		utils.RespondWithError(w, http.StatusBadGateway, err.Error())
	default:
		utils.RespondWithJSON(w, http.StatusOK, h.Ytdlp.Status(r.Context()))
	}
}
