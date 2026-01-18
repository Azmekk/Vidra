package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

type DownloadStatus string

const (
	StatusPending     DownloadStatus = "pending"
	StatusDownloading DownloadStatus = "downloading"
	StatusEncoding    DownloadStatus = "encoding"
	StatusFinished    DownloadStatus = "finished"
	StatusError       DownloadStatus = "error"
)

type DownloadProgressDTO struct {
	Percent         float64        `json:"percent"`
	EncodingPercent float64        `json:"encodingPercent"`
	Speed           string         `json:"speed"`
	ETA             string         `json:"eta"`
	Status          DownloadStatus `json:"status"`
	LastOutput      string         `json:"last_output"`
}

type DownloadProgress struct {
	mu              sync.RWMutex
	Percent         float64
	EncodingPercent float64
	Speed           string
	ETA             string
	Status          DownloadStatus
	LastOutput      string
}

func (p *DownloadProgress) Update(percent, encodingPercent float64, speed, eta string, status DownloadStatus, lastOutput string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Percent = percent
	p.EncodingPercent = encodingPercent
	p.Speed = speed
	p.ETA = eta
	p.Status = status
	p.LastOutput = lastOutput
}

func (p *DownloadProgress) GetSnapshot() DownloadProgressDTO {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return DownloadProgressDTO{
		Percent:         p.Percent,
		EncodingPercent: p.EncodingPercent,
		Speed:           p.Speed,
		ETA:             p.ETA,
		Status:          p.Status,
		LastOutput:      p.LastOutput,
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

func (s *DownloaderService) GetProgress(id string) (DownloadProgressDTO, bool) {
	val, ok := s.progress.Load(id)
	if !ok {
		return DownloadProgressDTO{}, false
	}
	return val.(*DownloadProgress).GetSnapshot(), true
}

func (s *DownloaderService) GetAllProgress() map[string]DownloadProgressDTO {
	allProgress := make(map[string]DownloadProgressDTO)
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

func (s *DownloaderService) StartDownload(ctx context.Context, id pgtype.UUID, url string, formatID string, finalBaseName string) {
	idStr := id.String()

	log.Printf("INFO [%s]: Initializing download task for URL: %s\n", idStr, url)

	prog := &DownloadProgress{Status: StatusPending}
	s.progress.Store(idStr, prog)

	go func() {
		// 1. Download as guid.ext
		f := formatID
		if f == "" {
			f = "bestvideo+bestaudio/best"
		} else {
			f = f + "+bestaudio/best"
		}

		tempPathPattern := filepath.Join("downloads", idStr+".%(ext)s")
		log.Printf("INFO [%s]: Starting yt-dlp download with format: %s\n", idStr, f)
		prog.Update(0, 0, "", "", StatusDownloading, "Starting download...")

		cmd := exec.Command("yt-dlp", "-f", f, "-o", tempPathPattern, "--newline", url)
		var fullOutput bytes.Buffer

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("ERROR [%s]: Failed to create stdout pipe: %v\n", idStr, err)
			prog.Update(0, 0, "", "", StatusError, "Failed to create stdout pipe: "+err.Error())
			return
		}
		cmd.Stderr = &fullOutput

		if err := cmd.Start(); err != nil {
			log.Printf("ERROR [%s]: Failed to start yt-dlp: %v\n", idStr, err)
			prog.Update(0, 0, "", "", StatusError, "Failed to start yt-dlp: "+err.Error())
			s.queries.CreateError(context.Background(), database.CreateErrorParams{
				VideoID:      id,
				ErrorMessage: err.Error(),
				Command:      "yt-dlp (start)",
				Output:       "",
			})
			return
		}

		// Use TeeReader to capture stdout while scanning
		multiReader := io.TeeReader(stdout, &fullOutput)
		scanner := bufio.NewScanner(multiReader)
		progressRegex := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)%\s+of\s+.*\s+at\s+(.*)\s+ETA\s+(.*)`)

		for scanner.Scan() {
			line := scanner.Text()
			matches := progressRegex.FindStringSubmatch(line)
			if len(matches) == 4 {
				percent, _ := strconv.ParseFloat(matches[1], 64)
				prog.Update(percent, 0, matches[2], matches[3], StatusDownloading, line)
			} else {
				prog.mu.Lock()
				prog.LastOutput = line
				prog.mu.Unlock()
			}
		}

		if err := cmd.Wait(); err != nil {
			outputStr := fullOutput.String()
			log.Printf("ERROR [%s]: yt-dlp download failed: %v\nOutput: %s\n", idStr, err, outputStr)
			prog.Update(prog.Percent, 0, prog.Speed, prog.ETA, StatusError, fmt.Sprintf("Download failed: %v", err))

			s.queries.CreateError(context.Background(), database.CreateErrorParams{
				VideoID:      id,
				ErrorMessage: err.Error(),
				Command:      "yt-dlp",
				Output:       outputStr,
			})
			return
		}

		log.Printf("INFO [%s]: Download completed. Searching for downloaded file...\n", idStr)

		// 2. Find the downloaded file
		files, _ := filepath.Glob(filepath.Join("downloads", idStr+".*"))
		if len(files) == 0 {
			msg := "Downloaded file not found in downloads directory"
			log.Printf("ERROR [%s]: %s\n", idStr, msg)
			prog.Update(100, 0, "", "", StatusError, msg)
			s.queries.CreateError(context.Background(), database.CreateErrorParams{
				VideoID:      id,
				ErrorMessage: msg,
				Command:      "file-glob",
				Output:       "",
			})
			return
		}
		tempFile := files[0]
		log.Printf("INFO [%s]: Found temporary file: %s\n", idStr, tempFile)

		// 3. Encode to H264 and rename
		finalFileName := finalBaseName + ".mp4"
		finalPath := filepath.Join("downloads", finalFileName)

		log.Printf("INFO [%s]: Starting ffmpeg encoding to H.264: %s\n", idStr, finalPath)
		prog.Update(100, 0, "", "", StatusEncoding, "Getting video duration...")

		// Get duration for progress calculation
		durationCmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", tempFile)
		durationOut, err := durationCmd.Output()
		duration := 0.0
		if err == nil {
			duration, _ = strconv.ParseFloat(strings.TrimSpace(string(durationOut)), 64)
		}
		log.Printf("INFO [%s]: Video duration: %.2fs\n", idStr, duration)

		prog.Update(100, 0, "", "", StatusEncoding, "Encoding to H264...")

		encodeCmd := exec.Command("ffmpeg", "-i", tempFile, "-c:v", "libx264", "-preset", "fast", "-c:a", "aac", "-progress", "-", "-y", finalPath)

		var encodeOutput bytes.Buffer
		encodeStdout, err := encodeCmd.StdoutPipe()
		if err != nil {
			log.Printf("ERROR [%s]: Failed to create ffmpeg stdout pipe: %v\n", idStr, err)
			prog.Update(100, 0, "", "", StatusError, "Failed to create ffmpeg stdout pipe: "+err.Error())
			return
		}
		encodeCmd.Stderr = &encodeOutput

		if err := encodeCmd.Start(); err != nil {
			log.Printf("ERROR [%s]: Failed to start ffmpeg: %v\n", idStr, err)
			prog.Update(100, 0, "", "", StatusError, "Failed to start ffmpeg: "+err.Error())
			return
		}

		encodeScanner := bufio.NewScanner(encodeStdout)
		for encodeScanner.Scan() {
			line := encodeScanner.Text()
			if after, ok := strings.CutPrefix(line, "out_time_ms="); ok {
				timeUsStr := after
				timeUs, _ := strconv.ParseFloat(timeUsStr, 64)
				if duration > 0 {
					encodingPercent := (timeUs / 1000000.0 / duration) * 100.0
					if encodingPercent > 100 {
						encodingPercent = 100
					}
					prog.Update(100, encodingPercent, "", "", StatusEncoding, "Encoding in progress...")
				}
			}
		}

		if err := encodeCmd.Wait(); err != nil {
			outputStr := encodeOutput.String()
			log.Printf("ERROR [%s]: ffmpeg encoding failed: %v\nOutput: %s\n", idStr, err, outputStr)
			prog.Update(100, 0, "", "", StatusError, fmt.Sprintf("Encoding failed: %v\nOutput: %s", err, outputStr))

			s.queries.CreateError(context.Background(), database.CreateErrorParams{
				VideoID:      id,
				ErrorMessage: err.Error(),
				Command:      "ffmpeg",
				Output:       outputStr,
			})
			return
		}

		log.Printf("INFO [%s]: Encoding successful. Cleaning up temporary file: %s\n", idStr, tempFile)

		// 4. Cleanup temp file
		if err := os.Remove(tempFile); err != nil {
			log.Printf("WARN [%s]: Failed to remove temporary file: %v\n", idStr, err)
		}

		// 5. Update database
		log.Printf("INFO [%s]: Updating database with final file name and status.\n", idStr)
		prog.Update(100, 100, "", "", StatusFinished, "Processing complete")

		_, err = s.queries.UpdateVideoFiles(context.Background(), database.UpdateVideoFilesParams{
			ID:       id,
			FileName: pgtype.Text{String: finalFileName, Valid: true},
		})
		if err != nil {
			log.Printf("ERROR [%s]: Failed to update video file name in database: %v\n", idStr, err)
		}

		_, err = s.queries.UpdateVideoStatus(context.Background(), database.UpdateVideoStatusParams{
			ID:             id,
			DownloadStatus: "completed",
		})
		if err != nil {
			log.Printf("ERROR [%s]: Failed to update video status in database: %v\n", idStr, err)
		}

		log.Printf("SUCCESS [%s]: Video download and processing finished successfully.\n", idStr)
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
