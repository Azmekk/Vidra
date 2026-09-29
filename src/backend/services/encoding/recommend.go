package encoding

import (
	"fmt"
	"strings"
)

type Goal string

const (
	GoalOriginal   Goal = "original"
	GoalCompatible Goal = "compatible"
	GoalBalanced   Goal = "balanced"
	GoalSmallest   Goal = "smallest"
	GoalFastest    Goal = "fastest"
	GoalCustom     Goal = "custom"
)

var Goals = []Goal{GoalOriginal, GoalCompatible, GoalBalanced, GoalSmallest, GoalFastest, GoalCustom}

// Request is what a user asks for: a goal, or a custom profile.
type Request struct {
	Goal    Goal     `json:"goal"`
	Profile *Profile `json:"profile,omitempty"`
}

type Recommendation struct {
	Goal          Goal     `json:"goal"`
	Skip          bool     `json:"skip"`
	Profile       Profile  `json:"profile"`
	Label         string   `json:"label"`
	IOSCompatible bool     `json:"iosCompatible"`
	Reasons       []string `json:"reasons"`
}

func (r Request) Validate(caps *Capabilities) error {
	switch r.Goal {
	case GoalCustom:
		if r.Profile == nil {
			return fmt.Errorf("custom goal requires a profile")
		}
		return r.Profile.Validate(caps)
	case GoalOriginal, GoalCompatible, GoalBalanced, GoalSmallest, GoalFastest:
		return nil
	}
	return fmt.Errorf("unknown goal %q", r.Goal)
}

// Resolve turns a request into a concrete recommendation for a source.
func (r Request) Resolve(caps *Capabilities, src Source) Recommendation {
	if r.Goal == GoalCustom && r.Profile != nil {
		p := *r.Profile
		return finish(caps, src, GoalCustom, p, []string{"Using your custom settings."})
	}
	return Recommend(caps, r.Goal, src)
}

// Recommend picks encoding settings for a goal using simple, explainable rules.
func Recommend(caps *Capabilities, goal Goal, src Source) Recommendation {
	vf := Normalize(src.VideoCodec)
	af := NormalizeAudio(src.AudioCodec)
	efficient := vf == FamilyHEVC || vf == FamilyAV1 || vf == FamilyVP9
	sourceCompatible := IsIOSCompatible(vf, af, "mp4")
	bpp := bitsPerPixel(src)
	var reasons []string

	switch goal {
	case GoalOriginal:
		return Recommendation{Goal: goal, Skip: true, Label: "Original",
			IOSCompatible: IsIOSCompatible(vf, af, src.Container),
			Reasons:       []string{"Keeping the file exactly as downloaded."}}

	case GoalCompatible:
		if sourceCompatible {
			reasons = append(reasons, fmt.Sprintf("Source is already %s with %s audio, so it is only repackaged as MP4. No quality loss.", familyLabel(vf, vf.String()), strings.ToUpper(af)))
			return finish(caps, src, goal, Profile{Mode: ModeRemux, Container: "mp4"}, reasons)
		}
		if vf == FamilyH264 || vf == FamilyHEVC {
			reasons = append(reasons, "Video is already iPhone compatible and is kept as is; only the audio is converted to AAC.")
			return finish(caps, src, goal, Profile{Mode: ModeEncode, VideoEncoder: copyVideo.Name, Container: "mp4", AudioEncoder: "aac"}, reasons)
		}
		enc := pick(caps, "libx264", "h264_videotoolbox", "h264_nvenc", "h264_qsv")
		reasons = append(reasons, fmt.Sprintf("%s video does not play everywhere on iOS; converting to H.264, which every device supports.", familyLabel(vf, "This")))
		p := Profile{Mode: ModeEncode, VideoEncoder: enc.Name, Level: LevelHigh, Container: "mp4", AudioEncoder: audioFor(af, "aac")}
		p = capFps(p, src, 60, &reasons)
		return finish(caps, src, goal, p, reasons)

	case GoalBalanced:
		if efficient && bpp > 0 && bpp < 0.06 {
			reasons = append(reasons, fmt.Sprintf("Source is already efficiently compressed %s (%.3f bits/pixel); re-encoding would not make it meaningfully smaller.", familyLabel(vf, ""), bpp))
			return finish(caps, src, goal, Profile{Mode: ModeRemux, Container: containerFor(vf, af)}, reasons)
		}
		enc := pick(caps, "libx265", "hevc_videotoolbox", "hevc_nvenc", "hevc_qsv", "libx264")
		reasons = append(reasons, fmt.Sprintf("%s gives roughly 40%% smaller files than H.264 at the same quality and still plays on Apple devices.", familyLabel(enc.Family, enc.Label)))
		p := Profile{Mode: ModeEncode, VideoEncoder: enc.Name, Level: LevelBalanced, Container: "mp4", AudioEncoder: audioFor(af, "aac")}
		p = capFps(p, src, 60, &reasons)
		return finish(caps, src, goal, p, reasons)

	case GoalSmallest:
		enc := pick(caps, "libsvtav1", "av1_nvenc", "av1_qsv", "libx265", "hevc_nvenc", "libx264")
		reasons = append(reasons, fmt.Sprintf("%s is the most efficient available encoder.", familyLabel(enc.Family, enc.Label)))
		p := Profile{Mode: ModeEncode, VideoEncoder: enc.Name, Level: LevelCompact, Container: "mp4", AudioEncoder: "libopus", AudioBitrate: 96}
		if _, ok := caps.FindAudio("libopus"); !ok {
			p.AudioEncoder, p.AudioBitrate = "aac", 128
		}
		if src.Height == 0 || src.Height > 1080 {
			p.MaxHeight = 1080
			reasons = append(reasons, "Resolution is capped at 1080p, which looks sharp on phones and laptops.")
		}
		p = capFps(p, src, 30, &reasons)
		if enc.Family == FamilyAV1 {
			reasons = append(reasons, "AV1 needs a recent device (iPhone 15 Pro or newer) to play smoothly.")
		}
		return finish(caps, src, goal, p, reasons)

	case GoalFastest:
		if sourceCompatible {
			reasons = append(reasons, "Source already plays everywhere; repackaging is instant.")
			return finish(caps, src, goal, Profile{Mode: ModeRemux, Container: "mp4"}, reasons)
		}
		enc := pick(caps, "h264_nvenc", "h264_qsv", "h264_videotoolbox", "h264_vaapi", "h264_amf", "libx264")
		p := Profile{Mode: ModeEncode, VideoEncoder: enc.Name, Level: LevelHigh, Preset: enc.FastPreset, Container: "mp4", AudioEncoder: audioFor(af, "aac")}
		if enc.Hardware() {
			reasons = append(reasons, fmt.Sprintf("Using %s hardware encoding, many times faster than software.", enc.Label))
		} else {
			reasons = append(reasons, "No working hardware encoder was found; using x264 with a fast preset.")
		}
		return finish(caps, src, goal, p, reasons)
	}

	return Recommend(caps, GoalCompatible, src)
}

