package handlers

import (
	"net/http"

	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/services/encoding"
	"github.com/Azmekk/Vidra/backend/utils"
)

type SettingsHandler struct {
	Settings *services.SettingsService
	Caps     *encoding.Capabilities
}

func NewSettingsHandler(settings *services.SettingsService, caps *encoding.Capabilities) *SettingsHandler {
	return &SettingsHandler{Settings: settings, Caps: caps}
}

// GetSettings godoc
// @Summary Get application settings
// @ID getSettings
// @Tags settings
// @Produce json
// @Success 200 {object} services.Settings
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/settings [get]
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.Settings.Get(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, settings)
}

// UpdateSettings godoc
// @Summary Update application settings
// @ID updateSettings
// @Tags settings
// @Accept json
// @Produce json
// @Param settings body services.Settings true "Settings"
// @Success 200 {object} services.Settings
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/settings [put]
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req services.Settings
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	if err := req.Validate(h.Caps); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	settings, err := h.Settings.Update(r.Context(), req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, settings)
}
