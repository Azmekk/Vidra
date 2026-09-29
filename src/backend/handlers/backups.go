package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/utils"
	"github.com/go-chi/chi/v5"
)

type BackupHandler struct {
	Backups *services.BackupService
}

func NewBackupHandler(backups *services.BackupService) *BackupHandler {
	return &BackupHandler{Backups: backups}
}

type BackupOverviewResponse struct {
	Available bool                       `json:"available"`
	Version   string                     `json:"version"`
	Targets   []services.BackupTargetDTO `json:"targets"`
}

// ListBackupTargets godoc
// @Summary List backup targets
// @Description Secrets in each config are masked.
// @ID listBackupTargets
// @Tags backups
// @Produce json
// @Success 200 {object} BackupOverviewResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/backups/targets [get]
func (h *BackupHandler) List(w http.ResponseWriter, r *http.Request) {
	targets, err := h.Backups.List(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, BackupOverviewResponse{
		Available: h.Backups.Available(), Version: h.Backups.Version(), Targets: targets,
	})
}

// CreateBackupTarget godoc
// @Summary Create a backup target
// @ID createBackupTarget
// @Tags backups
// @Accept json
// @Produce json
// @Param request body services.BackupTargetInput true "Target"
// @Success 201 {object} services.BackupTargetDTO
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/backups/targets [post]
func (h *BackupHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req services.BackupTargetInput
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	t, err := h.Backups.Create(r.Context(), req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, t)
}

// UpdateBackupTarget godoc
// @Summary Update a backup target
// @Description Send masked or empty secrets to keep the stored values. The provider cannot change.
// @ID updateBackupTarget
// @Tags backups
// @Accept json
// @Produce json
// @Param id path string true "Target ID"
// @Param request body services.BackupTargetInput true "Target"
// @Success 200 {object} services.BackupTargetDTO
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/backups/targets/{id} [put]
func (h *BackupHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req services.BackupTargetInput
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	t, err := h.Backups.Update(r.Context(), chi.URLParam(r, "id"), req)
	if errors.Is(err, sql.ErrNoRows) {
		utils.RespondWithError(w, http.StatusNotFound, "backup target not found")
		return
	}
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, t)
}

// DeleteBackupTarget godoc
// @Summary Delete a backup target
// @Description Remote files are left untouched.
// @ID deleteBackupTarget
// @Tags backups
// @Param id path string true "Target ID"
// @Success 204
// @Router /api/backups/targets/{id} [delete]
func (h *BackupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.Backups.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// TestBackupTarget godoc
// @Summary Test a backup target
// @Description Creates the remote folder if needed and lists it.
// @ID testBackupTarget
// @Tags backups
// @Param id path string true "Target ID"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/backups/targets/{id}/test [post]
func (h *BackupHandler) Test(w http.ResponseWriter, r *http.Request) {
	if err := h.Backups.Test(r.Context(), chi.URLParam(r, "id")); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RunBackupTarget godoc
// @Summary Run a full backup now
// @Description Copies every stored file and a database snapshot. Progress arrives as backup_status events.
// @ID runBackupTarget
// @Tags backups
// @Param id path string true "Target ID"
// @Success 202
// @Failure 409 {object} utils.ErrorResponse
// @Router /api/backups/targets/{id}/run [post]
func (h *BackupHandler) Run(w http.ResponseWriter, r *http.Request) {
	if err := h.Backups.Run(chi.URLParam(r, "id")); err != nil {
		utils.RespondWithError(w, http.StatusConflict, err.Error())
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
