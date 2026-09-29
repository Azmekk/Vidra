package services

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/utils"
)

// FileRename is a file renamed inside the downloads directory.
type FileRename struct {
	From string
	To   string
}

type FileNameSyncResult struct {
	Videos  int `json:"videos"`
	Renamed int `json:"renamed"`
	Missing int `json:"missing"`
}

// VersionFileBase is the file name, without extension, a version is stored under.
func VersionFileBase(v Video, f database.VideoFile) string {
	stored := 0
	for _, f := range v.Files {
		if f.Status == FileCompleted && f.FileName != nil {
			stored++
		}
	}
	if stored > 1 {
		return utils.SanitizeFilename(fmt.Sprintf("%s (%s)", v.Name, strings.ReplaceAll(f.Label, " · ", " ")))
	}
	return utils.SanitizeFilename(v.Name)
}

// OnFilesRenamed registers a callback with files renamed on disk.
func (d *DownloaderService) OnFilesRenamed(fn func([]FileRename)) {
	d.onRename = append(d.onRename, fn)
}

func (d *DownloaderService) renamed(renames []FileRename) {
	if len(renames) == 0 {
		return
	}
	for _, fn := range d.onRename {
		fn(renames)
	}
}

// SyncVideoFileNames renames a video's files to match its current name.
func (d *DownloaderService) SyncVideoFileNames(ctx context.Context, videoID string) (Video, error) {
	return d.syncVideo(ctx, videoID, "")
}

// syncVideo renames a video's files, leaving fresh (a file that was never backed up) out of the rename callbacks.
func (d *DownloaderService) syncVideo(ctx context.Context, videoID, fresh string) (Video, error) {
	d.names.Lock()
	defer d.names.Unlock()
	v, renames, err := d.syncLocked(ctx, videoID, fresh)
	d.renamed(renames)
	return v, err
}

func (d *DownloaderService) syncLocked(ctx context.Context, videoID, fresh string) (Video, []FileRename, error) {
	taken, err := d.takenNames(ctx)
	if err != nil {
		return Video{}, nil, err
	}
	v, renames, _, err := d.syncNames(ctx, videoID, taken)
	return v, slices.DeleteFunc(renames, func(r FileRename) bool { return r.From == fresh }), err
}

// SyncFileNames renames every stored file to match its video's current name.
func (d *DownloaderService) SyncFileNames(ctx context.Context) (FileNameSyncResult, error) {
	d.names.Lock()
	defer d.names.Unlock()
	ids, err := d.queries.ListVideoIDs(ctx)
	if err != nil {
		return FileNameSyncResult{}, err
	}
	taken, err := d.takenNames(ctx)
	if err != nil {
		return FileNameSyncResult{}, err
	}
	var res FileNameSyncResult
	var all []FileRename
	defer func() { d.renamed(all) }()
	for _, id := range ids {
		_, renames, missing, err := d.syncNames(ctx, id, taken)
		all = append(all, renames...)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return res, err
		}
		res.Videos++
		res.Renamed += len(renames)
		res.Missing += missing
	}
	slog.Info("file names synced", "videos", res.Videos, "renamed", res.Renamed, "missing", res.Missing)
	return res, nil
}

