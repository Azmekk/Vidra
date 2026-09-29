package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Azmekk/Vidra/backend/gen/database"
)

const (
	BackupProviderS3    = "s3"
	BackupProviderDrive = "drive"
	BackupProviderMega  = "mega"
	BackupProviderLocal = "local"

	backupSecretMask  = "••••••••"
	backupSnapshots   = 7
	backupDebounce    = 5 * time.Minute
	backupTimeout     = 6 * time.Hour
	backupMaxInterval = 24 * 30
)

type backupField struct {
	Key     string
	Secret  bool
	Obscure bool
}

// backupProviders lists the rclone options each provider accepts from the UI,
// plus fixed options that are always set.
var backupProviders = map[string]struct {
	Type   string
	Fields []backupField
	Fixed  map[string]string
}{
	BackupProviderS3: {Type: "s3", Fields: []backupField{
		{Key: "provider"}, {Key: "endpoint"}, {Key: "region"}, {Key: "access_key_id"}, {Key: "secret_access_key", Secret: true},
	}, Fixed: map[string]string{"env_auth": "false", "no_check_bucket": "true"}},
	BackupProviderDrive: {Type: "drive", Fields: []backupField{
		{Key: "client_id"}, {Key: "client_secret", Secret: true}, {Key: "token", Secret: true}, {Key: "root_folder_id"},
	}, Fixed: map[string]string{"scope": "drive"}},
	BackupProviderMega: {Type: "mega", Fields: []backupField{
		{Key: "user"}, {Key: "pass", Secret: true, Obscure: true},
	}},
	BackupProviderLocal: {Type: "local"},
}

var ErrRcloneMissing = errors.New("rclone is not installed")

type BackupTargetInput struct {
	Name            string            `json:"name"`
	Provider        string            `json:"provider"`
	Config          map[string]string `json:"config"`
	Path            string            `json:"path"`
	IncludeVideos   bool              `json:"includeVideos"`
	IncludeDatabase bool              `json:"includeDatabase"`
	Enabled         bool              `json:"enabled"`
	IntervalHours   int               `json:"intervalHours"`
}

type BackupTargetDTO struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Provider        string            `json:"provider"`
	Config          map[string]string `json:"config"`
	Path            string            `json:"path"`
	IncludeVideos   bool              `json:"includeVideos"`
	IncludeDatabase bool              `json:"includeDatabase"`
	Enabled         bool              `json:"enabled"`
	IntervalHours   int               `json:"intervalHours"`
	LastRunAt       *string           `json:"lastRunAt,omitempty"`
	LastFullAt      *string           `json:"lastFullAt,omitempty"`
	LastStatus      string            `json:"lastStatus"`
	LastError       string            `json:"lastError"`
	CreatedAt       string            `json:"createdAt"`
}

type backupJob struct {
	targetID string
	upload   []string
	remove   []string
	moves    []FileRename
	snapshot bool
	full     bool
}

type BackupService struct {
	queries      *database.Queries
	db           *sql.DB
	ws           *WebSocketService
	downloadsDir string
	rclone       string
	version      string

	jobs    chan backupJob
	mu      sync.Mutex
	dirty   bool
	running map[string]bool
	queued  map[string]bool
}

func NewBackupService(queries *database.Queries, db *sql.DB, ws *WebSocketService, downloadsDir string) *BackupService {
	s := &BackupService{
		queries: queries, db: db, ws: ws, downloadsDir: downloadsDir,
		jobs: make(chan backupJob, 256), running: map[string]bool{}, queued: map[string]bool{},
	}
	if bin, err := exec.LookPath("rclone"); err == nil {
		s.rclone = bin
		if out, err := exec.Command(bin, "version").Output(); err == nil {
			s.version = strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
		}
	}
	return s
}

func (s *BackupService) Available() bool { return s.rclone != "" }
func (s *BackupService) Version() string { return s.version }

