package encoding

import (
	"bufio"
	"bytes"
	"context"
	"log"
	"maps"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type EncoderInfo struct {
	VideoEncoder
	Curated   bool `json:"curated"`
	Hardware  bool `json:"hardware"`
	IOS       bool `json:"ios"`
	Available bool `json:"available"`
}

type AudioInfo struct {
	AudioEncoder
	Curated bool `json:"curated"`
}

type Capabilities struct {
	FFmpegVersion string        `json:"ffmpegVersion"`
	Video         []EncoderInfo `json:"video"`
	Audio         []AudioInfo   `json:"audio"`
	Containers    []string      `json:"containers"`

	video map[string]EncoderInfo
	audio map[string]AudioInfo
}

func (c *Capabilities) FindVideo(name string) (EncoderInfo, bool) {
	e, ok := c.video[name]
	return e, ok && e.Available
}

func (c *Capabilities) FindAudio(name string) (AudioInfo, bool) {
	a, ok := c.audio[name]
	return a, ok
}

// firstAvailable returns the first usable encoder in preference order.
func (c *Capabilities) firstAvailable(names ...string) (EncoderInfo, bool) {
	for _, n := range names {
		if e, ok := c.FindVideo(n); ok {
			return e, true
		}
	}
	return EncoderInfo{}, false
}

type ffmpegEncoder struct {
	kind        byte
	name        string
	description string
}

// Detect lists ffmpeg's encoders and test-encodes a frame with each
// hardware encoder so only ones that actually work are offered.
func Detect(ctx context.Context) *Capabilities {
	listed, version := listEncoders(ctx)
	byName := make(map[string]ffmpegEncoder, len(listed))
	for _, e := range listed {
		byName[e.name] = e
	}

	caps := &Capabilities{
		FFmpegVersion: version,
		Containers:    []string{"mp4", "mkv", "webm", "mov"},
		video:         map[string]EncoderInfo{},
		audio:         map[string]AudioInfo{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, enc := range videoCatalog {
		if _, ok := byName[enc.Name]; !ok {
			continue
		}
		info := EncoderInfo{VideoEncoder: enc, Curated: true, Hardware: enc.Hardware(), IOS: enc.IOS(), Available: true}
		if !enc.Hardware() {
			caps.video[enc.Name] = info
			continue
		}
		wg.Go(func() {
			info.Available = probeEncoder(ctx, enc)
			mu.Lock()
			caps.video[enc.Name] = info
			mu.Unlock()
		})
	}
	wg.Wait()

	caps.video[copyVideo.Name] = EncoderInfo{VideoEncoder: copyVideo, Curated: true, IOS: true, Available: true}
	caps.audio[copyAudio.Name] = AudioInfo{AudioEncoder: copyAudio, Curated: true}
	for _, a := range audioCatalog {
		if _, ok := byName[a.Name]; ok {
			caps.audio[a.Name] = AudioInfo{AudioEncoder: a, Curated: true}
		}
	}

	for _, e := range listed {
		switch e.kind {
		case 'V':
			if _, ok := caps.video[e.name]; !ok {
				caps.video[e.name] = EncoderInfo{VideoEncoder: VideoEncoder{
					Name: e.name, Label: e.name, Description: e.description, Family: FamilyAny,
					Containers: []string{"mkv"},
				}, Available: true}
			}
		case 'A':
			if _, ok := caps.audio[e.name]; !ok {
				caps.audio[e.name] = AudioInfo{AudioEncoder: AudioEncoder{
					Name: e.name, Label: e.name, Description: e.description, Containers: []string{"mkv"}, DefaultBitrate: 160,
				}}
			}
		}
	}

	videoOrder := []string{copyVideo.Name}
	for _, e := range videoCatalog {
		videoOrder = append(videoOrder, e.Name)
	}
	audioOrder := []string{copyAudio.Name}
	for _, a := range audioCatalog {
		audioOrder = append(audioOrder, a.Name)
	}
	caps.Video = ordered(caps.video, videoOrder)
	caps.Audio = ordered(caps.audio, audioOrder)

	available := 0
	for _, e := range caps.Video {
		if e.Hardware && e.Available {
			available++
		}
	}
	log.Printf("ffmpeg %s: %d video / %d audio encoders, %d working hardware encoders\n",
		version, len(caps.Video), len(caps.Audio), available)
	return caps
}

// ordered puts preferred names first, then the rest alphabetically.
func ordered[T any](m map[string]T, preferred []string) []T {
	out := make([]T, 0, len(m))
	seen := make(map[string]bool, len(m))
	for _, n := range preferred {
		if v, ok := m[n]; ok && !seen[n] {
			out = append(out, v)
			seen[n] = true
		}
	}
	for _, n := range slices.Sorted(maps.Keys(m)) {
		if !seen[n] {
			out = append(out, m[n])
		}
	}
	return out
}

func listEncoders(ctx context.Context) ([]ffmpegEncoder, string) {
	out, err := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-encoders").Output()
	if err != nil {
		log.Printf("WARN: failed to list ffmpeg encoders: %v\n", err)
		return nil, ""
	}

	var encoders []ffmpegEncoder
	started := false
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "------") {
			started = true
			continue
		}
		fields := strings.Fields(line)
		if !started || len(fields) < 2 || len(fields[0]) != 6 {
			continue
		}
		encoders = append(encoders, ffmpegEncoder{
			kind:        fields[0][0],
			name:        fields[1],
			description: strings.TrimSpace(strings.TrimPrefix(line, fields[0]+" "+fields[1])),
		})
	}

	version := ""
	if v, err := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-version").Output(); err == nil {
		if f := strings.Fields(string(v)); len(f) >= 3 {
			version = f[2]
		}
	}
	return encoders, version
}

func probeEncoder(ctx context.Context, enc VideoEncoder) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	args := []string{"-hide_banner", "-loglevel", "error"}
	args = append(args, enc.InitArgs...)
	args = append(args, "-f", "lavfi", "-i", "color=black:s=256x256:r=30:d=0.2")
	if enc.Filter != "" {
		args = append(args, "-vf", enc.Filter)
	}
	args = append(args, "-frames:v", "1", "-c:v", enc.Name, "-f", "null", "-")
	return exec.CommandContext(ctx, "ffmpeg", args...).Run() == nil
}

func itoa(i int) string { return strconv.Itoa(i) }
