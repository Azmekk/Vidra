package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"slices"
	"sync"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/google/uuid"
)

type FileStatus = string

const (
	FileQueued      FileStatus = "queued"
	FileDownloading FileStatus = "downloading"
	FileEncoding    FileStatus = "encoding"
	FileCompleted   FileStatus = "completed"
	FileError       FileStatus = "error"
	FileCanceled    FileStatus = "canceled"
)

const (
	KindOriginal = "original"
	KindEncode   = "encode"
)

var ErrNotFound = errors.New("not found")

type VideoFile = database.VideoFile

// Video is a logical video together with all of its file versions.
type Video struct {
	database.Video
	Files []VideoFile
}

func (v Video) clone() Video {
	v.Files = slices.Clone(v.Files)
	return v
}

func (v Video) File(id string) (database.VideoFile, bool) {
	for _, f := range v.Files {
		if f.ID == id {
			return f, true
		}
	}
	return database.VideoFile{}, false
}

func (v Video) Primary() (database.VideoFile, bool) {
	if v.PrimaryFileID == nil {
		return database.VideoFile{}, false
	}
	return v.File(*v.PrimaryFileID)
}

func (v Video) Original() (database.VideoFile, bool) {
	for _, f := range v.Files {
		if f.Kind == KindOriginal && f.Status == FileCompleted {
			return f, true
		}
	}
	return database.VideoFile{}, false
}

// Status summarises the video: an in-flight job wins, then the primary version.
func (v Video) Status() string {
	status := ""
	for _, f := range v.Files {
		switch f.Status {
		case FileDownloading, FileEncoding:
			return f.Status
		case FileQueued:
			status = FileQueued
		}
	}
	if status != "" {
		return status
	}
	if p, ok := v.Primary(); ok {
		return p.Status
	}
	if n := len(v.Files); n > 0 {
		return v.Files[n-1].Status
	}
	return FileQueued
}

func NewID() string { return uuid.Must(uuid.NewV7()).String() }

type ListQuery struct {
	Search string
	Order  string
	Offset int
	Limit  int
}

func (q ListQuery) defaultOrder() bool {
	return q.Search == "" && (q.Order == "" || q.Order == "created_at_desc")
}

// VideoStore owns all video persistence and keeps the newest videos in
// memory so the default library view never touches the disk.
type VideoStore struct {
	queries *database.Queries
	ws      *WebSocketService

	mu     sync.RWMutex
	size   int
	recent []Video
	total  int64
}

func NewVideoStore(queries *database.Queries, ws *WebSocketService) *VideoStore {
	return &VideoStore{queries: queries, ws: ws}
}

// Warm (re)loads the cache with the newest size videos.
func (s *VideoStore) Warm(ctx context.Context, size int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.size = size
	return s.warmLocked(ctx)
}

func (s *VideoStore) warmLocked(ctx context.Context) error {
	total, err := s.queries.CountVideos(ctx, "")
	if err != nil {
		return err
	}
	rows, err := s.queries.ListVideos(ctx, database.ListVideosParams{Limit: int64(s.size)})
	if err != nil {
		return err
	}
	videos, err := s.attach(ctx, rows)
	if err != nil {
		return err
	}
	s.recent, s.total = videos, total
	log.Printf("Video cache warmed with %d of %d videos\n", len(videos), total)
	return nil
}

func (s *VideoStore) List(ctx context.Context, q ListQuery) ([]Video, int64, error) {
	if q.defaultOrder() {
		s.mu.RLock()
		complete := int64(len(s.recent)) == s.total
		if q.Offset+q.Limit <= len(s.recent) || complete {
			end := min(q.Offset+q.Limit, len(s.recent))
			var page []Video
			if q.Offset < end {
				page = make([]Video, 0, end-q.Offset)
				for _, v := range s.recent[q.Offset:end] {
					page = append(page, v.clone())
				}
			}
			total := s.total
			s.mu.RUnlock()
			return page, total, nil
		}
		s.mu.RUnlock()
	}

	total, err := s.queries.CountVideos(ctx, q.Search)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.queries.ListVideos(ctx, database.ListVideosParams{
		Search: q.Search, Ordering: q.Order, Limit: int64(q.Limit), Offset: int64(q.Offset),
	})
	if err != nil {
		return nil, 0, err
	}
	videos, err := s.attach(ctx, rows)
	return videos, total, err
}

func (s *VideoStore) Get(ctx context.Context, id string) (Video, error) {
	s.mu.RLock()
	if i := s.index(id); i >= 0 {
		v := s.recent[i].clone()
		s.mu.RUnlock()
		return v, nil
	}
	s.mu.RUnlock()
	return s.load(ctx, id)
}

// GetByFile finds the video owning a file version.
func (s *VideoStore) GetByFile(ctx context.Context, fileID string) (Video, database.VideoFile, error) {
	s.mu.RLock()
	for _, v := range s.recent {
		if f, ok := v.File(fileID); ok {
			v := v.clone()
			s.mu.RUnlock()
			return v, f, nil
		}
	}
	s.mu.RUnlock()

	f, err := s.queries.GetVideoFile(ctx, fileID)
	if err != nil {
		return Video{}, database.VideoFile{}, notFound(err)
	}
	v, err := s.load(ctx, f.VideoID)
	return v, f, err
}

