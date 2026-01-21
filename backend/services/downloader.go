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
	"github.com/Azmekk/Vidra/backend/utils"
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
	StatusFinished    DownloadStatus = "completed"
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

func (p *DownloadProgress) Update(ws *WebSocketService, id string, percent, encodingPercent float64, speed, eta string, status DownloadStatus, lastOutput string) {
	p.mu.Lock()
	p.Percent = percent
	p.EncodingPercent = encodingPercent
	p.Speed = speed
	p.ETA = eta
	p.Status = status
	p.LastOutput = lastOutput
	p.mu.Unlock()

	if ws != nil {
		ws.Broadcast(WsEventProgress, map[string]interface{}{
			"id":              id,
			"percent":         percent,
			"encodingPercent": encodingPercent,
			"speed":           speed,
			"eta":             eta,
			"status":          status,
			"last_output":     lastOutput,
		})
	}
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
	ws       *WebSocketService
}

func NewDownloaderService(queries *database.Queries, ws *WebSocketService) *DownloaderService {
	return &DownloaderService{
		queries: queries,
		ws:      ws,
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

func (s *DownloaderService) DeleteVideoFiles(fileName, thumbnailFileName string) {
	if fileName != "" {
		path := filepath.Join("downloads", fileName)
		log.Printf("INFO: Deleting video file: %s\n", path)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Printf("WARN: Failed to delete video file %s: %v\n", path, err)
		}
	}
	if thumbnailFileName != "" {
		path := filepath.Join("downloads", thumbnailFileName)
		log.Printf("INFO: Deleting thumbnail file: %s\n", path)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Printf("WARN: Failed to delete thumbnail file %s: %v\n", path, err)
		}
	}
}

