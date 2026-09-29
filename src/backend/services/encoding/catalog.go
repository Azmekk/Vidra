package encoding

type Family string

const (
	FamilyH264 Family = "h264"
	FamilyHEVC Family = "hevc"
	FamilyAV1  Family = "av1"
	FamilyVP9  Family = "vp9"
	FamilyCopy Family = "copy"
	FamilyAny  Family = "other"
)

type Level string

const (
	LevelVisuallyLossless Level = "visually_lossless"
	LevelHigh             Level = "high"
	LevelBalanced         Level = "balanced"
	LevelCompact          Level = "compact"
	LevelTiny             Level = "tiny"
)

var Levels = []Level{LevelVisuallyLossless, LevelHigh, LevelBalanced, LevelCompact, LevelTiny}

// QualityParam describes an encoder's native quality control. Levels holds
// values tuned for 1080p; other resolutions are shifted by heightOffset.
type QualityParam struct {
	Args          []string      `json:"-"`
	Min           int           `json:"min"`
	Max           int           `json:"max"`
	LowerIsBetter bool          `json:"lowerIsBetter"`
	Levels        map[Level]int `json:"levels"`
}

type VideoEncoder struct {
	Name          string        `json:"name"`
	Label         string        `json:"label"`
	Description   string        `json:"description"`
	Family        Family        `json:"family"`
	HWType        string        `json:"hwType,omitempty"`
	Quality       *QualityParam `json:"quality,omitempty"`
	PresetFlag    string        `json:"-"`
	Presets       []string      `json:"presets,omitempty"`
	DefaultPreset string        `json:"defaultPreset,omitempty"`
	FastPreset    string        `json:"-"`
	SlowPreset    string        `json:"-"`
	Containers    []string      `json:"containers"`
	ExtraArgs     []string      `json:"-"`
	InitArgs      []string      `json:"-"`
	Filter        string        `json:"-"`
	TenBit        bool          `json:"-"`
}

type AudioEncoder struct {
	Name           string   `json:"name"`
	Label          string   `json:"label"`
	Description    string   `json:"description"`
	Containers     []string `json:"containers"`
	DefaultBitrate int      `json:"defaultBitrate,omitempty"`
	Lossless       bool     `json:"lossless,omitempty"`
	IOS            bool     `json:"ios"`
}

func (e VideoEncoder) Hardware() bool { return e.HWType != "" }

func (e VideoEncoder) IOS() bool { return e.Family == FamilyH264 || e.Family == FamilyHEVC }

var (
	allContainers   = []string{"mp4", "mkv", "mov"}
	modernContainer = []string{"mp4", "mkv", "webm"}
	vp9Containers   = []string{"webm", "mkv", "mp4"}
	x264Presets     = []string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow"}
	qsvPresets      = []string{"veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow"}
	nvencPresets    = []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}
	amfPresets      = []string{"speed", "balanced", "quality"}
)

func levels(v ...int) map[Level]int {
	m := make(map[Level]int, len(Levels))
	for i, l := range Levels {
		m[l] = v[i]
	}
	return m
}

func numbered(from, to int) []string {
	out := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		out = append(out, itoa(i))
	}
	return out
}

func crf(min, max int, lv ...int) *QualityParam {
	return &QualityParam{Args: []string{"-crf"}, Min: min, Max: max, LowerIsBetter: true, Levels: levels(lv...)}
}

func param(flag string, min, max int, lowerIsBetter bool, lv ...int) *QualityParam {
	return &QualityParam{Args: []string{flag}, Min: min, Max: max, LowerIsBetter: lowerIsBetter, Levels: levels(lv...)}
}

var vaapiInit = []string{"-vaapi_device", "/dev/dri/renderD128"}