func (s *VideoStore) Create(ctx context.Context, name string, sourceTitle *string, url string) (Video, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, err := s.queries.CreateVideo(ctx, database.CreateVideoParams{
		ID: NewID(), Name: name, SourceTitle: sourceTitle, OriginalUrl: url,
	})
	if err != nil {
		return Video{}, err
	}
	v := Video{Video: row}
	s.total++
	if s.size > 0 {
		s.recent = slices.Insert(s.recent, 0, v)
		if len(s.recent) > s.size {
			s.recent = s.recent[:s.size]
		}
	}
	s.ws.Broadcast(WsEventVideoCreated, ToVideoDTO(v))
	return v.clone(), nil
}

func (s *VideoStore) Rename(ctx context.Context, id, name string) (Video, error) {
	return s.mutate(ctx, id, func(q *database.Queries) error {
		_, err := q.UpdateVideoName(ctx, database.UpdateVideoNameParams{ID: id, Name: name})
		return err
	})
}

func (s *VideoStore) UpdateSource(ctx context.Context, p database.UpdateVideoSourceParams) (Video, error) {
	return s.mutate(ctx, p.ID, func(q *database.Queries) error {
		_, err := q.UpdateVideoSource(ctx, p)
		return err
	})
}

func (s *VideoStore) SetPrimary(ctx context.Context, videoID string, fileID *string) (Video, error) {
	return s.mutate(ctx, videoID, func(q *database.Queries) error {
		_, err := q.SetPrimaryFile(ctx, database.SetPrimaryFileParams{ID: videoID, PrimaryFileID: fileID})
		return err
	})
}

func (s *VideoStore) CreateFile(ctx context.Context, p database.CreateVideoFileParams) (database.VideoFile, error) {
	p.ID = NewID()
	var file database.VideoFile
	_, err := s.mutate(ctx, p.VideoID, func(q *database.Queries) (err error) {
		file, err = q.CreateVideoFile(ctx, p)
		return err
	})
	return file, err
}

func (s *VideoStore) UpdateFileStatus(ctx context.Context, videoID, fileID, status string) error {
	_, err := s.mutate(ctx, videoID, func(q *database.Queries) error {
		_, err := q.UpdateFileStatus(ctx, database.UpdateFileStatusParams{ID: fileID, Status: status})
		return err
	})
	return err
}

func (s *VideoStore) UpdateFileMedia(ctx context.Context, videoID string, p database.UpdateFileMediaParams) (database.VideoFile, error) {
	var file database.VideoFile
	_, err := s.mutate(ctx, videoID, func(q *database.Queries) (err error) {
		file, err = q.UpdateFileMedia(ctx, p)
		return err
	})
	return file, err
}

func (s *VideoStore) DeleteFile(ctx context.Context, videoID, fileID string) (Video, error) {
	return s.mutate(ctx, videoID, func(q *database.Queries) error {
		return q.DeleteVideoFile(ctx, fileID)
	})
}

func (s *VideoStore) Delete(ctx context.Context, id string) (Video, error) {
	v, err := s.Get(ctx, id)
	if err != nil {
		return Video{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.queries.DeleteVideo(ctx, id); err != nil {
		return Video{}, err
	}
	if i := s.index(id); i >= 0 {
		if err := s.warmLocked(ctx); err != nil {
			s.recent = slices.Delete(s.recent, i, i+1)
			s.total--
		}
	} else {
		s.total--
	}
	s.ws.Broadcast(WsEventVideoDeleted, map[string]string{"id": id})
	return v, nil
}

// MarkInterrupted flags jobs that were running when the server stopped.
func (s *VideoStore) MarkInterrupted(ctx context.Context) ([]database.VideoFile, error) {
	files, err := s.queries.MarkInterruptedFiles(ctx)
	if err != nil || len(files) == 0 {
		return files, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return files, s.warmLocked(ctx)
}

func (s *VideoStore) mutate(ctx context.Context, videoID string, fn func(q *database.Queries) error) (Video, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fn(s.queries); err != nil {
		return Video{}, notFound(err)
	}
	v, err := s.load(ctx, videoID)
	if err != nil {
		return Video{}, err
	}
	if i := s.index(videoID); i >= 0 {
		s.recent[i] = v
	}
	s.ws.Broadcast(WsEventVideoUpdated, ToVideoDTO(v))
	return v.clone(), nil
}

func (s *VideoStore) index(id string) int {
	return slices.IndexFunc(s.recent, func(v Video) bool { return v.ID == id })
}

func (s *VideoStore) load(ctx context.Context, id string) (Video, error) {
	row, err := s.queries.GetVideo(ctx, id)
	if err != nil {
		return Video{}, notFound(err)
	}
	videos, err := s.attach(ctx, []database.Video{row})
	if err != nil {
		return Video{}, err
	}
	return videos[0], nil
}

func (s *VideoStore) attach(ctx context.Context, rows []database.Video) ([]Video, error) {
	videos := make([]Video, len(rows))
	if len(rows) == 0 {
		return videos, nil
	}
	ids := make([]string, len(rows))
	byID := make(map[string]int, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
		byID[r.ID] = i
		videos[i] = Video{Video: r}
	}
	files, err := s.queries.ListFilesByVideoIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		i := byID[f.VideoID]
		videos[i].Files = append(videos[i].Files, f)
	}
	return videos, nil
}

func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
