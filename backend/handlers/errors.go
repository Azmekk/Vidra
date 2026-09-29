package handlers

import (
	"net/http"
	"strings"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/utils"
)

type ErrorHandler struct {
	Queries *database.Queries
}

func NewErrorHandler(queries *database.Queries) *ErrorHandler {
	return &ErrorHandler{Queries: queries}
}

type ErrorResponse struct {
	ID           string  `json:"id"`
	VideoID      *string `json:"videoId,omitempty"`
	FileID       *string `json:"fileId,omitempty"`
	ErrorMessage string  `json:"errorMessage"`
	Command      string  `json:"command"`
	Output       string  `json:"output"`
	CreatedAt    string  `json:"createdAt"`
}

type PaginatedErrorResponse struct {
	TotalCount  int64           `json:"totalCount"`
	TotalPages  int             `json:"totalPages"`
	CurrentPage int             `json:"currentPage"`
	Limit       int             `json:"limit"`
	Errors      []ErrorResponse `json:"errors"`
}

// ListRecentErrors godoc
// @Summary List recent errors
// @ID listRecentErrors
// @Tags errors
// @Produce json
// @Param search query string false "Search by message, command or video ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {object} PaginatedErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/errors [get]
func (h *ErrorHandler) ListRecentErrors(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	page, limit := utils.Pagination(r, 10)

	total, err := h.Queries.CountErrors(r.Context(), search)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows, err := h.Queries.ListRecentErrors(r.Context(), database.ListRecentErrorsParams{
		Search: search, Limit: int64(limit), Offset: int64((page - 1) * limit),
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	errs := make([]ErrorResponse, len(rows))
	for i, e := range rows {
		errs[i] = ErrorResponse{
			ID: e.ID, VideoID: e.VideoID, FileID: e.FileID, ErrorMessage: e.ErrorMessage,
			Command: e.Command, Output: e.Output, CreatedAt: e.CreatedAt,
		}
	}
	utils.RespondWithJSON(w, http.StatusOK, PaginatedErrorResponse{
		TotalCount:  total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		CurrentPage: page,
		Limit:       limit,
		Errors:      errs,
	})
}