var videoCatalog = []VideoEncoder{
	{Name: "libx264", Label: "H.264 (x264)", Description: "Plays everywhere. The safest choice.", Family: FamilyH264,
		Quality: crf(0, 51, 17, 20, 23, 26, 30), PresetFlag: "-preset", Presets: x264Presets,
		DefaultPreset: "medium", FastPreset: "veryfast", SlowPreset: "slow", Containers: allContainers,
		ExtraArgs: []string{"-pix_fmt", "yuv420p"}},
	{Name: "libx265", Label: "HEVC (x265)", Description: "~40% smaller than H.264, plays on Apple devices.", Family: FamilyHEVC,
		Quality: crf(0, 51, 19, 23, 26, 29, 33), PresetFlag: "-preset", Presets: x264Presets,
		DefaultPreset: "medium", FastPreset: "veryfast", SlowPreset: "slow", Containers: allContainers,
		ExtraArgs: []string{"-pix_fmt", "yuv420p", "-x265-params", "log-level=error"}},
	{Name: "libsvtav1", Label: "AV1 (SVT-AV1)", Description: "Best compression. Newer devices only.", Family: FamilyAV1,
		Quality: crf(0, 63, 22, 28, 34, 40, 48), PresetFlag: "-preset", Presets: numbered(0, 13),
		DefaultPreset: "8", FastPreset: "11", SlowPreset: "5", Containers: modernContainer,
		ExtraArgs: []string{"-pix_fmt", "yuv420p10le"}, TenBit: true},
	{Name: "libaom-av1", Label: "AV1 (libaom)", Description: "Reference AV1 encoder. Very slow.", Family: FamilyAV1,
		Quality: crf(0, 63, 20, 26, 32, 38, 46), PresetFlag: "-cpu-used", Presets: numbered(0, 8),
		DefaultPreset: "6", FastPreset: "8", SlowPreset: "4", Containers: modernContainer,
		ExtraArgs: []string{"-b:v", "0", "-row-mt", "1"}},
	{Name: "librav1e", Label: "AV1 (rav1e)", Description: "Rust AV1 encoder.", Family: FamilyAV1,
		Quality: param("-qp", 0, 255, true, 60, 80, 100, 130, 170), PresetFlag: "-speed", Presets: numbered(0, 10),
		DefaultPreset: "6", FastPreset: "9", SlowPreset: "4", Containers: modernContainer},
	{Name: "libvpx-vp9", Label: "VP9 (libvpx)", Description: "Open codec, good for the web.", Family: FamilyVP9,
		Quality: crf(0, 63, 20, 26, 32, 38, 45), PresetFlag: "-cpu-used", Presets: numbered(0, 5),
		DefaultPreset: "4", FastPreset: "5", SlowPreset: "2", Containers: vp9Containers,
		ExtraArgs: []string{"-b:v", "0", "-deadline", "good", "-row-mt", "1"}},

	{Name: "h264_nvenc", Label: "H.264 (NVIDIA)", Description: "GPU accelerated, very fast.", Family: FamilyH264, HWType: "nvenc",
		Quality: param("-cq", 0, 51, true, 19, 23, 27, 31, 35), PresetFlag: "-preset", Presets: nvencPresets,
		DefaultPreset: "p5", FastPreset: "p2", SlowPreset: "p7", Containers: allContainers,
		ExtraArgs: []string{"-rc", "vbr", "-b:v", "0", "-pix_fmt", "yuv420p"}},
	{Name: "hevc_nvenc", Label: "HEVC (NVIDIA)", Description: "GPU accelerated HEVC.", Family: FamilyHEVC, HWType: "nvenc",
		Quality: param("-cq", 0, 51, true, 21, 25, 29, 33, 37), PresetFlag: "-preset", Presets: nvencPresets,
		DefaultPreset: "p5", FastPreset: "p2", SlowPreset: "p7", Containers: allContainers,
		ExtraArgs: []string{"-rc", "vbr", "-b:v", "0", "-pix_fmt", "yuv420p"}},
	{Name: "av1_nvenc", Label: "AV1 (NVIDIA)", Description: "GPU accelerated AV1 (RTX 40+).", Family: FamilyAV1, HWType: "nvenc",
		Quality: param("-cq", 0, 63, true, 25, 30, 36, 42, 48), PresetFlag: "-preset", Presets: nvencPresets,
		DefaultPreset: "p5", FastPreset: "p2", SlowPreset: "p7", Containers: modernContainer,
		ExtraArgs: []string{"-rc", "vbr", "-b:v", "0"}},

	{Name: "h264_qsv", Label: "H.264 (Intel QSV)", Description: "Intel GPU accelerated.", Family: FamilyH264, HWType: "qsv",
		Quality: param("-global_quality", 1, 51, true, 20, 23, 26, 29, 33), PresetFlag: "-preset", Presets: qsvPresets,
		DefaultPreset: "medium", FastPreset: "veryfast", SlowPreset: "slow", Containers: allContainers},
	{Name: "hevc_qsv", Label: "HEVC (Intel QSV)", Description: "Intel GPU accelerated HEVC.", Family: FamilyHEVC, HWType: "qsv",
		Quality: param("-global_quality", 1, 51, true, 22, 25, 28, 31, 35), PresetFlag: "-preset", Presets: qsvPresets,
		DefaultPreset: "medium", FastPreset: "veryfast", SlowPreset: "slow", Containers: allContainers},
	{Name: "av1_qsv", Label: "AV1 (Intel QSV)", Description: "Intel Arc AV1.", Family: FamilyAV1, HWType: "qsv",
		Quality: param("-global_quality", 1, 51, true, 24, 28, 32, 36, 40), PresetFlag: "-preset", Presets: qsvPresets,
		DefaultPreset: "medium", FastPreset: "veryfast", SlowPreset: "slow", Containers: modernContainer},
	{Name: "vp9_qsv", Label: "VP9 (Intel QSV)", Description: "Intel GPU accelerated VP9.", Family: FamilyVP9, HWType: "qsv",
		Quality: param("-global_quality", 1, 51, true, 20, 25, 30, 35, 40), PresetFlag: "-preset", Presets: qsvPresets,
		DefaultPreset: "medium", FastPreset: "veryfast", SlowPreset: "slow", Containers: vp9Containers},

	{Name: "h264_vaapi", Label: "H.264 (VAAPI)", Description: "Linux GPU accelerated.", Family: FamilyH264, HWType: "vaapi",
		Quality: param("-qp", 0, 51, true, 20, 23, 26, 29, 33), Containers: allContainers,
		InitArgs: vaapiInit, Filter: "format=nv12,hwupload"},
	{Name: "hevc_vaapi", Label: "HEVC (VAAPI)", Description: "Linux GPU accelerated HEVC.", Family: FamilyHEVC, HWType: "vaapi",
		Quality: param("-qp", 0, 51, true, 22, 25, 28, 31, 35), Containers: allContainers,
		InitArgs: vaapiInit, Filter: "format=nv12,hwupload"},
	{Name: "av1_vaapi", Label: "AV1 (VAAPI)", Description: "Linux GPU accelerated AV1.", Family: FamilyAV1, HWType: "vaapi",
		Quality: param("-qp", 0, 255, true, 80, 100, 120, 150, 180), Containers: modernContainer,
		InitArgs: vaapiInit, Filter: "format=nv12,hwupload"},

	{Name: "h264_videotoolbox", Label: "H.264 (Apple)", Description: "macOS hardware encoder.", Family: FamilyH264, HWType: "videotoolbox",
		Quality: param("-q:v", 1, 100, false, 75, 65, 55, 45, 35), Containers: allContainers},
	{Name: "hevc_videotoolbox", Label: "HEVC (Apple)", Description: "macOS hardware HEVC.", Family: FamilyHEVC, HWType: "videotoolbox",
		Quality: param("-q:v", 1, 100, false, 70, 60, 50, 40, 30), Containers: allContainers},

	{Name: "h264_amf", Label: "H.264 (AMD)", Description: "AMD GPU accelerated.", Family: FamilyH264, HWType: "amf",
		Quality: param("-qp_i", 0, 51, true, 20, 23, 26, 29, 33), PresetFlag: "-quality", Presets: amfPresets,
		DefaultPreset: "balanced", FastPreset: "speed", SlowPreset: "quality", Containers: allContainers,
		ExtraArgs: []string{"-rc", "cqp"}},
	{Name: "hevc_amf", Label: "HEVC (AMD)", Description: "AMD GPU accelerated HEVC.", Family: FamilyHEVC, HWType: "amf",
		Quality: param("-qp_i", 0, 51, true, 22, 25, 28, 31, 35), PresetFlag: "-quality", Presets: amfPresets,
		DefaultPreset: "balanced", FastPreset: "speed", SlowPreset: "quality", Containers: allContainers,
		ExtraArgs: []string{"-rc", "cqp"}},
	{Name: "av1_amf", Label: "AV1 (AMD)", Description: "AMD GPU accelerated AV1.", Family: FamilyAV1, HWType: "amf",
		Quality: param("-qp_i", 0, 255, true, 80, 100, 120, 150, 180), PresetFlag: "-quality", Presets: amfPresets,
		DefaultPreset: "balanced", FastPreset: "speed", SlowPreset: "quality", Containers: modernContainer,
		ExtraArgs: []string{"-rc", "cqp"}},
}

