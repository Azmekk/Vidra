// Command pg2sqlite copies a Vidra 1.x Postgres database into the SQLite
// format used by Vidra 2. Existing files on disk are kept as they are.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Azmekk/Vidra/backend/gen/database"
	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/services/encoding"
	"github.com/jackc/pgx/v5"
)

type pgVideo struct {
	ID, Name, OriginalURL, Status string
	FileName, Thumbnail           *string
	FileSize                      *int64
	CreatedAt, UpdatedAt          time.Time
}

type pgError struct {
	ID, Message, Command, Output string
	VideoID                      *string
	CreatedAt                    time.Time
}

type pgSettings struct {
	ProxyURL, VideoCodec, AudioCodec, Theme string
	ReEncode                                bool
	CRF                                     int
}

func main() {
	pgURL := flag.String("pg", os.Getenv("DATABASE_URL"), "Postgres connection URL (defaults to $DATABASE_URL)")
	out := flag.String("out", filepath.Join("data", "vidra.db"), "SQLite database to create")
	downloads := flag.String("downloads", "downloads", "downloads directory, used to read media info with ffprobe")
	force := flag.Bool("force", false, "replace existing videos and errors in the target database")
	flag.Parse()

	if *pgURL == "" {
		log.Fatal("missing -pg connection URL")
	}
	if err := run(context.Background(), *pgURL, *out, *downloads, *force); err != nil {
		log.Fatalf("%v", err)
	}
}

func run(ctx context.Context, pgURL, out, downloads string, force bool) error {
	pg, err := pgx.Connect(ctx, pgURL)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pg.Close(ctx)

	videos, err := readVideos(ctx, pg)
	if err != nil {
		return err
	}
	errs, err := readErrors(ctx, pg)
	if err != nil {
		return err
	}
	settings, err := readSettings(ctx, pg)
	if err != nil {
		log.Printf("WARN: settings not migrated: %v", err)
	}

	db, err := services.OpenDatabase(out)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "PRAGMA defer_foreign_keys = ON"); err != nil {
		return err
	}
	q := database.New(tx)

	existing, err := q.CountAllVideos(ctx)
	if err != nil {
		return err
	}
	if existing > 0 {
		if !force {
			return fmt.Errorf("%s already contains %d videos; use -force to replace them", out, existing)
		}
		if err := q.DeleteAllErrors(ctx); err != nil {
			return err
		}
		if err := q.DeleteAllVideos(ctx); err != nil {
			return err
		}
	}

	for _, v := range videos {
		if err := importVideo(ctx, q, v, downloads); err != nil {
			return fmt.Errorf("video %s: %w", v.ID, err)
		}
	}
	for _, e := range errs {
		if err := q.ImportError(ctx, database.ImportErrorParams{
			ID: e.ID, VideoID: e.VideoID, ErrorMessage: e.Message, Command: e.Command, Output: e.Output,
			CreatedAt: timestamp(e.CreatedAt),
		}); err != nil {
			return fmt.Errorf("error %s: %w", e.ID, err)
		}
	}
	if settings != nil {
		if _, err := services.NewSettingsService(q).Update(ctx, convertSettings(*settings)); err != nil {
			return fmt.Errorf("settings: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Migrated %d videos and %d errors into %s", len(videos), len(errs), out)
	return nil
}

func importVideo(ctx context.Context, q *database.Queries, v pgVideo, downloads string) error {
	fileID := services.NewID()
	status := services.FileError
	if v.Status == "completed" && v.FileName != nil {
		status = services.FileCompleted
	}

	file := database.ImportVideoFileParams{
		ID: fileID, VideoID: v.ID, Kind: services.KindOriginal, Label: "Original", Status: status,
		FileName: v.FileName, FileSize: v.FileSize,
		CreatedAt: timestamp(v.CreatedAt), UpdatedAt: timestamp(v.UpdatedAt),
	}
	if v.FileName != nil {
		if p, err := services.ProbeFile(ctx, filepath.Join(downloads, *v.FileName)); err == nil {
			file.Container, file.VideoCodec, file.AudioCodec = ptr(p.Container), ptr(p.VideoCodec), ptr(p.AudioCodec)
			file.Width, file.Height, file.Fps = ptr(int64(p.Width)), ptr(int64(p.Height)), ptr(p.Fps)
			file.Bitrate, file.FileSize, file.IosCompatible = ptr(p.Bitrate), ptr(p.Size), p.IOSCompatible()
			if p.Height > 0 {
				file.Label = fmt.Sprintf("Original · %dp", p.Height)
			}
		} else {
			log.Printf("WARN: could not probe %s: %v", *v.FileName, err)
		}
	}

	var primary *string
	if status == services.FileCompleted {
		primary = &fileID
	}
	if err := q.ImportVideo(ctx, database.ImportVideoParams{
		ID: v.ID, Name: v.Name, OriginalUrl: v.OriginalURL, ThumbnailFileName: v.Thumbnail,
		PrimaryFileID: primary, CreatedAt: timestamp(v.CreatedAt), UpdatedAt: timestamp(v.UpdatedAt),
	}); err != nil {
		return err
	}
	return q.ImportVideoFile(ctx, file)
}

func readVideos(ctx context.Context, pg *pgx.Conn) ([]pgVideo, error) {
	rows, err := pg.Query(ctx, `SELECT id::text, name, original_url, download_status, file_name,
		thumbnail_file_name, file_size, created_at, updated_at FROM videos ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("read videos: %w", err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (pgVideo, error) {
		var v pgVideo
		err := r.Scan(&v.ID, &v.Name, &v.OriginalURL, &v.Status, &v.FileName, &v.Thumbnail, &v.FileSize, &v.CreatedAt, &v.UpdatedAt)
		return v, err
	})
}

func readErrors(ctx context.Context, pg *pgx.Conn) ([]pgError, error) {
	rows, err := pg.Query(ctx, `SELECT id::text, video_id::text, error_message, command, output, created_at
		FROM errors ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("read errors: %w", err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (pgError, error) {
		var e pgError
		err := r.Scan(&e.ID, &e.VideoID, &e.Message, &e.Command, &e.Output, &e.CreatedAt)
		return e, err
	})
}

func readSettings(ctx context.Context, pg *pgx.Conn) (*pgSettings, error) {
	var s pgSettings
	err := pg.QueryRow(ctx, `SELECT proxy_url, default_re_encode, default_video_codec, default_audio_codec,
		default_crf, theme FROM settings WHERE id = 1`).
		Scan(&s.ProxyURL, &s.ReEncode, &s.VideoCodec, &s.AudioCodec, &s.CRF, &s.Theme)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func convertSettings(s pgSettings) services.Settings {
	req := encoding.Request{Goal: encoding.GoalOriginal}
	if s.ReEncode {
		container := "mp4"
		if s.VideoCodec == "libvpx-vp9" || s.VideoCodec == "vp9_qsv" {
			container = "webm"
		}
		crf := s.CRF
		req = encoding.Request{Goal: encoding.GoalCustom, Profile: &encoding.Profile{
			Mode: encoding.ModeEncode, VideoEncoder: s.VideoCodec, Quality: &crf,
			Container: container, AudioEncoder: s.AudioCodec,
		}}
	}
	return services.Settings{
		ProxyURL:                s.ProxyURL,
		Theme:                   s.Theme,
		PreferCompatibleFormats: true,
		DefaultEncoding:         req,
		KeepOriginal:            true,
		CacheSize:               100,
		MaxConcurrentDownloads:  3,
		MaxConcurrentEncodes:    1,
	}
}

func timestamp(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z") }

func ptr[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}
