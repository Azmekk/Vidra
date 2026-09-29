package handlers

import (
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

type UpdateYtdlpResponse struct {
	Output string `json:"output"`
}

// UpdateYtdlp godoc
// @Summary Update yt-dlp
// @Description Runs yt-dlp -U to update the binary
// @ID updateYtdlp
// @Tags ytdlp
// @Produce json
// @Success 200 {object} UpdateYtdlpResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/yt-dlp/update [post]
func (h *YtDlpHandler) UpdateYtdlp(w http.ResponseWriter, r *http.Request) {
	output, err := h.Ytdlp.UpdateCommand(r.Context()).CombinedOutput()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Update failed: "+err.Error()+"\n"+string(output))
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, UpdateYtdlpResponse{Output: strings.TrimSpace(string(output))})
}
