package handlers

import (
	"net/http"

	"github.com/Azmekk/Vidra/backend/utils"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

type SystemInfoResponse struct {
	Status        string  `json:"status"`
	DiskUsageGB   float64 `json:"diskUsageGB"`
	DownloadsSize int64   `json:"downloadsSize"`
}

// GetSystemInfo godoc
// @Summary Get system information
// @Description Get server status and downloads directory size
// @ID getSystemInfo
// @Tags system
// @Produce json
// @Success 200 {object} SystemInfoResponse
// @Router /api/system/info [get]
func (h *SystemHandler) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	size, err := utils.GetDirSize("downloads")
	if err != nil {
		// If directory doesn't exist yet, it's fine, size is 0
		size = 0
	}

	utils.RespondWithJSON(w, http.StatusOK, SystemInfoResponse{
		Status:        "ok",
		DiskUsageGB:   float64(size) / (1024 * 1024 * 1024),
		DownloadsSize: size,
	})
}
