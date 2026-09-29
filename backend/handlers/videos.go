package handlers

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/services/encoding"
	"github.com/Azmekk/Vidra/backend/utils"
	"github.com/go-chi/chi/v5"
)

type VideoHandler struct {
	Store      *services.VideoStore
	Downloader *services.DownloaderService
	Ytdlp      *services.YtdlpService
	Settings   *services.SettingsService
}

func NewVideoHandler(store *services.VideoStore, downloader *services.DownloaderService, ytdlp *services.YtdlpService, settings *services.SettingsService) *VideoHandler {
	return &VideoHandler{Store: store, Downloader: downloader, Ytdlp: ytdlp, Settings: settings}
}

type MetadataRequest struct {
	URL string `json:"url"`
}

// GetMetadata godoc
// @Summary Get video metadata and format options
// @ID getMetadata
// @Tags videos
// @Accept json
// @Produce json
// @Param request body MetadataRequest true "Video URL"
// @Success 200 {object} services.VideoMetadata
// @Failure 400 {object} utils.ErrorResponse
// @Failure 502 {object} utils.ErrorResponse
// @Router /api/videos/metadata [post]
func (h *VideoHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	var req MetadataRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	url, ok := sanitizeURL(w, req.URL)
	if !ok {
		return
	}
	metadata, err := h.Ytdlp.GetMetadata(r.Context(), url)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadGateway, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, metadata)
}

type CreateVideoRequest struct {
	Name        string           `json:"name"`
	URL         string           `json:"url"`
	SourceTitle string           `json:"sourceTitle,omitempty"`
	FormatID    string           `json:"formatId,omitempty"`
	Encoding    encoding.Request `json:"encoding"`
}

// CreateVideo godoc
// @Summary Download a video with chosen format and encoding
// @ID createVideo
// @Tags videos
// @Accept json
// @Produce json
// @Param video body CreateVideoRequest true "Download options"
// @Success 201 {object} services.VideoDTO
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/videos [post]
func (h *VideoHandler) CreateVideo(w http.ResponseWriter, r *http.Request) {
	var req CreateVideoRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	if err := req.Encoding.Validate(h.Downloader.Capabilities()); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = utils.RandomName()
	}
	var sourceTitle *string
	if req.SourceTitle != "" {
		sourceTitle = &req.SourceTitle
	}
	h.startDownload(w, r, name, sourceTitle, req.URL, req.FormatID, req.Encoding)
}

type QuickDownloadRequest struct {
	URL string `json:"url"`
}

// QuickDownload godoc
// @Summary Start a download immediately with default settings and a random name
// @ID quickDownload
// @Tags videos
// @Accept json
// @Produce json
// @Param request body QuickDownloadRequest true "Video URL"
// @Success 201 {object} services.VideoDTO
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/videos/quick [post]
func (h *VideoHandler) QuickDownload(w http.ResponseWriter, r *http.Request) {
	var req QuickDownloadRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	settings := h.Settings.MustGet(r.Context())
	h.startDownload(w, r, utils.RandomName(), nil, req.URL, "", settings.DefaultEncoding)
}

func (h *VideoHandler) startDownload(w http.ResponseWriter, r *http.Request, name string, sourceTitle *string, rawURL, formatID string, req encoding.Request) {
	url, ok := sanitizeURL(w, rawURL)
	if !ok {
		return
	}
	video, err := h.Store.Create(r.Context(), name, sourceTitle, url)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if _, err := h.Downloader.StartDownload(r.Context(), video, url, formatID, req); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	video, err = h.Store.Get(r.Context(), video.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, services.ToVideoDTO(video))
}

// ListProgress godoc
// @Summary List progress of all running jobs
// @ID listProgress
// @Tags videos
// @Produce json
// @Success 200 {array} services.Progress
// @Router /api/videos/progress [get]
func (h *VideoHandler) ListProgress(w http.ResponseWriter, r *http.Request) {
	progress := h.Downloader.All()
	if progress == nil {
		progress = []services.Progress{}
	}
	utils.RespondWithJSON(w, http.StatusOK, progress)
}

// GetVideo godoc
// @Summary Get a video with all versions
// @ID getVideo
// @Tags videos
// @Produce json
// @Param id path string true "Video ID"
// @Success 200 {object} services.VideoDTO
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/videos/{id} [get]
func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {
	video, ok := h.video(w, r)
	if !ok {
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, services.ToVideoDTO(video))
}

type PaginatedVideoResponse struct {
	TotalCount  int64               `json:"totalCount"`
	TotalPages  int                 `json:"totalPages"`
	CurrentPage int                 `json:"currentPage"`
	Limit       int                 `json:"limit"`
	Videos      []services.VideoDTO `json:"videos"`
}

