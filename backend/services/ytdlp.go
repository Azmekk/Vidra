package services

import (
	"context"
	"os/exec"
)

type YtdlpService struct {
	settings *SettingsService
}

type YtdlpDownloadOptions struct {
	FormatID          string
	OutputPattern     string
	WriteThumbnail    bool
	ConvertThumbnails string
}

func NewYtdlpService(settings *SettingsService) *YtdlpService {
	return &YtdlpService{
		settings: settings,
	}
}

// baseArgs returns common arguments including proxy if configured
func (s *YtdlpService) baseArgs(ctx context.Context) []string {
	args := []string{}
	if proxyURL := s.settings.GetProxyURL(ctx); proxyURL != "" {
		args = append(args, "--proxy", proxyURL)
	}
	return args
}

// MetadataCommand builds a yt-dlp command for fetching video metadata
func (s *YtdlpService) MetadataCommand(ctx context.Context, url string) *exec.Cmd {
	args := []string{"--dump-json", "--flat-playlist", "--no-warnings"}
	args = append(args, s.baseArgs(ctx)...)
	args = append(args, url)
	return exec.CommandContext(ctx, "yt-dlp", args...)
}

// DownloadCommand builds a yt-dlp command for downloading a video
func (s *YtdlpService) DownloadCommand(ctx context.Context, url string, opts YtdlpDownloadOptions) *exec.Cmd {
	args := []string{"-f", opts.FormatID, "-o", opts.OutputPattern, "--newline"}

	if opts.WriteThumbnail {
		args = append(args, "--write-thumbnail")
	}
	if opts.ConvertThumbnails != "" {
		args = append(args, "--convert-thumbnails", opts.ConvertThumbnails)
	}

	args = append(args, s.baseArgs(ctx)...)
	args = append(args, url)
	return exec.Command("yt-dlp", args...)
}

// UpdateCommand builds a yt-dlp command for updating yt-dlp itself
func (s *YtdlpService) UpdateCommand(ctx context.Context) *exec.Cmd {
	args := []string{"-U"}
	args = append(args, s.baseArgs(ctx)...)
	return exec.CommandContext(ctx, "yt-dlp", args...)
}
