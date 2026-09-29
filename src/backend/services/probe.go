package services

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/services/encoding"
)

type ffprobeOutput struct {
	Streams []struct {
		CodecType  string `json:"codec_type"`
		CodecName  string `json:"codec_name"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
		BitRate    string `json:"bit_rate"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
		Size     string `json:"size"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
}

type Probe struct {
	encoding.Source
	Size int64
}

// ProbeFile reads codec, resolution and bitrate information with ffprobe.
func ProbeFile(ctx context.Context, path string) (Probe, error) {
	out, err := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-print_format", "json",
		"-show_format", "-show_streams", path).Output()
	if err != nil {
		return Probe{}, err
	}
	var raw ffprobeOutput
	if err := json.Unmarshal(out, &raw); err != nil {
		return Probe{}, err
	}

	p := Probe{Source: encoding.Source{Container: containerOf(path)}}
	p.Duration, _ = strconv.ParseFloat(raw.Format.Duration, 64)
	p.Size, _ = strconv.ParseInt(raw.Format.Size, 10, 64)
	formatBitrate, _ := strconv.ParseInt(raw.Format.BitRate, 10, 64)

	for _, s := range raw.Streams {
		switch s.CodecType {
		case "video":
			if p.VideoCodec != "" || s.CodecName == "mjpeg" || s.CodecName == "png" {
				continue
			}
			p.VideoCodec, p.Width, p.Height = s.CodecName, s.Width, s.Height
			p.Fps = parseRate(s.RFrameRate)
			p.Bitrate, _ = strconv.ParseInt(s.BitRate, 10, 64)
		case "audio":
			if p.AudioCodec == "" {
				p.AudioCodec = s.CodecName
			}
		}
	}
	if p.Bitrate == 0 {
		p.Bitrate = formatBitrate
	}
	return p, nil
}

func (p Probe) IOSCompatible() bool {
	return encoding.IsIOSCompatible(encoding.Normalize(p.VideoCodec), encoding.NormalizeAudio(p.AudioCodec), p.Container)
}

func (p Probe) mediaParams(fileID, fileName, label string) database.UpdateFileMediaParams {
	return database.UpdateFileMediaParams{
		ID:            fileID,
		FileName:      &fileName,
		Label:         label,
		Status:        FileCompleted,
		Container:     nonEmpty(p.Container),
		VideoCodec:    nonEmpty(p.VideoCodec),
		AudioCodec:    nonEmpty(p.AudioCodec),
		Width:         positive(int64(p.Width)),
		Height:        positive(int64(p.Height)),
		Fps:           positive(p.Fps),
		Bitrate:       positive(p.Bitrate),
		FileSize:      positive(p.Size),
		IosCompatible: p.IOSCompatible(),
	}
}

// SourceOf converts a stored file version back into encoder input information.
func SourceOf(f database.VideoFile) encoding.Source {
	return encoding.Source{
		VideoCodec: deref(f.VideoCodec),
		AudioCodec: deref(f.AudioCodec),
		Container:  deref(f.Container),
		Width:      int(deref(f.Width)),
		Height:     int(deref(f.Height)),
		Fps:        deref(f.Fps),
		Bitrate:    deref(f.Bitrate),
	}
}

func containerOf(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if ext == "m4v" {
		return "mp4"
	}
	return ext
}

func parseRate(r string) float64 {
	num, den, ok := strings.Cut(r, "/")
	n, _ := strconv.ParseFloat(num, 64)
	if !ok {
		return n
	}
	d, _ := strconv.ParseFloat(den, 64)
	if d == 0 {
		return 0
	}
	return n / d
}

func nonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func positive[T int64 | float64](v T) *T {
	if v <= 0 {
		return nil
	}
	return &v
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