// ListVideos godoc
// @Summary List videos
// @Description Paginated list with optional search. The newest videos are served from memory.
// @ID listVideos
// @Tags videos
// @Produce json
// @Param search query string false "Search by name, title or URL"
// @Param order query string false "name_asc, name_desc, created_at_asc, created_at_desc"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 12, max: 100)"
// @Success 200 {object} PaginatedVideoResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/videos [get]
func (h *VideoHandler) ListVideos(w http.ResponseWriter, r *http.Request) {
	page, limit := utils.Pagination(r, 12)
	videos, total, err := h.Store.List(r.Context(), services.ListQuery{
		Search: strings.TrimSpace(r.URL.Query().Get("search")),
		Order:  r.URL.Query().Get("order"),
		Offset: (page - 1) * limit,
		Limit:  limit,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	dtos := make([]services.VideoDTO, len(videos))
	for i, v := range videos {
		dtos[i] = services.ToVideoDTO(v)
	}
	utils.RespondWithJSON(w, http.StatusOK, PaginatedVideoResponse{
		TotalCount:  total,
		TotalPages:  int((total + int64(limit) - 1) / int64(limit)),
		CurrentPage: page,
		Limit:       limit,
		Videos:      dtos,
	})
}

type UpdateVideoRequest struct {
	Name string `json:"name"`
}

// UpdateVideo godoc
// @Summary Rename a video
// @Description Renaming is instant and works while the video is still downloading.
// @ID updateVideo
// @Tags videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID"
// @Param video body UpdateVideoRequest true "New name"
// @Success 200 {object} services.VideoDTO
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/videos/{id} [put]
func (h *VideoHandler) UpdateVideo(w http.ResponseWriter, r *http.Request) {
	var req UpdateVideoRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 300 {
		utils.RespondWithError(w, http.StatusBadRequest, "name must be between 1 and 300 characters")
		return
	}
	video, err := h.Store.Rename(r.Context(), chi.URLParam(r, "id"), name)
	if respondStoreError(w, err) {
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, services.ToVideoDTO(video))
}

// DeleteVideo godoc
// @Summary Delete a video and all its versions
// @ID deleteVideo
// @Tags videos
// @Param id path string true "Video ID"
// @Success 204
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/videos/{id} [delete]
func (h *VideoHandler) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	_, err := h.Downloader.DeleteVideo(r.Context(), chi.URLParam(r, "id"))
	if respondStoreError(w, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetThumbnail godoc
// @Summary Get a video's thumbnail
// @ID getThumbnail
// @Tags videos
// @Produce jpeg
// @Param id path string true "Video ID"
// @Success 200 {file} binary
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/videos/{id}/thumbnail [get]
func (h *VideoHandler) GetThumbnail(w http.ResponseWriter, r *http.Request) {
	video, ok := h.video(w, r)
	if !ok {
		return
	}
	if video.ThumbnailFileName == nil {
		utils.RespondWithError(w, http.StatusNotFound, "no thumbnail")
		return
	}
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	http.ServeFile(w, r, h.Downloader.Path(*video.ThumbnailFileName))
}

type CreateVersionRequest struct {
	SourceFileID string           `json:"sourceFileId,omitempty"`
	Encoding     encoding.Request `json:"encoding"`
	MakePrimary  *bool            `json:"makePrimary,omitempty"`
}

// CreateVersion godoc
// @Summary Re-encode a video into a new version
// @Description Encodes from the given version, or from the original (else the default) when omitted.
// @ID createVersion
// @Tags versions
// @Accept json
// @Produce json
// @Param id path string true "Video ID"
// @Param request body CreateVersionRequest true "Encoding options"
// @Success 201 {object} services.VideoFileDTO
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Router /api/videos/{id}/files [post]
func (h *VideoHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	var req CreateVersionRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	video, ok := h.video(w, r)
	if !ok {
		return
	}
	source, ok := sourceVersion(video, req.SourceFileID)
	if !ok {
		utils.RespondWithError(w, http.StatusNotFound, "source version not found")
		return
	}
	makePrimary := req.MakePrimary == nil || *req.MakePrimary
	file, err := h.Downloader.StartEncode(r.Context(), video, source, req.Encoding, makePrimary)
	switch {
	case errors.Is(err, services.ErrNothingToDo), errors.Is(err, services.ErrFileNotReady):
		utils.RespondWithError(w, http.StatusConflict, err.Error())
	case err != nil:
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
	default:
		utils.RespondWithJSON(w, http.StatusCreated, services.ToFileDTO(file))
	}
}

type UpdateVersionRequest struct {
	Primary bool `json:"primary"`
}

// UpdateVersion godoc
// @Summary Make a version the default one
// @ID updateVersion
// @Tags versions
// @Accept json
// @Produce json
// @Param id path string true "Video ID"
// @Param fileId path string true "Version ID"
// @Param request body UpdateVersionRequest true "Changes"
// @Success 200 {object} services.VideoDTO
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Router /api/videos/{id}/files/{fileId} [patch]
func (h *VideoHandler) UpdateVersion(w http.ResponseWriter, r *http.Request) {
	var req UpdateVersionRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	video, file, ok := h.version(w, r)
	if !ok {
		return
	}
	if req.Primary {
		if file.Status != services.FileCompleted {
			utils.RespondWithError(w, http.StatusConflict, "only finished versions can be the default")
			return
		}
		var err error
		video, err = h.Store.SetPrimary(r.Context(), video.ID, &file.ID)
		if respondStoreError(w, err) {
			return
		}
	}
	utils.RespondWithJSON(w, http.StatusOK, services.ToVideoDTO(video))
}

// DeleteVersion godoc
// @Summary Delete a version
// @Description The last remaining version cannot be deleted; delete the video instead.
// @ID deleteVersion
// @Tags versions
// @Produce json
// @Param id path string true "Video ID"
// @Param fileId path string true "Version ID"
// @Success 200 {object} services.VideoDTO
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Router /api/videos/{id}/files/{fileId} [delete]
func (h *VideoHandler) DeleteVersion(w http.ResponseWriter, r *http.Request) {
	video, file, ok := h.version(w, r)
	if !ok {
		return
	}
	if len(video.Files) <= 1 {
		utils.RespondWithError(w, http.StatusConflict, "cannot delete the only version; delete the video instead")
		return
	}
	video, err := h.Downloader.DeleteFile(r.Context(), video.ID, file.ID)
	if respondStoreError(w, err) {
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, services.ToVideoDTO(video))
}

// CancelVersion godoc
// @Summary Cancel a queued or running download/encode
// @ID cancelVersion
// @Tags versions
// @Param id path string true "Video ID"
// @Param fileId path string true "Version ID"
// @Success 204
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/videos/{id}/files/{fileId}/cancel [post]
func (h *VideoHandler) CancelVersion(w http.ResponseWriter, r *http.Request) {
	if _, file, ok := h.version(w, r); ok {
		if !h.Downloader.Cancel(file.ID) {
			utils.RespondWithError(w, http.StatusNotFound, "no running job for this version")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// GetFile godoc
// @Summary Stream or download a version
// @Description Supports range requests. With download=1 the file is sent as an attachment named after the video.
// @ID getFile
// @Tags versions
// @Produce octet-stream
// @Param fileId path string true "Version ID"
// @Param download query bool false "Send as attachment"
// @Success 200 {file} binary
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/files/{fileId} [get]
func (h *VideoHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	video, file, err := h.Store.GetByFile(r.Context(), chi.URLParam(r, "fileId"))
	if respondStoreError(w, err) {
		return
	}
	if file.Status != services.FileCompleted || file.FileName == nil {
		utils.RespondWithError(w, http.StatusNotFound, "file is not ready")
		return
	}
	path := h.Downloader.Path(*file.FileName)
	f, err := os.Open(path)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "file missing on disk")
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ext := filepath.Ext(*file.FileName)
	name := utils.SanitizeFilename(video.Name)
	if len(video.Files) > 1 {
		name = utils.SanitizeFilename(fmt.Sprintf("%s (%s)", video.Name, strings.ReplaceAll(file.Label, " · ", " ")))
	}
	disposition := "inline"
	if r.URL.Query().Get("download") == "1" {
		disposition = "attachment"
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": name + ext}))
	w.Header().Set("Cache-Control", "private, max-age=86400")
	http.ServeContent(w, r, name+ext, info.ModTime(), f)
}

func (h *VideoHandler) video(w http.ResponseWriter, r *http.Request) (services.Video, bool) {
	video, err := h.Store.Get(r.Context(), chi.URLParam(r, "id"))
	return video, !respondStoreError(w, err)
}

func (h *VideoHandler) version(w http.ResponseWriter, r *http.Request) (services.Video, services.VideoFile, bool) {
	video, ok := h.video(w, r)
	if !ok {
		return video, services.VideoFile{}, false
	}
	file, ok := video.File(chi.URLParam(r, "fileId"))
	if !ok {
		utils.RespondWithError(w, http.StatusNotFound, "version not found")
	}
	return video, file, ok
}

func sourceVersion(v services.Video, id string) (services.VideoFile, bool) {
	if id != "" {
		return v.File(id)
	}
	if f, ok := v.Original(); ok {
		return f, true
	}
	return v.Primary()
}

func sanitizeURL(w http.ResponseWriter, raw string) (string, bool) {
	url, err := utils.SanitizeURL(strings.TrimSpace(raw))
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid URL")
		return "", false
	}
	return url, true
}

func respondStoreError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, services.ErrNotFound):
		utils.RespondWithError(w, http.StatusNotFound, "not found")
	default:
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
	return true
}