// Start runs the upload worker and the snapshot scheduler until ctx ends.
func (s *BackupService) Start(ctx context.Context) {
	if !s.Available() {
		log.Println("rclone not found, backups are disabled")
		return
	}
	for _, t := range s.mustList(ctx) {
		if t.LastStatus == "running" {
			_, _ = s.queries.SetBackupTargetStatus(ctx, database.SetBackupTargetStatusParams{
				ID: t.ID, LastStatus: "error", LastError: "interrupted by a restart",
			})
		}
	}
	go s.worker(ctx)
	go s.scheduler(ctx)
}

func (s *BackupService) FileCompleted(v Video, f database.VideoFile) {
	if f.FileName == nil {
		return
	}
	names := []string{*f.FileName}
	if v.ThumbnailFileName != nil {
		names = append(names, *v.ThumbnailFileName)
	}
	s.markDirty()
	s.enqueue(backupJob{upload: names})
}

func (s *BackupService) FilesDeleted(names []string) {
	s.markDirty()
	s.enqueue(backupJob{remove: names})
}

// FilesRenamed moves renamed files on every remote.
func (s *BackupService) FilesRenamed(renames []FileRename) {
	s.markDirty()
	s.enqueue(backupJob{moves: renames})
}

func (s *BackupService) Run(id string) error {
	if !s.Available() {
		return ErrRcloneMissing
	}
	return s.queueFull(id)
}

// queueFull enqueues a full backup of one target unless one is already queued or running.
func (s *BackupService) queueFull(id string) error {
	s.mu.Lock()
	if s.running[id] || s.queued[id] {
		s.mu.Unlock()
		return errors.New("a backup is already running for this target")
	}
	s.queued[id] = true
	s.mu.Unlock()
	if !s.enqueue(backupJob{targetID: id, full: true, snapshot: true}) {
		s.mu.Lock()
		delete(s.queued, id)
		s.mu.Unlock()
		return errors.New("the backup queue is full, try again later")
	}
	return nil
}

// Test checks that the remote is reachable and writable.
func (s *BackupService) Test(ctx context.Context, id string) error {
	if !s.Available() {
		return ErrRcloneMissing
	}
	t, err := s.queries.GetBackupTarget(ctx, id)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err := s.exec(ctx, t, nil, "mkdir", s.remote(t, "")); err != nil {
		return err
	}
	_, err = s.exec(ctx, t, nil, "lsf", "--max-depth", "1", s.remote(t, ""))
	return err
}

func (s *BackupService) List(ctx context.Context) ([]BackupTargetDTO, error) {
	rows, err := s.queries.ListBackupTargets(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]BackupTargetDTO, len(rows))
	for i, t := range rows {
		out[i] = toBackupDTO(t)
	}
	return out, nil
}

func (s *BackupService) Create(ctx context.Context, in BackupTargetInput) (BackupTargetDTO, error) {
	cfg, err := s.prepareConfig(in, nil)
	if err != nil {
		return BackupTargetDTO{}, err
	}
	t, err := s.queries.CreateBackupTarget(ctx, database.CreateBackupTargetParams{
		ID: NewID(), Name: strings.TrimSpace(in.Name), Provider: in.Provider, Config: cfg,
		Path: cleanRemotePath(in.Path), IncludeVideos: in.IncludeVideos, IncludeDatabase: in.IncludeDatabase, Enabled: in.Enabled,
		IntervalHours: int64(in.IntervalHours),
	})
	if err != nil {
		return BackupTargetDTO{}, err
	}
	return toBackupDTO(t), nil
}

func (s *BackupService) Update(ctx context.Context, id string, in BackupTargetInput) (BackupTargetDTO, error) {
	existing, err := s.queries.GetBackupTarget(ctx, id)
	if err != nil {
		return BackupTargetDTO{}, err
	}
	in.Provider = existing.Provider
	cfg, err := s.prepareConfig(in, decodeConfig(existing.Config))
	if err != nil {
		return BackupTargetDTO{}, err
	}
	t, err := s.queries.UpdateBackupTarget(ctx, database.UpdateBackupTargetParams{
		ID: id, Name: strings.TrimSpace(in.Name), Config: cfg, Path: cleanRemotePath(in.Path),
		IncludeVideos: in.IncludeVideos, IncludeDatabase: in.IncludeDatabase, Enabled: in.Enabled,
		IntervalHours: int64(in.IntervalHours),
	})
	if err != nil {
		return BackupTargetDTO{}, err
	}
	return toBackupDTO(t), nil
}

