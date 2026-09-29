package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, ErrorResponse{Error: message})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

// DecodeJSON decodes a request body, rejecting unknown fields and oversized bodies.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return false
	}
	return true
}

// Pagination reads page and limit query parameters with sane bounds.
func Pagination(r *http.Request, defaultLimit int) (page, limit int) {
	page, limit = 1, defaultLimit
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = min(l, 100)
	}
	return page, limit
}
