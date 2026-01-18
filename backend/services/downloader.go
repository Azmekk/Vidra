package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"sync"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type VideoOption struct {
	FormatID   string  `json:"format_id"`
	Extension  string  `json:"extension"`
	Resolution string  `json:"resolution"`
	Note       string  `json:"note"`
	FileSize   float64 `json:"file_size"`
	VCodec     string  `json:"vcodec"`
	ACodec     string  `json:"acodec"`
}

type VideoMetadata struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Duration    float64       `json:"duration"`
	Thumbnail   string        `json:"thumbnail"`
	Options     []VideoOption `json:"options"`
}

type DownloadProgress struct {
	mu         sync.RWMutex
	Percent    float64 `json:"percent"`
	Speed      string  `json:"speed"`
	ETA        string  `json:"eta"`
	Status     string  `json:"status"` // pending, downloading, finished, error
	LastOutput string  `json:"last_output"`
}

func (p *DownloadProgress) Update(percent float64, speed, eta, status, lastOutput string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Percent = percent
	p.Speed = speed
	p.ETA = eta
	p.Status = status
	p.LastOutput = lastOutput
}

func (p *DownloadProgress) GetSnapshot() DownloadProgress {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return DownloadProgress{
		Percent:    p.Percent,
		Speed:      p.Speed,
		ETA:        p.ETA,
		Status:     p.Status,
		LastOutput: p.LastOutput,
	}
}

type DownloaderService struct {
	progress sync.Map // map[string]*DownloadProgress
	queries  *database.Queries
}

func NewDownloaderService(queries *database.Queries) *DownloaderService {
	return &DownloaderService{
		queries: queries,
	}
}

func (s *DownloaderService) GetProgress(id string) (DownloadProgress, bool) {
	val, ok := s.progress.Load(id)
	if !ok {
		return DownloadProgress{}, false
	}
	return val.(*DownloadProgress).GetSnapshot(), true
}

func (s *DownloaderService) GetAllProgress() map[string]DownloadProgress {
	allProgress := make(map[string]DownloadProgress)
	s.progress.Range(func(key, value interface{}) bool {
		id := key.(string)
		prog := value.(*DownloadProgress)
		allProgress[id] = prog.GetSnapshot()
		return true
	})
	return allProgress
}

func (s *DownloaderService) UpdateYtdlp(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "-U")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (s *DownloaderService) GetVideoMetadata(ctx context.Context, url string) (*VideoMetadata, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "--dump-json", "--flat-playlist", url)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	metadata := &VideoMetadata{
		Title:       getString(raw, "title"),
		Description: getString(raw, "description"),
		Duration:    getFloat(raw, "duration"),
		Thumbnail:   getString(raw, "thumbnail"),
	}

	if formats, ok := raw["formats"].([]interface{}); ok {
		for _, f := range formats {
			fmtObj := f.(map[string]interface{})
			metadata.Options = append(metadata.Options, VideoOption{
				FormatID:   getString(fmtObj, "format_id"),
				Extension:  getString(fmtObj, "ext"),
				Resolution: getString(fmtObj, "resolution"),
				Note:       getString(fmtObj, "format_note"),
				FileSize:   getFloat(fmtObj, "filesize"),
				VCodec:     getString(fmtObj, "vcodec"),
				ACodec:     getString(fmtObj, "acodec"),
			})
		}
	}

	return metadata, nil
}

func (s *DownloaderService) StartDownload(ctx context.Context, id pgtype.UUID, url string, formatID string, outputPath string) {
	var idStr string
	id.Scan(&idStr)

	prog := &DownloadProgress{Status: "pending"}
	s.progress.Store(idStr, prog)

	go func() {
		prog.Update(0, "", "", "downloading", "Starting process...")
		cmd := exec.Command("yt-dlp", "-f", formatID, "-o", outputPath, "--newline", url)
		
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			prog.Update(0, "", "", "error", err.Error())
			return
		}

		if err := cmd.Start(); err != nil {
			prog.Update(0, "", "", "error", err.Error())
			return
		}

		scanner := bufio.NewScanner(stdout)
		progressRegex := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)%\s+of\s+.*\s+at\s+(.*)\s+ETA\s+(.*)`)

		for scanner.Scan() {
			line := scanner.Text()
			matches := progressRegex.FindStringSubmatch(line)
			if len(matches) == 4 {
				percent, _ := strconv.ParseFloat(matches[1], 64)
				prog.Update(percent, matches[2], matches[3], "downloading", line)
			} else {
				// Update just the last output if it's not a progress line
				prog.mu.Lock()
				prog.LastOutput = line
				prog.mu.Unlock()
			}
		}

		if err := cmd.Wait(); err != nil {
			prog.Update(prog.Percent, prog.Speed, prog.ETA, "error", fmt.Sprintf("Process exited with error: %v", err))
		} else {
			prog.Update(100, "", "", "finished", "Download completed")
			
			// Update database status
			s.queries.UpdateVideoStatus(context.Background(), database.UpdateVideoStatusParams{
				ID:             id,
				DownloadStatus: "completed",
			})
		}
	}()
}
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}
