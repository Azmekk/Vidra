package services

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type YtdlpService struct {
	settings *SettingsService
	managed  string

	installMu sync.Mutex
	mu        sync.Mutex
	releases  []YtdlpRelease
	fetched   time.Time
}

type YtdlpDownloadOptions struct {
	Format        string
	OutputPattern string
}

// NewYtdlpService manages its own yt-dlp binary in dataDir/bin, so updates
// survive container recreation. The yt-dlp on PATH is used until it exists.
func NewYtdlpService(settings *SettingsService, dataDir string) *YtdlpService {
	name := "yt-dlp"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return &YtdlpService{settings: settings, managed: filepath.Join(dataDir, "bin", name)}
}

func (s *YtdlpService) binary() string {
	if _, err := os.Stat(s.managed); err == nil {
		return s.managed
	}
	return "yt-dlp"
}

func (s *YtdlpService) baseArgs(ctx context.Context) []string {
	args := []string{"--no-warnings", "--ignore-config"}
	if proxyURL := s.settings.MustGet(ctx).ProxyURL; proxyURL != "" {
		args = append(args, "--proxy", proxyURL)
	}
	return args
}

func (s *YtdlpService) MetadataCommand(ctx context.Context, url string) *exec.Cmd {
	args := append([]string{"--dump-json", "--no-playlist"}, s.baseArgs(ctx)...)
	return exec.CommandContext(ctx, s.binary(), append(args, "--", url)...)
}

// DownloadCommand downloads a single video plus its thumbnail and info JSON.
func (s *YtdlpService) DownloadCommand(ctx context.Context, url string, opts YtdlpDownloadOptions) *exec.Cmd {
	args := []string{
		"-o", opts.OutputPattern,
		"--newline",
		"--no-playlist",
		"--no-part",
		"--no-mtime",
		"--concurrent-fragments", "4",
		"--write-thumbnail", "--convert-thumbnails", "jpg",
		"--write-info-json", "--no-clean-info-json",
	}
	if opts.Format != "" {
		args = append(args, "-f", opts.Format)
	}
	if s.settings.MustGet(ctx).PreferCompatibleFormats {
		args = append(args, "-S", "vcodec:h264,res,acodec:m4a", "--merge-output-format", "mp4")
	}
	args = append(args, s.baseArgs(ctx)...)
	return exec.CommandContext(ctx, s.binary(), append(args, "--", url)...)
}
