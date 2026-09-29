package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/services/encoding"
)

var (
	ErrNothingToDo   = errors.New("the source already matches these settings")
	ErrFileNotReady  = errors.New("file version is not ready")
	ytdlpProgressRe  = regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)%\s+of\s+.*?\s+at\s+(.+?)\s+ETA\s+(\S+)`)
	skipDownloadExts = []string{".jpg", ".jpeg", ".png", ".webp", ".json", ".part", ".ytdl", ".temp"}
)

const maxOutput = 64 << 10

type DownloaderService struct {
	jobRegistry
	store     *VideoStore
	queries   *database.Queries
	settings  *SettingsService
	ytdlp     *YtdlpService
	caps      *encoding.Capabilities
	dir       string
	downloads *pool
	encodes   *pool
	onFile    []func(Video, database.VideoFile)
	onDelete  []func(names []string)
}

func NewDownloaderService(store *VideoStore, queries *database.Queries, ws *WebSocketService, settings *SettingsService,
	ytdlp *YtdlpService, caps *encoding.Capabilities, dir string) *DownloaderService {
	s := settings.MustGet(context.Background())
	d := &DownloaderService{
		jobRegistry: jobRegistry{ws: ws},
		store:       store,
		queries:     queries,
		settings:    settings,
		ytdlp:       ytdlp,
		caps:        caps,
		dir:         dir,
		downloads:   newPool(s.MaxConcurrentDownloads),
		encodes:     newPool(s.MaxConcurrentEncodes),
	}
	settings.OnChange(func(s Settings) {
		d.downloads.setLimit(s.MaxConcurrentDownloads)
		d.encodes.setLimit(s.MaxConcurrentEncodes)
	})
	return d
}

// OnFileCompleted registers a callback for every finished file version.
func (d *DownloaderService) OnFileCompleted(fn func(Video, database.VideoFile)) {
	d.onFile = append(d.onFile, fn)
}

// OnFilesDeleted registers a callback with the names of files removed from disk.
func (d *DownloaderService) OnFilesDeleted(fn func(names []string)) {
	d.onDelete = append(d.onDelete, fn)
}

func (d *DownloaderService) deleted(names []string) {
	if len(names) == 0 {
		return
	}
	for _, fn := range d.onDelete {
		fn(names)
	}
}

func (d *DownloaderService) Capabilities() *encoding.Capabilities { return d.caps }

func (d *DownloaderService) Path(name string) string {
	return filepath.Join(d.dir, filepath.Base(name))
}

// RecoverInterrupted marks jobs cut off by a restart and removes their partial files.
func (d *DownloaderService) RecoverInterrupted(ctx context.Context) {
	files, err := d.store.MarkInterrupted(ctx)
	if err != nil {
		log.Printf("WARN: failed to recover interrupted jobs: %v\n", err)
		return
	}
	for _, f := range files {
		d.removeTemp(f.ID)
		d.recordError(f.VideoID, f.ID, "restart", "Interrupted by server restart", "")
	}
}

// StartDownload creates the original file version and downloads it in the background.
func (d *DownloaderService) StartDownload(ctx context.Context, v Video, url, formatID string, req encoding.Request) (database.VideoFile, error) {
	file, err := d.store.CreateFile(ctx, database.CreateVideoFileParams{
		VideoID: v.ID, Kind: KindOriginal, Label: "Original", Status: FileQueued,
	})
	if err != nil {
		return database.VideoFile{}, err
	}
	j := d.start(v.ID, file.ID)
	go d.runDownload(j, v.ID, file, url, formatID, req)
	return file, nil
}

// StartEncode creates a new version of a video from an existing completed version.
func (d *DownloaderService) StartEncode(ctx context.Context, v Video, source database.VideoFile, req encoding.Request, makePrimary bool) (database.VideoFile, error) {
	if source.Status != FileCompleted || source.FileName == nil {
		return database.VideoFile{}, ErrFileNotReady
	}
	if err := req.Validate(d.caps); err != nil {
		return database.VideoFile{}, err
	}
	rec := req.Resolve(d.caps, SourceOf(source))
	if rec.Skip {
		return database.VideoFile{}, ErrNothingToDo
	}
	return d.enqueueEncode(ctx, v.ID, source, rec, makePrimary, false)
}

func (d *DownloaderService) enqueueEncode(ctx context.Context, videoID string, source database.VideoFile, rec encoding.Recommendation, makePrimary, removeSource bool) (database.VideoFile, error) {
	profile, err := json.Marshal(rec.Profile)
	if err != nil {
		return database.VideoFile{}, err
	}
	profileStr := string(profile)
	file, err := d.store.CreateFile(ctx, database.CreateVideoFileParams{
		VideoID: videoID, Kind: KindEncode, SourceFileID: &source.ID, Label: rec.Label,
		Status: FileQueued, EncodingProfile: &profileStr,
	})
	if err != nil {
		return database.VideoFile{}, err
	}
	j := d.start(videoID, file.ID)
	go d.runEncode(j, videoID, file, source, rec, makePrimary, removeSource)
	return file, nil
}

func (d *DownloaderService) runDownload(j *job, videoID string, file database.VideoFile, url, formatID string, req encoding.Request) {
	defer d.done(file.ID)
	ctx := j.ctx

	if err := d.downloads.acquire(ctx); err != nil {
		d.fail(j, videoID, file.ID, "yt-dlp", err, "")
		return
	}
	defer d.downloads.release()
	d.setStatus(j, videoID, file.ID, FileDownloading)

	format := ""
	if formatID != "" {
		format = formatID + "+bestaudio/" + formatID
	}
	cmd := d.ytdlp.DownloadCommand(ctx, url, YtdlpDownloadOptions{
		Format:        format,
		OutputPattern: filepath.Join(d.dir, file.ID+".%(ext)s"),
	})
	output, err := runStreaming(cmd, func(line string) {
		if m := ytdlpProgressRe.FindStringSubmatch(line); m != nil {
			percent, _ := strconv.ParseFloat(m[1], 64)
			j.update(Progress{VideoID: videoID, FileID: file.ID, Stage: FileDownloading, Percent: percent, Speed: m[2], ETA: m[3]}, false)
		}
	})
	if err != nil {
		d.removeTemp(file.ID)
		d.fail(j, videoID, file.ID, cmd.String(), err, output)
		return
	}

	path, err := d.locate(file.ID)
	if err != nil {
		d.fail(j, videoID, file.ID, "locate", err, output)
		return
	}
	d.applyInfo(ctx, videoID, file.ID)

	probe, err := ProbeFile(ctx, path)
	if err != nil {
		d.fail(j, videoID, file.ID, "ffprobe", err, "")
		return
	}
	label := "Original"
	if probe.Height > 0 {
		label += fmt.Sprintf(" · %dp", probe.Height)
	}
	file, err = d.complete(ctx, j, videoID, probe.mediaParams(file.ID, filepath.Base(path), label), true)
	if err != nil {
		return
	}

	rec := req.Resolve(d.caps, probe.Source)
	if rec.Skip {
		return
	}
	keep := d.settings.MustGet(ctx).KeepOriginal
	if _, err := d.enqueueEncode(context.Background(), videoID, file, rec, true, !keep); err != nil {
		d.recordError(videoID, file.ID, "encode", err.Error(), "")
	}
}

func (d *DownloaderService) runEncode(j *job, videoID string, file, source database.VideoFile, rec encoding.Recommendation, makePrimary, removeSource bool) {
	defer d.done(file.ID)
	ctx := j.ctx

	if err := d.encodes.acquire(ctx); err != nil {
		d.fail(j, videoID, file.ID, "ffmpeg", err, "")
		return
	}
	defer d.encodes.release()
	d.setStatus(j, videoID, file.ID, FileEncoding)

	src := SourceOf(source)
	input := d.Path(*source.FileName)
	outName := file.ID + rec.Profile.Extension()
	if src.Duration == 0 {
		if p, err := ProbeFile(ctx, input); err == nil {
			src.Duration = p.Duration
		}
	}

	args, err := rec.Profile.Args(d.caps, src, input, d.Path(outName))
	if err != nil {
		d.fail(j, videoID, file.ID, "ffmpeg", err, "")
		return
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := runStreaming(cmd, func(line string) {
		us, ok := strings.CutPrefix(line, "out_time_us=")
		if !ok || src.Duration <= 0 {
			return
		}
		t, _ := strconv.ParseFloat(us, 64)
		percent := min(t/1e6/src.Duration*100, 100)
		j.update(Progress{VideoID: videoID, FileID: file.ID, Stage: FileEncoding, Percent: percent}, false)
	})
	if err != nil {
		_ = os.Remove(d.Path(outName))
		d.fail(j, videoID, file.ID, cmd.String(), err, output)
		return
	}

	probe, err := ProbeFile(ctx, d.Path(outName))
	if err != nil {
		d.fail(j, videoID, file.ID, "ffprobe", err, "")
		return
	}
	if _, err := d.complete(ctx, j, videoID, probe.mediaParams(file.ID, outName, rec.Label), makePrimary); err != nil {
		return
	}
	if removeSource {
		if _, err := d.DeleteFile(context.Background(), videoID, source.ID); err != nil {
			log.Printf("WARN [%s]: failed to remove source version: %v\n", file.ID, err)
		}
	}
}

// complete stores final media info and optionally makes the version primary.
func (d *DownloaderService) complete(ctx context.Context, j *job, videoID string, params database.UpdateFileMediaParams, makePrimary bool) (database.VideoFile, error) {
	file, err := d.store.UpdateFileMedia(ctx, videoID, params)
	if err != nil {
		d.fail(j, videoID, params.ID, "database", err, "")
		return file, err
	}
	v, err := d.store.Get(ctx, videoID)
	if err == nil && (makePrimary || v.PrimaryFileID == nil) {
		v, err = d.store.SetPrimary(ctx, videoID, &file.ID)
	}
	j.update(Progress{VideoID: videoID, FileID: file.ID, Stage: FileCompleted, Percent: 100}, true)
	if err == nil {
		for _, fn := range d.onFile {
			fn(v, file)
		}
	}
	return file, nil
}

// DeleteFile cancels any running job for a version and removes it from disk.
func (d *DownloaderService) DeleteFile(ctx context.Context, videoID, fileID string) (Video, error) {
	v, err := d.store.Get(ctx, videoID)
	if err != nil {
		return Video{}, err
	}
	file, ok := v.File(fileID)
	if !ok {
		return Video{}, ErrNotFound
	}
	d.Cancel(fileID)
	if v.PrimaryFileID != nil && *v.PrimaryFileID == fileID {
		var next *string
		for _, f := range v.Files {
			if f.ID != fileID && f.Status == FileCompleted {
				next = &f.ID
			}
		}
		if _, err := d.store.SetPrimary(ctx, videoID, next); err != nil {
			return Video{}, err
		}
	}
	v, err = d.store.DeleteFile(ctx, videoID, fileID)
	if err != nil {
		return Video{}, err
	}
	d.removeFile(file)
	if file.FileName != nil {
		d.deleted([]string{*file.FileName})
	}
	return v, nil
}

// DeleteVideo cancels all jobs and removes every file belonging to a video.
func (d *DownloaderService) DeleteVideo(ctx context.Context, id string) (Video, error) {
	v, err := d.store.Get(ctx, id)
	if err != nil {
		return Video{}, err
	}
	for _, f := range v.Files {
		d.Cancel(f.ID)
	}
	if _, err := d.store.Delete(ctx, id); err != nil {
		return Video{}, err
	}
	var names []string
	for _, f := range v.Files {
		d.removeFile(f)
		if f.FileName != nil {
			names = append(names, *f.FileName)
		}
	}
	if v.ThumbnailFileName != nil {
		d.remove(*v.ThumbnailFileName)
		names = append(names, *v.ThumbnailFileName)
	}
	d.deleted(names)
	return v, nil
}

func (d *DownloaderService) setStatus(j *job, videoID, fileID, status string) {
	if err := d.store.UpdateFileStatus(j.ctx, videoID, fileID, status); err != nil {
		log.Printf("WARN [%s]: failed to set status %s: %v\n", fileID, status, err)
	}
	j.update(Progress{VideoID: videoID, FileID: fileID, Stage: status}, true)
}

func (d *DownloaderService) fail(j *job, videoID, fileID, command string, err error, output string) {
	status := FileError
	if j.ctx.Err() != nil {
		status = FileCanceled
	} else {
		log.Printf("ERROR [%s]: %s failed: %v\n", fileID, command, err)
		d.recordError(videoID, fileID, command, errorMessage(err, output), output)
	}
	if err := d.store.UpdateFileStatus(context.Background(), videoID, fileID, status); err != nil && !errors.Is(err, ErrNotFound) {
		log.Printf("WARN [%s]: failed to set status %s: %v\n", fileID, status, err)
	}
	j.update(Progress{VideoID: videoID, FileID: fileID, Stage: status}, true)
}

// errorMessage prefers the last "ERROR:" line a tool printed over a bare exit status.
func errorMessage(err error, output string) string {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if msg, ok := strings.CutPrefix(strings.TrimSpace(lines[i]), "ERROR:"); ok {
			return strings.TrimSpace(msg)
		}
	}
	return err.Error()
}

func (d *DownloaderService) recordError(videoID, fileID, command, message, output string) {
	if len(output) > maxOutput {
		output = output[len(output)-maxOutput:]
	}
	err := d.queries.CreateError(context.Background(), database.CreateErrorParams{
		ID: NewID(), VideoID: &videoID, FileID: &fileID, ErrorMessage: message, Command: command, Output: output,
	})
	if err != nil {
		log.Printf("WARN: failed to record error: %v\n", err)
	}
}

func (d *DownloaderService) locate(fileID string) (string, error) {
	matches, _ := filepath.Glob(filepath.Join(d.dir, fileID+".*"))
	for _, m := range matches {
		ext := strings.ToLower(filepath.Ext(m))
		if !strings.HasSuffix(m, ".info.json") && !slices.Contains(skipDownloadExts, ext) {
			return m, nil
		}
	}
	return "", fmt.Errorf("downloaded video file not found")
}

// applyInfo stores title, uploader and thumbnail written next to the download.
func (d *DownloaderService) applyInfo(ctx context.Context, videoID, fileID string) {
	params := database.UpdateVideoSourceParams{ID: videoID}

	infoPath := filepath.Join(d.dir, fileID+".info.json")
	if data, err := os.ReadFile(infoPath); err == nil {
		var info ytdlpInfo
		if json.Unmarshal(data, &info) == nil {
			params.SourceTitle = nonEmpty(info.Title)
			params.Uploader = nonEmpty(info.Uploader)
			params.Duration = positive(info.Duration)
		}
		_ = os.Remove(infoPath)
	}

	thumb := filepath.Join(d.dir, fileID+".jpg")
	if _, err := os.Stat(thumb); err == nil {
		name := videoID + ".jpg"
		if err := os.Rename(thumb, d.Path(name)); err == nil {
			params.ThumbnailFileName = &name
		}
	}

	if _, err := d.store.UpdateSource(ctx, params); err != nil {
		log.Printf("WARN [%s]: failed to store source info: %v\n", fileID, err)
	}
}

func (d *DownloaderService) removeTemp(fileID string) {
	matches, _ := filepath.Glob(filepath.Join(d.dir, fileID+".*"))
	for _, m := range matches {
		_ = os.Remove(m)
	}
}

func (d *DownloaderService) removeFile(f database.VideoFile) {
	if f.FileName != nil {
		d.remove(*f.FileName)
	}
	d.removeTemp(f.ID)
}

func (d *DownloaderService) remove(name string) {
	if err := os.Remove(d.Path(name)); err != nil && !os.IsNotExist(err) {
		log.Printf("WARN: failed to delete %s: %v\n", name, err)
	}
}

// runStreaming runs cmd, feeding stdout lines to onLine, and returns the
// tail of combined output for error reporting. Cancelling the command's
// context kills its whole process tree.
func runStreaming(cmd *exec.Cmd, onLine func(string)) (string, error) {
	tail := &tailBuffer{}
	cmd.Stdout = &lineWriter{tail: tail, onLine: onLine}
	cmd.Stderr = &lineWriter{tail: tail}
	cmd.WaitDelay = 5 * time.Second
	killTree(cmd)
	err := cmd.Run()
	return tail.String(), err
}

type tailBuffer struct {
	mu  sync.Mutex
	buf []byte
}

func (t *tailBuffer) add(line []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(append(t.buf, line...), '\n')
	if len(t.buf) > maxOutput*2 {
		t.buf = append([]byte(nil), t.buf[len(t.buf)-maxOutput:]...)
	}
}

func (t *tailBuffer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.buf)
}

// lineWriter splits process output into lines; ffmpeg and yt-dlp may use CR.
type lineWriter struct {
	tail    *tailBuffer
	onLine  func(string)
	pending []byte
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.pending = append(w.pending, p...)
	for {
		i := bytes.IndexAny(w.pending, "\r\n")
		if i < 0 {
			break
		}
		if line := w.pending[:i]; len(line) > 0 {
			w.tail.add(line)
			if w.onLine != nil {
				w.onLine(string(line))
			}
		}
		w.pending = w.pending[i+1:]
	}
	return len(p), nil
}
