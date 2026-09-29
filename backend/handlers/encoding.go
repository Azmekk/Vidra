package handlers

import (
	"net/http"

	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/services/encoding"
	"github.com/Azmekk/Vidra/backend/utils"
)

type EncodingHandler struct {
	Store *services.VideoStore
	Caps  *encoding.Capabilities
}

func NewEncodingHandler(store *services.VideoStore, caps *encoding.Capabilities) *EncodingHandler {
	return &EncodingHandler{Store: store, Caps: caps}
}

// GetCapabilities godoc
// @Summary List encoders supported by this server's ffmpeg
// @Description Hardware encoders are test-encoded at startup; only working ones are marked available.
// @ID getEncodingCapabilities
// @Tags encoding
// @Produce json
// @Success 200 {object} encoding.Capabilities
// @Router /api/encoding/capabilities [get]
func (h *EncodingHandler) GetCapabilities(w http.ResponseWriter, _ *http.Request) {
	utils.RespondWithJSON(w, http.StatusOK, h.Caps)
}

type RecommendRequest struct {
	Encoding     encoding.Request `json:"encoding"`
	SourceFileID string           `json:"sourceFileId,omitempty"`
	Source       *encoding.Source `json:"source,omitempty"`
}

// Recommend godoc
// @Summary Recommend encoding settings for a goal
// @Description Pass either a stored version (sourceFileId) or source stream info from yt-dlp metadata.
// @ID recommendEncoding
// @Tags encoding
// @Accept json
// @Produce json
// @Param request body RecommendRequest true "Goal and source"
// @Success 200 {object} encoding.Recommendation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/encoding/recommend [post]
func (h *EncodingHandler) Recommend(w http.ResponseWriter, r *http.Request) {
	var req RecommendRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	if err := req.Encoding.Validate(h.Caps); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var src encoding.Source
	switch {
	case req.SourceFileID != "":
		_, file, err := h.Store.GetByFile(r.Context(), req.SourceFileID)
		if respondStoreError(w, err) {
			return
		}
		src = services.SourceOf(file)
	case req.Source != nil:
		src = *req.Source
	}
	utils.RespondWithJSON(w, http.StatusOK, req.Encoding.Resolve(h.Caps, src))
}
