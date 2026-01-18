package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Azmekk/Vidra/backend/gen/database"
)

type ErrorHandler struct {
	Queries *database.Queries
}

func NewErrorHandler(queries *database.Queries) *ErrorHandler {
	return &ErrorHandler{Queries: queries}
}

type ErrorResponse struct {
	ID           string `json:"id"`
	VideoID      string `json:"videoId"`
	ErrorMessage string `json:"errorMessage"`
	Command      string `json:"command"`
	Output       string `json:"output"`
	CreatedAt    string `json:"createdAt"`
}

func mapErrorToResponse(e database.Error) ErrorResponse {
	var idStr, videoIdStr string
	e.ID.Scan(&idStr)
	e.VideoID.Scan(&videoIdStr)

	return ErrorResponse{
		ID:           idStr,
		VideoID:      videoIdStr,
		ErrorMessage: e.ErrorMessage,
		Command:      e.Command,
		Output:       e.Output,
		CreatedAt:    e.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ListRecentErrors godoc
// @Summary List recent errors
// @Description Get a list of the most recent system errors
// @ID listRecentErrors
// @Tags system
// @Accept json
// @Produce json
// @Param limit query int false "Limit number of results" default(10)
// @Success 200 {array} ErrorResponse
// @Failure 500 {object} map[string]string
// @Router /api/system/errors [get]
func (h *ErrorHandler) ListRecentErrors(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := int32(10)
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(l)
		}
	}

	errors, err := h.Queries.ListRecentErrors(r.Context(), limit)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]ErrorResponse, len(errors))
	for i, e := range errors {
		responses[i] = mapErrorToResponse(e)
	}

	h.respondWithJSON(w, http.StatusOK, responses)
}

func (h *ErrorHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *ErrorHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
