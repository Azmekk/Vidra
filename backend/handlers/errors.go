package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/jackc/pgx/v5/pgtype"
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

	return ErrorResponse{
		ID:           e.ID.String(),
		VideoID:      e.VideoID.String(),
		ErrorMessage: e.ErrorMessage,
		Command:      e.Command,
		Output:       e.Output,
		CreatedAt:    e.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
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
// @Description Get a paginated list of the most recent system errors with optional searching
// @ID listRecentErrors
// @Tags errors
// @Accept json
// @Produce json
// @Param search query string false "Search by error message or command"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 10)"
// @Success 200 {object} PaginatedErrorResponse
// @Failure 500 {object} map[string]string
// @Router /api/errors [get]
func (h *ErrorHandler) ListRecentErrors(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	searchParam := pgtype.Text{String: search, Valid: true}

	totalCount, err := h.Queries.CountErrors(r.Context(), searchParam)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	errors, err := h.Queries.ListRecentErrors(r.Context(), database.ListRecentErrorsParams{
		Search: searchParam,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]ErrorResponse, len(errors))
	for i, e := range errors {
		responses[i] = mapErrorToResponse(e)
	}

	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))

	h.respondWithJSON(w, http.StatusOK, PaginatedErrorResponse{
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: page,
		Limit:       limit,
		Errors:      responses,
	})
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
