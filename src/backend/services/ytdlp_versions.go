package services

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	ytdlpDownloadBase = "https://github.com/yt-dlp/yt-dlp/releases/download"
	ytdlpReleasesAPI  = "https://api.github.com/repos/yt-dlp/yt-dlp/releases?per_page=30"
	ytdlpCheckEvery   = 24 * time.Hour
	ytdlpReleasesTTL  = time.Hour
)

var (
	ErrInvalidYtdlpVersion = errors.New("version must look like 2025.09.26")
	ytdlpVersionRe         = regexp.MustCompile(`^\d{4}\.\d{2}\.\d{2}(\.\d+)?$`)
)

type YtdlpRelease struct {
	Version     string `json:"version"`
	PublishedAt string `json:"publishedAt"`
}

type YtdlpStatus struct {
	Version       string         `json:"version,omitempty"`
	Latest        string         `json:"latest,omitempty"`
	Pinned        string         `json:"pinned,omitempty"`
	Releases      []YtdlpRelease `json:"releases"`
	ReleasesError string         `json:"releasesError,omitempty"`
}

// Start installs the pinned or latest yt-dlp now and checks again daily.
func (s *YtdlpService) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(ytdlpCheckEvery)
		defer ticker.Stop()
		for {
			if err := s.Sync(ctx, true); err != nil {
				slog.Warn("yt-dlp update check failed", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// Sync installs the pinned version, or the latest release when nothing is pinned.
func (s *YtdlpService) Sync(ctx context.Context, fresh bool) error {
	target, err := s.settings.YtdlpPin(ctx)
	if err != nil {
		return err
	}
	if target == "" {
		releases, err := s.Releases(ctx, fresh)
		if err != nil {
			return err
		}
		if len(releases) == 0 {
			return errors.New("no yt-dlp releases found")
		}
		target = releases[0].Version
	}
	return s.install(ctx, target)
}

// Pin installs a specific version and keeps it until unpinned.
func (s *YtdlpService) Pin(ctx context.Context, version string) error {
	if !ytdlpVersionRe.MatchString(version) {
		return ErrInvalidYtdlpVersion
	}
	if err := s.install(ctx, version); err != nil {
		return err
	}
	if err := s.settings.SetYtdlpPin(ctx, version); err != nil {
		return err
	}
	slog.Info("yt-dlp pinned", "version", version)
	return nil
}

// Unpin follows the latest release again and installs it.
func (s *YtdlpService) Unpin(ctx context.Context) error {
	if err := s.settings.SetYtdlpPin(ctx, ""); err != nil {
		return err
	}
	slog.Info("yt-dlp unpinned, following the latest release")
	return s.Sync(ctx, true)
}

func (s *YtdlpService) Status(ctx context.Context) YtdlpStatus {
	status := YtdlpStatus{Version: s.Version(ctx), Releases: []YtdlpRelease{}}
	status.Pinned, _ = s.settings.YtdlpPin(ctx)
	releases, err := s.Releases(ctx, false)
	if err != nil {
		status.ReleasesError = err.Error()
	}
	if len(releases) > 0 {
		status.Releases, status.Latest = releases, releases[0].Version
	}
	return status
}

// Version reports the installed version, or "" when yt-dlp cannot run.
func (s *YtdlpService) Version(ctx context.Context) string {
	return binaryVersion(ctx, s.binary())
}

// Releases lists recent stable releases, newest first. A failed refresh
// falls back to the last successful list.
func (s *YtdlpService) Releases(ctx context.Context, fresh bool) ([]YtdlpRelease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !fresh && s.releases != nil && time.Since(s.fetched) < ytdlpReleasesTTL {
		return s.releases, nil
	}
	releases, err := s.fetchReleases(ctx)
	if err != nil {
		return s.releases, fmt.Errorf("cannot list yt-dlp releases: %w", err)
	}
	s.releases, s.fetched = releases, time.Now()
	return releases, nil
}

func (s *YtdlpService) fetchReleases(ctx context.Context) ([]YtdlpRelease, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ytdlpReleasesAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Vidra")
	res, err := s.client(ctx).Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned %s", res.Status)
	}
	var raw []struct {
		Tag        string    `json:"tag_name"`
		Published  time.Time `json:"published_at"`
		Prerelease bool      `json:"prerelease"`
		Draft      bool      `json:"draft"`
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, err
	}
	releases := make([]YtdlpRelease, 0, len(raw))
	for _, r := range raw {
		if !r.Prerelease && !r.Draft && ytdlpVersionRe.MatchString(r.Tag) {
			releases = append(releases, YtdlpRelease{Version: r.Tag, PublishedAt: r.Published.UTC().Format("2006-01-02T15:04:05.000Z")})
		}
	}
	return releases, nil
}

// install downloads a release next to the managed binary, checks that it
// runs and atomically replaces the old one.
func (s *YtdlpService) install(ctx context.Context, version string) error {
	s.installMu.Lock()
	defer s.installMu.Unlock()
	if _, err := os.Stat(s.managed); err == nil && binaryVersion(ctx, s.managed) == version {
		return nil
	}
	previous := s.Version(ctx)
	if err := os.MkdirAll(filepath.Dir(s.managed), 0o755); err != nil {
		return err
	}
	ext := filepath.Ext(s.managed)
	tmp := strings.TrimSuffix(s.managed, ext) + ".new" + ext
	defer os.Remove(tmp)

	if err := s.download(ctx, ytdlpDownloadBase+"/"+version+"/"+ytdlpAsset(), tmp); err != nil {
		return fmt.Errorf("download yt-dlp %s: %w", version, err)
	}
	if got := binaryVersion(ctx, tmp); got != version {
		return fmt.Errorf("downloaded yt-dlp reports version %q, expected %s", got, version)
	}
	if err := os.Rename(tmp, s.managed); err != nil {
		return err
	}
	slog.Info("yt-dlp installed", "version", version, "previous", cmp.Or(previous, "none"))
	return nil
}

func (s *YtdlpService) download(ctx context.Context, src, dst string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return err
	}
	res, err := s.client(ctx).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub returned %s", res.Status)
	}
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, res.Body); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// client honours the proxy from Settings, like yt-dlp itself.
func (s *YtdlpService) client(ctx context.Context) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if proxy := s.settings.MustGet(ctx).ProxyURL; proxy != "" {
		if u, err := url.Parse(proxy); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{Transport: transport}
}

func binaryVersion(ctx context.Context, bin string) string {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, bin, "--version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ytdlpAsset() string {
	switch runtime.GOOS {
	case "windows":
		return "yt-dlp.exe"
	case "darwin":
		return "yt-dlp_macos"
	default:
		return "yt-dlp"
	}
}
