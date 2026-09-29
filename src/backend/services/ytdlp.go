package services

import (
	"context"
	"os/exec"
)

type YtdlpService struct {
	settings *SettingsService
}

type YtdlpDownloadOptions struct {
	Format        string
	OutputPattern string
}

func NewYtdlpService(settings *SettingsService) *YtdlpService {
	return &YtdlpService{settings: settings}
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
	return exec.CommandContext(ctx, "yt-dlp", append(args, "--", url)...)
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
	return exec.CommandContext(ctx, "yt-dlp", append(args, "--", url)...)
}

func (s *YtdlpService) UpdateCommand(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, "yt-dlp", append([]string{"-U"}, s.baseArgs(ctx)...)...)
}
