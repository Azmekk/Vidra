package handlers

import (
	"fmt"
	"net/http"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/utils"
)

type YtDlpHandler struct {
	Queries    *database.Queries
	Downloader *services.DownloaderService
}

func NewYtDlpHandler(queries *database.Queries, downloader *services.DownloaderService) *YtDlpHandler {
	return &YtDlpHandler{Queries: queries, Downloader: downloader}
}

// UpdateYtdlp godoc
// @Summary Update yt-dlp
// @Description Execute yt-dlp -U to update the binary
// @ID updateYtdlp
// @Tags ytdlp
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/yt-dlp/update [post]
func (h *YtDlpHandler) UpdateYtdlp(w http.ResponseWriter, r *http.Request) {
	output, err := h.Downloader.UpdateYtdlp(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Update failed: %v\nOutput: %s", err, output))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"output": output})
}
