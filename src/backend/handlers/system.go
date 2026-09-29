package handlers

import (
	"net/http"

	"github.com/Azmekk/Vidra/backend/utils"
)

type SystemHandler struct {
	DownloadsDir string
}

func NewSystemHandler(downloadsDir string) *SystemHandler {
	return &SystemHandler{DownloadsDir: downloadsDir}
}

type SystemInfoResponse struct {
	Status        string `json:"status"`
	DownloadsSize int64  `json:"downloadsSize"`
}

// GetSystemInfo godoc
// @Summary Get server status and downloads directory size
// @ID getSystemInfo
// @Tags system
// @Produce json
// @Success 200 {object} SystemInfoResponse
// @Router /api/system/info [get]
func (h *SystemHandler) GetSystemInfo(w http.ResponseWriter, _ *http.Request) {
	size, _ := utils.GetDirSize(h.DownloadsDir)
	utils.RespondWithJSON(w, http.StatusOK, SystemInfoResponse{Status: "ok", DownloadsSize: size})
}