func (s *DownloaderService) UpdateYtdlp(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "-U")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (s *DownloaderService) GetVideoMetadata(ctx context.Context, url string) (*VideoMetadata, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "--dump-json", "--flat-playlist", "--no-warnings", url)
	log.Printf("DEBUG: Getting metadata with command: %s\n", cmd.String())

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		log.Printf("ERROR: yt-dlp metadata failed: %v, stderr: %s\n", err, stderr.String())
		return nil, fmt.Errorf("failed to get metadata: %w (stderr: %s)", err, stderr.String())
	}

	// yt-dlp might output multiple lines if there's any noise,
	// we try to find the line that starts with {
	var raw map[string]interface{}
	lines := strings.Split(string(output), "\n")
	found := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") {
			if err := json.Unmarshal([]byte(line), &raw); err == nil {
				found = true
				break
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("failed to find valid JSON in yt-dlp output")
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

func (s *DownloaderService) StartDownload(ctx context.Context, id pgtype.UUID, url string, formatID string, finalBaseName string, reEncode bool) {
	idStr := id.String()
	finalBaseName = utils.SanitizeFilename(finalBaseName)

	log.Printf("INFO [%s]: Initializing download task for URL: %s (final name: %s)\n", idStr, url, finalBaseName)

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
		prog.Update(s.ws, idStr, 0, 0, "", "", StatusDownloading, "Starting download...")

		cmd := exec.Command("yt-dlp", "-f", f, "-o", tempPathPattern, "--write-thumbnail", "--convert-thumbnails", "jpg", "--newline", url)
		log.Printf("DEBUG [%s]: Executing command: %s\n", idStr, cmd.String())
		var fullOutput bytes.Buffer

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("ERROR [%s]: Failed to create stdout pipe: %v\n", idStr, err)
			prog.Update(s.ws, idStr, 0, 0, "", "", StatusError, "Failed to create stdout pipe: "+err.Error())
			return
		}
		cmd.Stderr = &fullOutput

		if err := cmd.Start(); err != nil {
			log.Printf("ERROR [%s]: Failed to start yt-dlp: %v\n", idStr, err)
			prog.Update(s.ws, idStr, 0, 0, "", "", StatusError, "Failed to start yt-dlp: "+err.Error())
			s.queries.CreateError(context.Background(), database.CreateErrorParams{
				VideoID:      id,
				ErrorMessage: err.Error(),
				Command:      "yt-dlp (start)",
				Output:       "",
			})
			s.queries.UpdateVideoStatus(context.Background(), database.UpdateVideoStatusParams{
				ID:             id,
				DownloadStatus: string(StatusError),
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
				prog.Update(s.ws, idStr, percent, 0, matches[2], matches[3], StatusDownloading, line)
			} else {
				prog.mu.Lock()
				prog.LastOutput = line
				prog.mu.Unlock()
			}
		}

		if err := cmd.Wait(); err != nil {
			outputStr := fullOutput.String()
			log.Printf("ERROR [%s]: yt-dlp download failed: %v\nOutput: %s\n", idStr, err, outputStr)
			prog.Update(s.ws, idStr, prog.Percent, 0, prog.Speed, prog.ETA, StatusError, fmt.Sprintf("Download failed: %v", err))

			s.queries.CreateError(context.Background(), database.CreateErrorParams{
				VideoID:      id,
				ErrorMessage: err.Error(),
				Command:      "yt-dlp",
				Output:       outputStr,
			})
			s.queries.UpdateVideoStatus(context.Background(), database.UpdateVideoStatusParams{
				ID:             id,
				DownloadStatus: string(StatusError),
			})
			return
		}

		log.Printf("INFO [%s]: Download completed. Searching for downloaded file...\n", idStr)

		// 2. Find the downloaded file
		files, _ := filepath.Glob(filepath.Join("downloads", idStr+".*"))
		if len(files) == 0 {
			msg := "Downloaded file not found in downloads directory"
			log.Printf("ERROR [%s]: %s\n", idStr, msg)
			prog.Update(s.ws, idStr, 100, 0, "", "", StatusError, msg)
			s.queries.CreateError(context.Background(), database.CreateErrorParams{
				VideoID:      id,
				ErrorMessage: msg,
				Command:      "file-glob",
				Output:       "",
			})
			s.queries.UpdateVideoStatus(context.Background(), database.UpdateVideoStatusParams{
				ID:             id,
				DownloadStatus: string(StatusError),
			})
			return
		}
		var tempFile string
		for _, f := range files {
			ext := strings.ToLower(filepath.Ext(f))
			// Skip thumbnails and temporary files
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".part" || ext == ".ytdl" {
				continue
			}
			tempFile = f
			break
		}

		if tempFile == "" {
			msg := "Downloaded video file not found in downloads directory (only found thumbnails)"
			log.Printf("ERROR [%s]: %s\n", idStr, msg)
			prog.Update(s.ws, idStr, 100, 0, "", "", StatusError, msg)
			s.queries.CreateError(context.Background(), database.CreateErrorParams{
				VideoID:      id,
				ErrorMessage: msg,
				Command:      "file-glob-check",
				Output:       "",
			})
			s.queries.UpdateVideoStatus(context.Background(), database.UpdateVideoStatusParams{
				ID:             id,
				DownloadStatus: string(StatusError),
			})
			return
		}
		log.Printf("INFO [%s]: Found temporary video file: %s\n", idStr, tempFile)

		// 3. Process video (Encode or Rename)
		var finalFileName string
		if reEncode {
			finalFileName = finalBaseName + ".mp4"
			finalPath := filepath.Join("downloads", finalFileName)
			tempEncodePath := filepath.Join("downloads", idStr+"_encoded.mp4")

			log.Printf("INFO [%s]: Starting ffmpeg encoding to H.264: %s\n", idStr, tempEncodePath)
			prog.Update(s.ws, idStr, 100, 0, "", "", StatusEncoding, "Getting video duration...")

			// Get duration for progress calculation
			durationCmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", tempFile)
			durationOut, err := durationCmd.Output()
			duration := 0.0
			if err == nil {
				duration, _ = strconv.ParseFloat(strings.TrimSpace(string(durationOut)), 64)
			}
			log.Printf("INFO [%s]: Video duration: %.2fs\n", idStr, duration)

			prog.Update(s.ws, idStr, 100, 0, "", "", StatusEncoding, "Encoding to H264...")

			encodeCmd := exec.Command("ffmpeg", "-i", tempFile, "-c:v", "libx264", "-crf", "23", "-c:a", "aac", "-progress", "-", "-y", tempEncodePath)

			var encodeOutput bytes.Buffer
			encodeStdout, err := encodeCmd.StdoutPipe()
			if err != nil {
				log.Printf("ERROR [%s]: Failed to create ffmpeg stdout pipe: %v\n", idStr, err)
				prog.Update(s.ws, idStr, 100, 0, "", "", StatusError, "Failed to create ffmpeg stdout pipe: "+err.Error())
				return
			}
			encodeCmd.Stderr = &encodeOutput

			if err := encodeCmd.Start(); err != nil {
				log.Printf("ERROR [%s]: Failed to start ffmpeg: %v\n", idStr, err)
				prog.Update(s.ws, idStr, 100, 0, "", "", StatusError, "Failed to start ffmpeg: "+err.Error())
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
						prog.Update(s.ws, idStr, 100, encodingPercent, "", "", StatusEncoding, "Encoding in progress...")
					}
				}
			}

			if err := encodeCmd.Wait(); err != nil {
				outputStr := encodeOutput.String()
				log.Printf("ERROR [%s]: ffmpeg encoding failed: %v\nOutput: %s\n", idStr, err, outputStr)
				prog.Update(s.ws, idStr, 100, 0, "", "", StatusError, fmt.Sprintf("Encoding failed: %v\nOutput: %s", err, outputStr))
				os.Remove(tempEncodePath) // Clean up partial encoded file

				s.queries.CreateError(context.Background(), database.CreateErrorParams{
					VideoID:      id,
					ErrorMessage: err.Error(),
					Command:      "ffmpeg",
					Output:       outputStr,
				})
				s.queries.UpdateVideoStatus(context.Background(), database.UpdateVideoStatusParams{
					ID:             id,
					DownloadStatus: string(StatusError),
				})
				return
			}

			// Move encoded file to final path
			if err := os.Rename(tempEncodePath, finalPath); err != nil {
				log.Printf("ERROR [%s]: Failed to rename encoded file: %v\n", idStr, err)
				prog.Update(s.ws, idStr, 100, 0, "", "", StatusError, "Failed to rename encoded file: "+err.Error())
				return
			}
			log.Printf("INFO [%s]: Encoding successful. Cleaning up temporary file: %s\n", idStr, tempFile)
		} else {
			log.Printf("INFO [%s]: Skipping re-encoding as requested.\n", idStr)
			prog.Update(s.ws, idStr, 100, 100, "", "", StatusEncoding, "Skipping encoding...")

			finalFileName = finalBaseName + filepath.Ext(tempFile)
			finalPath := filepath.Join("downloads", finalFileName)

			if err := os.Rename(tempFile, finalPath); err != nil {
				log.Printf("ERROR [%s]: Failed to rename downloaded file: %v\n", idStr, err)
				prog.Update(s.ws, idStr, 100, 0, "", "", StatusError, "Failed to rename downloaded file: "+err.Error())
				return
			}
			log.Printf("INFO [%s]: Rename successful: %s -> %s\n", idStr, tempFile, finalPath)
		}

		// 4. Handle thumbnail
		finalThumbnailName := finalBaseName + ".jpg"
		finalThumbnailPath := filepath.Join("downloads", finalThumbnailName)

		// yt-dlp saves thumbnail as idStr.jpg due to --convert-thumbnails jpg and our -o pattern
		tempThumbnailPath := filepath.Join("downloads", idStr+".jpg")
		if _, err := os.Stat(tempThumbnailPath); err == nil {
			log.Printf("INFO [%s]: Found thumbnail: %s, renaming to: %s\n", idStr, tempThumbnailPath, finalThumbnailPath)
			if err := os.Rename(tempThumbnailPath, finalThumbnailPath); err != nil {
				log.Printf("WARN [%s]: Failed to rename thumbnail: %v\n", idStr, err)
				finalThumbnailName = "" // Reset if rename failed
			}
		} else {
			log.Printf("WARN [%s]: Thumbnail not found at %s\n", idStr, tempThumbnailPath)
			finalThumbnailName = ""
		}

		// 5. Cleanup all remaining temporary files for this ID
		log.Printf("INFO [%s]: Cleaning up temporary files matching %s.*\n", idStr, idStr)
		remainingFiles, _ := filepath.Glob(filepath.Join("downloads", idStr+".*"))
		for _, f := range remainingFiles {
			if err := os.Remove(f); err != nil {
				if !os.IsNotExist(err) {
					log.Printf("WARN [%s]: Failed to remove temporary file %s: %v\n", idStr, f, err)
				}
			} else {
				log.Printf("INFO [%s]: Removed temporary file: %s\n", idStr, f)
			}
		}

		// 6. Update database
		log.Printf("INFO [%s]: Updating database with final file names and status.\n", idStr)
		prog.Update(s.ws, idStr, 100, 100, "", "", StatusFinished, "Processing complete")

		_, err = s.queries.UpdateVideoFiles(context.Background(), database.UpdateVideoFilesParams{
			ID:                id,
			FileName:          pgtype.Text{String: finalFileName, Valid: true},
			ThumbnailFileName: pgtype.Text{String: finalThumbnailName, Valid: finalThumbnailName != ""},
		})
		if err != nil {
			log.Printf("ERROR [%s]: Failed to update video file names in database: %v\n", idStr, err)
		}

		_, err = s.queries.UpdateVideoStatus(context.Background(), database.UpdateVideoStatusParams{
			ID:             id,
			DownloadStatus: string(StatusFinished),
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