func (s *BackupService) Delete(ctx context.Context, id string) error {
	return s.queries.DeleteBackupTarget(ctx, id)
}

func (s *BackupService) prepareConfig(in BackupTargetInput, previous map[string]string) (string, error) {
	spec, ok := backupProviders[in.Provider]
	if !ok {
		return "", fmt.Errorf("unknown provider %q", in.Provider)
	}
	if strings.TrimSpace(in.Name) == "" {
		return "", errors.New("name is required")
	}
	if in.Provider == BackupProviderLocal && !filepath.IsAbs(in.Path) {
		return "", errors.New("local backups need an absolute folder path")
	}
	if in.IntervalHours < 0 || in.IntervalHours > backupMaxInterval {
		return "", fmt.Errorf("interval must be between 0 and %d hours", backupMaxInterval)
	}
	cfg := map[string]string{}
	for _, f := range spec.Fields {
		v := strings.TrimSpace(in.Config[f.Key])
		if f.Secret && (v == backupSecretMask || v == "") && previous != nil {
			v = previous[f.Key]
		} else if f.Obscure && v != "" {
			if !s.Available() {
				return "", ErrRcloneMissing
			}
			out, err := exec.Command(s.rclone, "obscure", v).Output()
			if err != nil {
				return "", fmt.Errorf("rclone obscure: %w", err)
			}
			v = strings.TrimSpace(string(out))
		}
		if v != "" {
			cfg[f.Key] = v
		}
	}
	b, err := json.Marshal(cfg)
	return string(b), err
}

func (s *BackupService) enqueue(j backupJob) bool {
	if !s.Available() {
		return false
	}
	select {
	case s.jobs <- j:
		return true
	default:
		slog.Warn("backup queue is full, dropping job")
		return false
	}
}

func (s *BackupService) markDirty() {
	s.mu.Lock()
	s.dirty = true
	s.mu.Unlock()
}

func (s *BackupService) scheduler(ctx context.Context) {
	tick := time.NewTicker(backupDebounce)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			s.mu.Lock()
			dirty := s.dirty
			s.dirty = false
			s.mu.Unlock()
			if dirty {
				s.enqueue(backupJob{snapshot: true})
			}
			s.queueDue(ctx)
		}
	}
}

// queueDue starts a full backup of every enabled target whose interval has passed.
func (s *BackupService) queueDue(ctx context.Context) {
	for _, t := range s.mustList(ctx) {
		if !t.Enabled || t.IntervalHours <= 0 {
			continue
		}
		if t.LastFullAt != nil {
			last, err := time.Parse(time.RFC3339, *t.LastFullAt)
			if err == nil && time.Since(last) < time.Duration(t.IntervalHours)*time.Hour {
				continue
			}
		}
		_ = s.queueFull(t.ID)
	}
}

func (s *BackupService) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case j := <-s.jobs:
			s.process(ctx, j)
		}
	}
}

func (s *BackupService) process(ctx context.Context, j backupJob) {
	var targets []database.BackupTarget
	for _, t := range s.mustList(ctx) {
		if (j.targetID == "" && t.Enabled) || t.ID == j.targetID {
			targets = append(targets, t)
		}
	}
	var snapshot string
	if j.snapshot && slices.ContainsFunc(targets, func(t database.BackupTarget) bool { return t.IncludeDatabase }) {
		path, err := s.snapshotDB(ctx)
		if err != nil {
			s.recordError("database snapshot", err, "")
		} else {
			snapshot = path
			defer os.Remove(path)
		}
	}
	for _, t := range targets {
		s.runTarget(ctx, t, j, snapshot)
	}
}

