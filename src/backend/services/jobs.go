package services

import (
	"context"
	"slices"
	"sync"
	"time"
)

// pool is a counting semaphore whose limit can change at runtime.
type pool struct {
	mu      sync.Mutex
	limit   int
	active  int
	waiters []chan struct{}
}

func newPool(limit int) *pool { return &pool{limit: max(limit, 1)} }

func (p *pool) acquire(ctx context.Context) error {
	p.mu.Lock()
	if p.active < p.limit {
		p.active++
		p.mu.Unlock()
		return nil
	}
	ch := make(chan struct{})
	p.waiters = append(p.waiters, ch)
	p.mu.Unlock()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		p.mu.Lock()
		if i := slices.Index(p.waiters, ch); i >= 0 {
			p.waiters = slices.Delete(p.waiters, i, i+1)
			p.mu.Unlock()
			return ctx.Err()
		}
		p.mu.Unlock()
		p.release()
		return ctx.Err()
	}
}

func (p *pool) release() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active--
	p.grant()
}

func (p *pool) setLimit(limit int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.limit = max(limit, 1)
	p.grant()
}

func (p *pool) grant() {
	for p.active < p.limit && len(p.waiters) > 0 {
		ch := p.waiters[0]
		p.waiters = p.waiters[1:]
		p.active++
		close(ch)
	}
}

type Progress struct {
	VideoID string  `json:"videoId"`
	FileID  string  `json:"fileId"`
	Stage   string  `json:"stage"`
	Percent float64 `json:"percent"`
	Speed   string  `json:"speed,omitempty"`
	ETA     string  `json:"eta,omitempty"`
}

const progressInterval = 250 * time.Millisecond

type job struct {
	ctx    context.Context
	cancel context.CancelFunc
	ws     *WebSocketService

	mu       sync.Mutex
	progress Progress
	sentAt   time.Time
}

// update records progress and broadcasts it, throttled unless forced.
func (j *job) update(p Progress, force bool) {
	j.mu.Lock()
	j.progress = p
	send := force || time.Since(j.sentAt) >= progressInterval
	if send {
		j.sentAt = time.Now()
	}
	j.mu.Unlock()
	if send {
		j.ws.Broadcast(WsEventFileProgress, p)
	}
}

func (j *job) snapshot() Progress {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.progress
}

type jobRegistry struct {
	jobs sync.Map
	ws   *WebSocketService
}

func (r *jobRegistry) start(videoID, fileID string) *job {
	ctx, cancel := context.WithCancel(context.Background())
	j := &job{ctx: ctx, cancel: cancel, ws: r.ws, progress: Progress{VideoID: videoID, FileID: fileID, Stage: FileQueued}}
	r.jobs.Store(fileID, j)
	return j
}

func (r *jobRegistry) done(fileID string) {
	if v, ok := r.jobs.LoadAndDelete(fileID); ok {
		v.(*job).cancel()
	}
}

func (r *jobRegistry) Cancel(fileID string) bool {
	v, ok := r.jobs.Load(fileID)
	if ok {
		v.(*job).cancel()
	}
	return ok
}

func (r *jobRegistry) All() []Progress {
	var out []Progress
	r.jobs.Range(func(_, v any) bool {
		out = append(out, v.(*job).snapshot())
		return true
	})
	return out
}
