package encoding

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Mode string

const (
	ModeRemux  Mode = "remux"
	ModeEncode Mode = "encode"
)

// Profile is a fully resolved set of ffmpeg choices for one output version.
type Profile struct {
	Mode         Mode    `json:"mode"`
	VideoEncoder string  `json:"videoEncoder,omitempty"`
	Level        Level   `json:"level,omitempty"`
	Quality      *int    `json:"quality,omitempty"`
	VideoBitrate int     `json:"videoBitrate,omitempty"`
	Preset       string  `json:"preset,omitempty"`
	MaxHeight    int     `json:"maxHeight,omitempty"`
	MaxFps       float64 `json:"maxFps,omitempty"`
	Container    string  `json:"container"`
	AudioEncoder string  `json:"audioEncoder,omitempty"`
	AudioBitrate int     `json:"audioBitrate,omitempty"`
}

// Source describes the media being encoded, from ffprobe or yt-dlp format info.
type Source struct {
	VideoCodec string  `json:"videoCodec"`
	AudioCodec string  `json:"audioCodec"`
	Container  string  `json:"container"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Fps        float64 `json:"fps"`
	Bitrate    int64   `json:"bitrate"`
	Duration   float64 `json:"duration"`
}

func (p Profile) Validate(caps *Capabilities) error {
	if !slices.Contains(caps.Containers, p.Container) {
		return fmt.Errorf("unsupported container %q", p.Container)
	}
	if p.Mode == ModeRemux {
		return nil
	}
	if p.Mode != ModeEncode {
		return fmt.Errorf("unsupported mode %q", p.Mode)
	}
	v, ok := caps.FindVideo(p.VideoEncoder)
	if !ok {
		return fmt.Errorf("video encoder %q is not available", p.VideoEncoder)
	}
	if v.Curated && !slices.Contains(v.Containers, p.Container) {
		return fmt.Errorf("%s cannot be stored in %s", v.Label, p.Container)
	}
	if p.Preset != "" && !slices.Contains(v.Presets, p.Preset) {
		return fmt.Errorf("invalid preset %q for %s", p.Preset, v.Label)
	}
	if p.Quality != nil && v.Quality != nil && (*p.Quality < v.Quality.Min || *p.Quality > v.Quality.Max) {
		return fmt.Errorf("quality must be between %d and %d", v.Quality.Min, v.Quality.Max)
	}
	if p.AudioEncoder != "" {
		a, ok := caps.FindAudio(p.AudioEncoder)
		if !ok {
			return fmt.Errorf("audio encoder %q is not available", p.AudioEncoder)
		}
		if a.Curated && !slices.Contains(a.Containers, p.Container) {
			return fmt.Errorf("%s audio cannot be stored in %s", a.Label, p.Container)
		}
	}
	if p.MaxHeight < 0 || p.MaxFps < 0 || p.VideoBitrate < 0 || p.AudioBitrate < 0 {
		return fmt.Errorf("limits must be positive")
	}
	return nil
}

// OutputHeight is the height after applying MaxHeight, never upscaling.
func (p Profile) OutputHeight(src Source) int {
	if p.MaxHeight > 0 && (src.Height == 0 || p.MaxHeight < src.Height) {
		return p.MaxHeight
	}
	return src.Height
}

// Args builds the ffmpeg arguments for encoding input into output.
// Progress is written to stdout in -progress format.
func (p Profile) Args(caps *Capabilities, src Source, input, output string) ([]string, error) {
	if err := p.Validate(caps); err != nil {
		return nil, err
	}

	var args []string
	video := copyVideo
	if p.Mode == ModeEncode {
		v, _ := caps.FindVideo(p.VideoEncoder)
		video = v.VideoEncoder
	}
	args = append(args, "-hide_banner", "-nostats", "-loglevel", "error")
	args = append(args, video.InitArgs...)
	args = append(args, "-i", input, "-map", "0:v:0", "-map", "0:a:0?", "-map_metadata", "0")

	if p.Mode == ModeRemux || video.Name == copyVideo.Name {
		args = append(args, "-c:v", "copy")
	} else {
		args = append(args, p.videoArgs(video, src)...)
	}

	switch {
	case p.Mode == ModeRemux || p.AudioEncoder == "" || p.AudioEncoder == copyAudio.Name:
		args = append(args, "-c:a", "copy")
	default:
		args = append(args, "-c:a", p.AudioEncoder)
		a, _ := caps.FindAudio(p.AudioEncoder)
		if bitrate := cmp.Or(p.AudioBitrate, a.DefaultBitrate); bitrate > 0 && !a.Lossless {
			args = append(args, "-b:a", strconv.Itoa(bitrate)+"k")
		}
	}

	if p.Container == "mp4" || p.Container == "mov" {
		args = append(args, "-movflags", "+faststart")
		if p.outputVideoFamily(caps, src) == FamilyHEVC {
			args = append(args, "-tag:v", "hvc1")
		}
	}

	args = append(args, "-progress", "pipe:1", "-y", output)
	return args, nil
}

func (p Profile) videoArgs(v VideoEncoder, src Source) []string {
	var filters []string
	if h := p.OutputHeight(src); h > 0 && h != src.Height {
		filters = append(filters, fmt.Sprintf("scale=-2:%d", h))
	}
	if p.MaxFps > 0 && (src.Fps == 0 || p.MaxFps < src.Fps) {
		filters = append(filters, "fps="+strconv.FormatFloat(p.MaxFps, 'f', -1, 64))
	}
	if v.Filter != "" {
		filters = append(filters, v.Filter)
	}

	args := []string{}
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}
	args = append(args, "-c:v", v.Name)

	switch {
	case v.Quality != nil:
		q := v.Quality.valueFor(cmp.Or(p.Level, LevelBalanced), p.OutputHeight(src))
		if p.Quality != nil {
			q = *p.Quality
		}
		for _, flag := range v.Quality.Args {
			args = append(args, flag, strconv.Itoa(q))
		}
		if v.Quality.Args[0] == "-qp_i" {
			args = append(args, "-qp_p", strconv.Itoa(q))
		}
	case p.VideoBitrate > 0:
		args = append(args, "-b:v", strconv.Itoa(p.VideoBitrate)+"k")
	}

	if v.PresetFlag != "" {
		args = append(args, v.PresetFlag, cmp.Or(p.Preset, v.DefaultPreset))
	}
	return append(args, v.ExtraArgs...)
}

// Extension is the file extension for the profile's container.
func (p Profile) Extension() string { return "." + p.Container }

// Label is a short human description such as "HEVC · 1080p".
func (p Profile) Label(caps *Capabilities, src Source) string {
	res := ""
	if h := p.OutputHeight(src); h > 0 {
		res = " · " + strconv.Itoa(h) + "p"
	}
	if p.Mode == ModeRemux || p.VideoEncoder == copyVideo.Name {
		return "Remux · " + strings.ToUpper(p.Container) + res
	}
	name := p.VideoEncoder
	if v, ok := caps.FindVideo(p.VideoEncoder); ok {
		name = familyLabel(v.Family, v.Label)
	}
	return name + res
}

// IOSCompatible reports whether the output can be saved to Photos on iOS.
func (p Profile) IOSCompatible(caps *Capabilities, src Source) bool {
	return IsIOSCompatible(p.outputVideoFamily(caps, src), p.outputAudio(src), p.Container)
}

func (p Profile) outputVideoFamily(caps *Capabilities, src Source) Family {
	if p.Mode == ModeRemux || p.VideoEncoder == copyVideo.Name {
		return Normalize(src.VideoCodec)
	}
	v, _ := caps.FindVideo(p.VideoEncoder)
	return v.Family
}

func (p Profile) outputAudio(src Source) string {
	if p.Mode == ModeRemux || p.AudioEncoder == "" || p.AudioEncoder == copyAudio.Name {
		return NormalizeAudio(src.AudioCodec)
	}
	return NormalizeAudio(p.AudioEncoder)
}

// IsIOSCompatible checks codec/container combos that iOS Photos accepts.
func IsIOSCompatible(video Family, audio, container string) bool {
	if video != FamilyH264 && video != FamilyHEVC {
		return false
	}
	if container != "mp4" && container != "mov" {
		return false
	}
	return audio == "" || audio == "aac" || audio == "mp3" || audio == "alac"
}

// Normalize maps ffprobe codec names and RFC 6381 codec strings to a family.
func Normalize(codec string) Family {
	c := strings.ToLower(codec)
	switch {
	case c == "", c == "none":
		return ""
	case c == "h264", strings.HasPrefix(c, "avc"):
		return FamilyH264
	case c == "hevc", c == "h265", strings.HasPrefix(c, "hev1"), strings.HasPrefix(c, "hvc1"):
		return FamilyHEVC
	case c == "av1", strings.HasPrefix(c, "av01"):
		return FamilyAV1
	case c == "vp9", strings.HasPrefix(c, "vp09"), strings.HasPrefix(c, "vp9"):
		return FamilyVP9
	}
	return FamilyAny
}

func NormalizeAudio(codec string) string {
	c := strings.ToLower(codec)
	switch {
	case c == "", c == "none":
		return ""
	case c == "aac", c == "libfdk_aac", strings.HasPrefix(c, "mp4a"):
		return "aac"
	case c == "opus", c == "libopus":
		return "opus"
	case c == "mp3", c == "libmp3lame":
		return "mp3"
	}
	return c
}

func familyLabel(f Family, fallback string) string {
	switch f {
	case FamilyH264:
		return "H.264"
	case FamilyHEVC:
		return "HEVC"
	case FamilyAV1:
		return "AV1"
	case FamilyVP9:
		return "VP9"
	}
	return fallback
}