var copyVideo = VideoEncoder{Name: "copy", Label: "Keep video stream", Description: "No re-encode, no quality loss.",
	Family: FamilyCopy, Containers: []string{"mp4", "mkv", "mov", "webm"}}

var audioCatalog = []AudioEncoder{
	{Name: "aac", Label: "AAC", Description: "Universal compatibility.", Containers: []string{"mp4", "mkv", "mov"}, DefaultBitrate: 160, IOS: true},
	{Name: "libfdk_aac", Label: "AAC (Fraunhofer)", Description: "Higher quality AAC.", Containers: []string{"mp4", "mkv", "mov"}, DefaultBitrate: 128, IOS: true},
	{Name: "libopus", Label: "Opus", Description: "Best quality per bit.", Containers: []string{"mp4", "mkv", "webm"}, DefaultBitrate: 96},
	{Name: "libmp3lame", Label: "MP3", Description: "Legacy compatibility.", Containers: []string{"mp4", "mkv", "mov"}, DefaultBitrate: 192},
	{Name: "flac", Label: "FLAC", Description: "Lossless, large.", Containers: []string{"mp4", "mkv"}, Lossless: true},
}

var copyAudio = AudioEncoder{Name: "copy", Label: "Keep audio stream", Description: "No re-encode.",
	Containers: []string{"mp4", "mkv", "mov", "webm"}}

// valueFor returns the encoder value for a level, adjusted for output height.
func (q *QualityParam) valueFor(level Level, height int) int {
	v, ok := q.Levels[level]
	if !ok {
		v = q.Levels[LevelBalanced]
	}
	offset := heightOffset(height)
	if !q.LowerIsBetter {
		offset = -offset * 3
	} else if q.Max > 63 {
		offset *= 5
	}
	return clamp(v+offset, q.Min, q.Max)
}

// Lower resolutions need more bits per pixel, higher ones tolerate fewer.
func heightOffset(h int) int {
	switch {
	case h <= 0:
		return 0
	case h <= 480:
		return -2
	case h <= 720:
		return -1
	case h <= 1080:
		return 0
	case h <= 1440:
		return 1
	default:
		return 2
	}
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}
