package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azmekk/Vidra/backend/services/encoding"
)

type VideoOption struct {
	FormatID      string  `json:"formatId"`
	Extension     string  `json:"extension"`
	Resolution    string  `json:"resolution"`
	Width         int     `json:"width,omitempty"`
	Height        int     `json:"height,omitempty"`
	Fps           float64 `json:"fps,omitempty"`
	Note          string  `json:"note"`
	FileSize      int64   `json:"fileSize,omitempty"`
	Bitrate       float64 `json:"bitrate,omitempty"`
	VideoCodec    string  `json:"videoCodec"`
	AudioCodec    string  `json:"audioCodec"`
	DynamicRange  string  `json:"dynamicRange,omitempty"`
	HasAudio      bool    `json:"hasAudio"`
	IOSCompatible bool    `json:"iosCompatible"`
}

type VideoMetadata struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Uploader    string        `json:"uploader,omitempty"`
	Duration    float64       `json:"duration"`
	Thumbnail   string        `json:"thumbnail"`
	Extractor   string        `json:"extractor"`
	Options     []VideoOption `json:"options"`
}

type ytdlpFormat struct {
	FormatID       string  `json:"format_id"`
	Ext            string  `json:"ext"`
	Resolution     string  `json:"resolution"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	Fps            float64 `json:"fps"`
	FormatNote     string  `json:"format_note"`
	FileSize       int64   `json:"filesize"`
	FileSizeApprox int64   `json:"filesize_approx"`
	Tbr            float64 `json:"tbr"`
	VCodec         string  `json:"vcodec"`
	ACodec         string  `json:"acodec"`
	DynamicRange   string  `json:"dynamic_range"`
}

type ytdlpInfo struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Uploader    string        `json:"uploader"`
	Duration    float64       `json:"duration"`
	Thumbnail   string        `json:"thumbnail"`
	Extractor   string        `json:"extractor_key"`
	Formats     []ytdlpFormat `json:"formats"`
}

func (s *YtdlpService) GetMetadata(ctx context.Context, url string) (*VideoMetadata, error) {
	cmd := s.MetadataCommand(ctx, url)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	var info ytdlpInfo
	found := false
	for line := range strings.Lines(string(output)) {
		if strings.HasPrefix(strings.TrimSpace(line), "{") && json.Unmarshal([]byte(line), &info) == nil {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("failed to find valid JSON in yt-dlp output")
	}

	meta := &VideoMetadata{
		Title:       info.Title,
		Description: info.Description,
		Uploader:    info.Uploader,
		Duration:    info.Duration,
		Thumbnail:   info.Thumbnail,
		Extractor:   info.Extractor,
		Options:     make([]VideoOption, 0, len(info.Formats)),
	}
	for _, f := range info.Formats {
		if f.VCodec == "none" || (f.VCodec == "" && f.Height == 0) {
			continue
		}
		hasAudio := f.ACodec != "" && f.ACodec != "none"
		family := encoding.Normalize(f.VCodec)
		meta.Options = append(meta.Options, VideoOption{
			FormatID:      f.FormatID,
			Extension:     f.Ext,
			Resolution:    f.Resolution,
			Width:         f.Width,
			Height:        f.Height,
			Fps:           f.Fps,
			Note:          f.FormatNote,
			FileSize:      max(f.FileSize, f.FileSizeApprox),
			Bitrate:       f.Tbr,
			VideoCodec:    f.VCodec,
			AudioCodec:    f.ACodec,
			DynamicRange:  f.DynamicRange,
			HasAudio:      hasAudio,
			IOSCompatible: family == encoding.FamilyH264 || family == encoding.FamilyHEVC,
		})
	}
	return meta, nil
}
