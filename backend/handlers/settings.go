package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/utils"
)

type SettingsHandler struct {
	Settings *services.SettingsService
}

func NewSettingsHandler(settings *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{
		Settings: settings,
	}
}

type SettingsResponse struct {
	ProxyUrl          string `json:"proxyUrl"`
	DefaultReEncode   bool   `json:"defaultReEncode"`
	DefaultVideoCodec string `json:"defaultVideoCodec"`
	DefaultAudioCodec string `json:"defaultAudioCodec"`
	DefaultCrf        int    `json:"defaultCrf"`
	Theme             string `json:"theme"`
}

type UpdateSettingsRequest struct {
	ProxyUrl          string `json:"proxyUrl"`
	DefaultReEncode   bool   `json:"defaultReEncode"`
	DefaultVideoCodec string `json:"defaultVideoCodec"`
	DefaultAudioCodec string `json:"defaultAudioCodec"`
	DefaultCrf        int    `json:"defaultCrf"`
	Theme             string `json:"theme"`
}

func (r *UpdateSettingsRequest) Validate() error {
	validCodecs := map[string]bool{"libx264": true, "libvpx-vp9": true, "vp9_qsv": true}
	if !validCodecs[r.DefaultVideoCodec] {
		return fmt.Errorf("invalid video codec: must be libx264, libvpx-vp9, or vp9_qsv")
	}

	validAudioCodecs := map[string]bool{"aac": true, "libopus": true}
	if !validAudioCodecs[r.DefaultAudioCodec] {
		return fmt.Errorf("invalid audio codec: must be aac or libopus")
	}

	if r.DefaultCrf < 0 || r.DefaultCrf > 51 {
		return fmt.Errorf("invalid CRF value: must be between 0 and 51")
	}

	validThemes := map[string]bool{"light": true, "dark": true, "system": true}
	if !validThemes[r.Theme] {
		return fmt.Errorf("invalid theme: must be light, dark, or system")
	}

	return nil
}

// GetSettings godoc
// @Summary Get application settings
// @Description Get current application settings
// @ID getSettings
// @Tags settings
// @Produce json
// @Success 200 {object} SettingsResponse
// @Failure 500 {object} map[string]string
// @Router /api/settings [get]
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.Settings.GetSettings(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, SettingsResponse{
		ProxyUrl:          settings.ProxyUrl,
		DefaultReEncode:   settings.DefaultReEncode,
		DefaultVideoCodec: settings.DefaultVideoCodec,
		DefaultAudioCodec: settings.DefaultAudioCodec,
		DefaultCrf:        settings.DefaultCrf,
		Theme:             settings.Theme,
	})
}

// UpdateSettings godoc
// @Summary Update application settings
// @Description Update application settings
// @ID updateSettings
// @Tags settings
// @Accept json
// @Produce json
// @Param settings body UpdateSettingsRequest true "Settings to update"
// @Success 200 {object} SettingsResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/settings [put]
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := req.Validate(); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	settings, err := h.Settings.UpdateSettings(r.Context(), services.SettingsDTO{
		ProxyUrl:          req.ProxyUrl,
		DefaultReEncode:   req.DefaultReEncode,
		DefaultVideoCodec: req.DefaultVideoCodec,
		DefaultAudioCodec: req.DefaultAudioCodec,
		DefaultCrf:        req.DefaultCrf,
		Theme:             req.Theme,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, SettingsResponse{
		ProxyUrl:          settings.ProxyUrl,
		DefaultReEncode:   settings.DefaultReEncode,
		DefaultVideoCodec: settings.DefaultVideoCodec,
		DefaultAudioCodec: settings.DefaultAudioCodec,
		DefaultCrf:        settings.DefaultCrf,
		Theme:             settings.Theme,
	})
}
