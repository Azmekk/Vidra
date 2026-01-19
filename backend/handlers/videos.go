package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/utils"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type VideoHandler struct {
	Queries    *database.Queries
	Downloader *services.DownloaderService
	Ws         *services.WebSocketService
}

func NewVideoHandler(queries *database.Queries, downloader *services.DownloaderService, ws *services.WebSocketService) *VideoHandler {
	return &VideoHandler{
		Queries:    queries,
		Downloader: downloader,
		Ws:         ws,
	}
}

type MetadataRequest struct {
	URL string `json:"url"`
}

func (r *MetadataRequest) Validate() error {
	if r.URL == "" {
		return fmt.Errorf("url is required")
	}
	return nil
}

// GetMetadata godoc
// @Summary Get video metadata and options
// @Description Fetch available formats and metadata for a given URL using yt-dlp
// @ID getMetadata
// @Tags videos
// @Accept json
// @Produce json
// @Param request body MetadataRequest true "Video URL"
// @Success 200 {object} services.VideoMetadata
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/videos/metadata [post]
func (h *VideoHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	var req MetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := req.Validate(); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	sanitizedURL, err := utils.SanitizeURL(req.URL)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	metadata, err := h.Downloader.GetVideoMetadata(r.Context(), sanitizedURL)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, metadata)
}

type CreateVideoRequest struct {
	Name        string `json:"name"`
	DownloadURL string `json:"downloadUrl"`
	FormatID    string `json:"formatId"`
}

func (r *CreateVideoRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.DownloadURL == "" {
		return fmt.Errorf("downloadUrl is required")
	}
	return nil
}

type VideoResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	FileName          string `json:"fileName,omitempty"`
	ThumbnailFileName string `json:"thumbnailFileName,omitempty"`
	DownloadURL       string `json:"downloadUrl"`
	DownloadStatus    string `json:"downloadStatus"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

func mapVideoToResponse(v database.Video) VideoResponse {
	return VideoResponse{
		ID:                v.ID.String(),
		Name:              v.Name,
		FileName:          v.FileName.String,
		ThumbnailFileName: v.ThumbnailFileName.String,
		DownloadURL:       v.OriginalUrl,
		DownloadStatus:    v.DownloadStatus,
		CreatedAt:         v.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         v.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CreateVideo godoc
// @Summary Create a new video download task
// @Description Create a new video record and start background download
// @ID createVideo
// @Tags videos
// @Accept json
// @Produce json
// @Param video body CreateVideoRequest true "Video details"
// @Success 201 {object} VideoResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/videos [post]
func (h *VideoHandler) CreateVideo(w http.ResponseWriter, r *http.Request) {
	// Read body for logging
	bodyBytes, _ := io.ReadAll(r.Body)
	log.Printf("DEBUG: CreateVideo raw body: %s\n", string(bodyBytes))
	// Restore body for json decoder
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req CreateVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode CreateVideo request: %v\n", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	log.Printf("DEBUG: CreateVideo parsed request: %+v\n", req)

	if err := req.Validate(); err != nil {
		log.Printf("ERROR: CreateVideo validation failed: %v\n", err)
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	sanitizedURL, err := utils.SanitizeURL(req.DownloadURL)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	log.Printf("INFO: Received request to download video: Name='%s', URL='%s', FormatID='%s'\n", req.Name, sanitizedURL, req.FormatID)

	video, err := h.Queries.CreateVideo(r.Context(), database.CreateVideoParams{
		Name:           req.Name,
		OriginalUrl:    sanitizedURL,
		DownloadStatus: string(services.StatusDownloading),
	})
	if err != nil {
		log.Printf("ERROR: Failed to create video record in database: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	idStr := video.ID.String()
	log.Printf("INFO: Successfully created video record in database: ID=%s\n", idStr)

	// Start background download
	log.Printf("INFO: Starting background download for video ID=%s\n", idStr)
	h.Downloader.StartDownload(context.Background(), video.ID, sanitizedURL, req.FormatID, req.Name)

	h.Ws.Broadcast("video_created", mapVideoToResponse(video))

	utils.RespondWithJSON(w, http.StatusCreated, mapVideoToResponse(video))
}

// GetProgress godoc
// @Summary Get download progress
// @Description Get the current download progress of a video by ID
// @ID getProgress
// @Tags videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID"
// @Success 200 {object} services.DownloadProgressDTO
// @Failure 404 {object} map[string]string
// @Router /api/videos/{id}/progress [get]
func (h *VideoHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing video ID")
		return
	}
	progress, ok := h.Downloader.GetProgress(idStr)
	if !ok {
		utils.RespondWithError(w, http.StatusNotFound, "Progress not found for this ID")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, progress)
}

// ListAllProgress godoc
// @Summary List all video download progress
// @Description Get the current download progress for all active video downloads
// @ID listAllProgress
// @Tags videos
// @Accept json
// @Produce json
// @Success 200 {object} map[string]services.DownloadProgressDTO
// @Router /api/videos/progress [get]
func (h *VideoHandler) ListAllProgress(w http.ResponseWriter, r *http.Request) {
	allProgress := h.Downloader.GetAllProgress()
	utils.RespondWithJSON(w, http.StatusOK, allProgress)
}

// GetVideo godoc
// @Summary Get a video by ID
// @Description Get details of a specific video
// @ID getVideo
// @Tags videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID"
// @Success 200 {object} VideoResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/videos/{id} [get]
func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	video, err := h.Queries.GetVideo(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Video not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, mapVideoToResponse(video))
}

// ListVideos godoc
// @Summary List all videos
// @Description Get a list of all videos with optional searching and ordering
// @ID listVideos
// @Tags videos
// @Accept json
// @Produce json
// @Param search query string false "Search by name or URL"
// @Param order query string false "Order by (name_asc, name_desc, created_at_asc, created_at_desc, status_asc, status_desc)"
// @Success 200 {array} VideoResponse
// @Failure 500 {object} map[string]string
// @Router /api/videos [get]
func (h *VideoHandler) ListVideos(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	order := r.URL.Query().Get("order")

	videos, err := h.Queries.ListVideos(r.Context(), database.ListVideosParams{
		Search:   pgtype.Text{String: search, Valid: true},
		Ordering: pgtype.Text{String: order, Valid: true},
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]VideoResponse, len(videos))
	for i, v := range videos {
		responses[i] = mapVideoToResponse(v)
	}

	utils.RespondWithJSON(w, http.StatusOK, responses)
}

type UpdateVideoRequest struct {
	Name string `json:"name"`
}

func (r *UpdateVideoRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

// UpdateVideo godoc
// @Summary Update a video
// @Description Update video details like name
// @ID updateVideo
// @Tags videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID"
// @Param video body UpdateVideoRequest true "Updated video details"
// @Success 200 {object} VideoResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/videos/{id} [put]
func (h *VideoHandler) UpdateVideo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	var req UpdateVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := req.Validate(); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	video, err := h.Queries.UpdateVideoName(r.Context(), database.UpdateVideoNameParams{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, mapVideoToResponse(video))
}

// DeleteVideo godoc
// @Summary Delete a video
// @Description Delete a video record by ID and its associated files
// @ID deleteVideo
// @Tags videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/videos/{id} [delete]
func (h *VideoHandler) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	// Fetch video details to get filenames for cleanup
	video, err := h.Queries.GetVideo(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Video not found")
		return
	}

	// Delete from database first
	err = h.Queries.DeleteVideo(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Delete files from filesystem
	h.Downloader.DeleteVideoFiles(video.FileName.String, video.ThumbnailFileName.String)

	h.Ws.Broadcast("video_deleted", map[string]string{"id": idStr})

	w.WriteHeader(http.StatusNoContent)
}