func (s *BackupService) runTarget(ctx context.Context, t database.BackupTarget, j backupJob, snapshot string) {
	s.mu.Lock()
	s.running[t.ID] = true
	if j.full {
		delete(s.queued, t.ID)
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.running, t.ID)
		s.mu.Unlock()
	}()
	started := time.Now()
	if j.full {
		slog.Info("backup started", "target", t.Name)
		s.setStatus(ctx, t.ID, "running", "")
	}

	ctx, cancel := context.WithTimeout(ctx, backupTimeout)
	defer cancel()

	var errs []string
	fail := func(step string, err error, output string) {
		errs = append(errs, fmt.Sprintf("%s: %v", step, err))
		s.recordError("rclone "+step+" ("+t.Name+")", err, output)
	}

	uploaded, moved := 0, 0
	if t.IncludeVideos {
		upload, remove := slices.Clone(j.upload), slices.Clone(j.remove)
		for _, m := range j.moves {
			if _, err := s.exec(ctx, t, nil, "moveto", s.remote(t, "downloads/"+m.From), s.remote(t, "downloads/"+m.To)); err != nil {
				upload = append(upload, m.To)
				remove = append(remove, m.From)
			} else {
				moved++
			}
		}
		if j.full {
			names, err := s.queries.ListStoredFileNames(ctx)
			if err != nil {
				fail("list files", err, "")
			}
			for _, n := range names {
				if n != nil {
					upload = append(upload, *n)
				}
			}
		}
		upload = slices.DeleteFunc(upload, func(name string) bool {
			_, err := os.Stat(filepath.Join(s.downloadsDir, name))
			return err != nil
		})
		if len(upload) > 0 {
			args := []string{"copy", "--files-from-raw", "-", s.downloadsDir, s.remote(t, "downloads")}
			if !j.full {
				args = append(args, "--no-traverse")
			}
			if out, err := s.exec(ctx, t, fileList(upload), args...); err != nil {
				fail("upload", err, out)
			} else {
				uploaded = len(upload)
			}
		}
		if len(remove) > 0 {
			if out, err := s.exec(ctx, t, fileList(remove), "delete", "--files-from-raw", "-", s.remote(t, "downloads")); err != nil {
				fail("delete", err, out)
			}
		}
	}

	if t.IncludeDatabase && snapshot != "" {
		if err := s.uploadSnapshot(ctx, t, snapshot); err != nil {
			fail("database", err, "")
		}
	}

	if j.full {
		if err := s.queries.SetBackupTargetFullRun(context.Background(), t.ID); err != nil {
			slog.Warn("failed to record full backup", "error", err)
		}
	}
	worked := j.full || (t.IncludeDatabase && snapshot != "") || (t.IncludeVideos && (len(j.upload) > 0 || len(j.remove) > 0 || len(j.moves) > 0))
	switch {
	case len(errs) > 0:
		s.setStatus(context.Background(), t.ID, "error", strings.Join(errs, "; "))
	case worked:
		s.setStatus(context.Background(), t.ID, "ok", "")
		slog.Info("backup finished", "target", t.Name, "full", j.full, "uploaded", uploaded,
			"moved", moved, "removed", len(j.remove), "database", t.IncludeDatabase && snapshot != "", "took", since(started))
	}
}

func (s *BackupService) uploadSnapshot(ctx context.Context, t database.BackupTarget, snapshot string) error {
	stamp := time.Now().UTC().Format("20060102-150405")
	if out, err := s.exec(ctx, t, nil, "copyto", snapshot, s.remote(t, "db/vidra-"+stamp+".db")); err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	if out, err := s.exec(ctx, t, nil, "copyto", snapshot, s.remote(t, "db/vidra-latest.db")); err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	out, err := s.exec(ctx, t, nil, "lsf", "--files-only", "--include", "vidra-2*.db", s.remote(t, "db"))
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	var old []string
	for line := range strings.Lines(out) {
		if name := strings.TrimSpace(line); name != "" {
			old = append(old, name)
		}
	}
	slices.Sort(old)
	if len(old) <= backupSnapshots {
		return nil
	}
	if out, err := s.exec(ctx, t, fileList(old[:len(old)-backupSnapshots]), "delete", "--files-from-raw", "-", s.remote(t, "db")); err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

func (s *BackupService) snapshotDB(ctx context.Context) (string, error) {
	f, err := os.CreateTemp("", "vidra-*.db")
	if err != nil {
		return "", err
	}
	path := f.Name()
	_ = f.Close()
	_ = os.Remove(path)
	if _, err := s.db.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return "", err
	}
	return path, nil
}