// syncNames renames the files of one video on disk and in the database. taken
// holds every stored file name in lower case and is kept up to date.
func (d *DownloaderService) syncNames(ctx context.Context, videoID string, taken map[string]bool) (Video, []FileRename, int, error) {
	v, err := d.store.Get(ctx, videoID)
	if err != nil {
		return Video{}, nil, 0, err
	}
	var done []FileRename
	missing := 0
	rename := func(current, base string) (string, bool) {
		if _, err := os.Stat(d.Path(current)); err != nil {
			slog.Warn("stored file is missing", "video", videoID, "name", current)
			missing++
			return "", false
		}
		ext := filepath.Ext(current)
		if fitsBase(current, base, ext) {
			return "", false
		}
		name := d.freeName(base, ext, current, taken)
		if err := os.Rename(d.Path(current), d.Path(name)); err != nil {
			slog.Warn("failed to rename file", "video", videoID, "from", current, "to", name, "error", err)
			return "", false
		}
		delete(taken, strings.ToLower(current))
		taken[strings.ToLower(name)] = true
		done = append(done, FileRename{From: current, To: name})
		return name, true
	}

	files := map[string]string{}
	for _, f := range v.Files {
		if f.Status != FileCompleted || f.FileName == nil || d.pinned(f.ID) {
			continue
		}
		if name, ok := rename(*f.FileName, VersionFileBase(v, f)); ok {
			files[f.ID] = name
		}
	}
	var thumbnail *string
	if v.ThumbnailFileName != nil {
		if name, ok := rename(*v.ThumbnailFileName, utils.SanitizeFilename(v.Name)); ok {
			thumbnail = &name
		}
	}
	if len(done) == 0 {
		return v, nil, missing, nil
	}

	updated, err := d.store.SetFileNames(ctx, videoID, files, thumbnail)
	if err != nil {
		for _, r := range slices.Backward(done) {
			if e := os.Rename(d.Path(r.To), d.Path(r.From)); e != nil {
				slog.Error("failed to restore file name", "from", r.To, "to", r.From, "error", e)
			}
			delete(taken, strings.ToLower(r.To))
			taken[strings.ToLower(r.From)] = true
		}
		return v, nil, missing, err
	}
	return updated, done, missing, nil
}

func (d *DownloaderService) takenNames(ctx context.Context) (map[string]bool, error) {
	names, err := d.queries.ListStoredFileNames(ctx)
	if err != nil {
		return nil, err
	}
	taken := make(map[string]bool, len(names))
	for _, n := range names {
		if n != nil {
			taken[strings.ToLower(*n)] = true
		}
	}
	return taken, nil
}

// freeName picks base+ext, or "base (n)"+ext when that is taken by another file.
func (d *DownloaderService) freeName(base, ext, current string, taken map[string]bool) string {
	for i := 1; ; i++ {
		name := base + ext
		if i > 1 {
			name = fmt.Sprintf("%s (%d)%s", base, i, ext)
		}
		if strings.EqualFold(name, current) {
			return name
		}
		if taken[strings.ToLower(name)] {
			continue
		}
		if _, err := os.Lstat(d.Path(name)); errors.Is(err, fs.ErrNotExist) {
			return name
		}
	}
}

// fitsBase reports whether name is already base+ext or "base (n)"+ext.
func fitsBase(name, base, ext string) bool {
	if name == base+ext {
		return true
	}
	rest, ok := strings.CutPrefix(name, base+" (")
	if !ok {
		return false
	}
	n, ok := strings.CutSuffix(rest, ")"+ext)
	if !ok {
		return false
	}
	i, err := strconv.Atoi(n)
	return err == nil && i > 1 && strconv.Itoa(i) == n
}

// pinSource marks a version as in use by an encode so it is not renamed
// underneath ffmpeg, and returns its current state.
func (d *DownloaderService) pinSource(ctx context.Context, videoID, fileID string) (database.VideoFile, func(), error) {
	d.names.Lock()
	defer d.names.Unlock()
	v, err := d.store.Get(ctx, videoID)
	if err != nil {
		return database.VideoFile{}, nil, err
	}
	f, ok := v.File(fileID)
	if !ok || f.Status != FileCompleted || f.FileName == nil {
		return database.VideoFile{}, nil, ErrFileNotReady
	}
	d.pins.Lock()
	d.pinCount[fileID]++
	d.pins.Unlock()
	var once bool
	return f, func() {
		d.pins.Lock()
		defer d.pins.Unlock()
		if once {
			return
		}
		once = true
		if d.pinCount[fileID]--; d.pinCount[fileID] <= 0 {
			delete(d.pinCount, fileID)
		}
	}, nil
}

func (d *DownloaderService) pinned(fileID string) bool {
	d.pins.Lock()
	defer d.pins.Unlock()
	return d.pinCount[fileID] > 0
}
