package services

import (
	"encoding/json"
	"net/url"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/services/encoding"
)

type VideoDTO struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	SourceTitle   *string        `json:"sourceTitle,omitempty"`
	OriginalURL   string         `json:"originalUrl"`
	ThumbnailURL  *string        `json:"thumbnailUrl,omitempty"`
	Duration      *float64       `json:"duration,omitempty"`
	Uploader      *string        `json:"uploader,omitempty"`
	PrimaryFileID *string        `json:"primaryFileId,omitempty"`
	Status        string         `json:"status"`
	Files         []VideoFileDTO `json:"files"`
	CreatedAt     string         `json:"createdAt"`
	UpdatedAt     string         `json:"updatedAt"`
}

type VideoFileDTO struct {
	ID              string            `json:"id"`
	VideoID         string            `json:"videoId"`
	Kind            string            `json:"kind"`
	SourceFileID    *string           `json:"sourceFileId,omitempty"`
	Label           string            `json:"label"`
	Status          string            `json:"status"`
	Container       *string           `json:"container,omitempty"`
	VideoCodec      *string           `json:"videoCodec,omitempty"`
	AudioCodec      *string           `json:"audioCodec,omitempty"`
	Width           *int64            `json:"width,omitempty"`
	Height          *int64            `json:"height,omitempty"`
	Fps             *float64          `json:"fps,omitempty"`
	Bitrate         *int64            `json:"bitrate,omitempty"`
	FileSize        *int64            `json:"fileSize,omitempty"`
	IOSCompatible   bool              `json:"iosCompatible"`
	EncodingProfile *encoding.Profile `json:"encodingProfile,omitempty"`
	URL             *string           `json:"url,omitempty"`
	CreatedAt       string            `json:"createdAt"`
	UpdatedAt       string            `json:"updatedAt"`
}

func ToVideoDTO(v Video) VideoDTO {
	files := make([]VideoFileDTO, len(v.Files))
	for i, f := range v.Files {
		files[i] = ToFileDTO(f)
	}
	var thumb *string
	if v.ThumbnailFileName != nil {
		u := "/api/videos/" + v.ID + "/thumbnail?v=" + url.QueryEscape(*v.ThumbnailFileName)
		thumb = &u
	}
	return VideoDTO{
		ID:            v.ID,
		Name:          v.Name,
		SourceTitle:   v.SourceTitle,
		OriginalURL:   v.OriginalUrl,
		ThumbnailURL:  thumb,
		Duration:      v.Duration,
		Uploader:      v.Uploader,
		PrimaryFileID: v.PrimaryFileID,
		Status:        v.Status(),
		Files:         files,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
	}
}

func ToFileDTO(f database.VideoFile) VideoFileDTO {
	var profile *encoding.Profile
	if f.EncodingProfile != nil {
		var p encoding.Profile
		if json.Unmarshal([]byte(*f.EncodingProfile), &p) == nil {
			profile = &p
		}
	}
	var fileURL *string
	if f.Status == FileCompleted && f.FileName != nil {
		u := "/api/files/" + f.ID
		fileURL = &u
	}
	return VideoFileDTO{
		ID:              f.ID,
		VideoID:         f.VideoID,
		Kind:            f.Kind,
		SourceFileID:    f.SourceFileID,
		Label:           f.Label,
		Status:          f.Status,
		Container:       f.Container,
		VideoCodec:      f.VideoCodec,
		AudioCodec:      f.AudioCodec,
		Width:           f.Width,
		Height:          f.Height,
		Fps:             f.Fps,
		Bitrate:         f.Bitrate,
		FileSize:        f.FileSize,
		IOSCompatible:   f.IosCompatible,
		EncodingProfile: profile,
		URL:             fileURL,
		CreatedAt:       f.CreatedAt,
		UpdatedAt:       f.UpdatedAt,
	}
}