func (s *BackupService) exec(ctx context.Context, t database.BackupTarget, stdin []byte, args ...string) (string, error) {
	spec := backupProviders[t.Provider]
	cmd := exec.CommandContext(ctx, s.rclone, append([]string{"--log-level", "ERROR"}, args...)...)
	env := append(os.Environ(), "RCLONE_CONFIG="+os.DevNull, "RCLONE_CONFIG_VIDRA_TYPE="+spec.Type)
	for k, v := range spec.Fixed {
		env = append(env, "RCLONE_CONFIG_VIDRA_"+strings.ToUpper(k)+"="+v)
	}
	for k, v := range decodeConfig(t.Config) {
		env = append(env, "RCLONE_CONFIG_VIDRA_"+strings.ToUpper(k)+"="+v)
	}
	cmd.Env = env
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if i := strings.LastIndex(msg, "ERROR : "); i >= 0 {
			msg = strings.TrimSpace(msg[i+len("ERROR : "):])
		}
		if msg != "" {
			return string(out), errors.New(msg)
		}
	}
	return string(out), err
}

func (s *BackupService) remote(t database.BackupTarget, sub string) string {
	if t.Provider == BackupProviderLocal {
		return "vidra:" + filepath.Join(t.Path, filepath.FromSlash(sub))
	}
	return "vidra:" + path.Join(t.Path, sub)
}

func (s *BackupService) setStatus(ctx context.Context, id, status, message string) {
	t, err := s.queries.SetBackupTargetStatus(ctx, database.SetBackupTargetStatusParams{ID: id, LastStatus: status, LastError: message})
	if err != nil {
		slog.Warn("failed to update backup status", "error", err)
		return
	}
	s.ws.Broadcast(WsEventBackupStatus, toBackupDTO(t))
}

func (s *BackupService) recordError(command string, err error, output string) {
	slog.Error("backup failed", "step", command, "error", err)
	if len(output) > maxOutput {
		output = output[len(output)-maxOutput:]
	}
	if e := s.queries.CreateError(context.Background(), database.CreateErrorParams{
		ID: NewID(), ErrorMessage: err.Error(), Command: command, Output: output,
	}); e != nil {
		slog.Warn("failed to record error", "error", e)
	}
}

func (s *BackupService) mustList(ctx context.Context) []database.BackupTarget {
	targets, err := s.queries.ListBackupTargets(ctx)
	if err != nil {
		slog.Warn("failed to list backup targets", "error", err)
	}
	return targets
}

func toBackupDTO(t database.BackupTarget) BackupTargetDTO {
	cfg := decodeConfig(t.Config)
	for _, f := range backupProviders[t.Provider].Fields {
		if f.Secret && cfg[f.Key] != "" {
			cfg[f.Key] = backupSecretMask
		}
	}
	return BackupTargetDTO{
		ID: t.ID, Name: t.Name, Provider: t.Provider, Config: cfg, Path: t.Path,
		IncludeVideos: t.IncludeVideos, IncludeDatabase: t.IncludeDatabase, Enabled: t.Enabled, IntervalHours: int(t.IntervalHours),
		LastRunAt: t.LastRunAt, LastFullAt: t.LastFullAt, LastStatus: t.LastStatus, LastError: t.LastError, CreatedAt: t.CreatedAt,
	}
}

func decodeConfig(raw string) map[string]string {
	cfg := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &cfg)
	return cfg
}

func fileList(names []string) []byte {
	return []byte(strings.Join(names, "\n") + "\n")
}

func cleanRemotePath(p string) string {
	p = strings.TrimSpace(p)
	if filepath.IsAbs(p) || filepath.VolumeName(p) != "" {
		return filepath.Clean(p)
	}
	return strings.Trim(path.Clean("/"+strings.ReplaceAll(p, "\\", "/")), "/")
}