func finish(caps *Capabilities, src Source, goal Goal, p Profile, reasons []string) Recommendation {
	if p.Mode == ModeRemux && src.Container == p.Container {
		reasons = append(reasons, "The downloaded file already matches, so no new version is needed.")
		return Recommendation{Goal: goal, Skip: true, Profile: p, Label: "Original",
			IOSCompatible: p.IOSCompatible(caps, src), Reasons: reasons}
	}
	return Recommendation{
		Goal:          goal,
		Profile:       p,
		Label:         p.Label(caps, src),
		IOSCompatible: p.IOSCompatible(caps, src),
		Reasons:       reasons,
	}
}

func pick(caps *Capabilities, names ...string) VideoEncoder {
	if e, ok := caps.firstAvailable(names...); ok {
		return e.VideoEncoder
	}
	return copyVideo
}

func audioFor(source, target string) string {
	if NormalizeAudio(target) == source {
		return copyAudio.Name
	}
	return target
}

func containerFor(v Family, audio string) string {
	if IsIOSCompatible(v, audio, "mp4") || v == FamilyAV1 {
		return "mp4"
	}
	if v == FamilyVP9 && audio == "opus" {
		return "webm"
	}
	return "mkv"
}

func capFps(p Profile, src Source, limit float64, reasons *[]string) Profile {
	if src.Fps > limit+1 {
		p.MaxFps = limit
		*reasons = append(*reasons, fmt.Sprintf("Frame rate reduced from %.0f to %.0f fps.", src.Fps, limit))
	}
	return p
}

func bitsPerPixel(src Source) float64 {
	if src.Bitrate <= 0 || src.Width <= 0 || src.Height <= 0 {
		return 0
	}
	fps := src.Fps
	if fps <= 0 {
		fps = 30
	}
	return float64(src.Bitrate) / (float64(src.Width*src.Height) * fps)
}

func (f Family) String() string { return string(f) }
