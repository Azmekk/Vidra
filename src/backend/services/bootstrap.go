package services

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	schema "github.com/Azmekk/Vidra/backend/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

type Config struct {
	Port            string
	DBPath          string
	DownloadsDir    string
	InsecureCookies bool
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	return Config{
		Port:            envOr("PORT", "8080"),
		DBPath:          envOr("DB_PATH", filepath.Join("data", "vidra.db")),
		DownloadsDir:    envOr("DOWNLOADS_DIR", "downloads"),
		InsecureCookies: os.Getenv("VIDRA_INSECURE_COOKIES") == "true",
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// OpenDatabase opens the SQLite database, tunes it for a single-node web app
// and applies all pending migrations.
func OpenDatabase(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	pragmas := url.Values{}
	for _, p := range []string{
		"journal_mode(WAL)",
		"synchronous(NORMAL)",
		"busy_timeout(5000)",
		"foreign_keys(ON)",
		"temp_store(MEMORY)",
		"mmap_size(268435456)",
		"cache_size(-16000)",
	} {
		pragmas.Add("_pragma", p)
	}
	pragmas.Set("_txlock", "immediate")

	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?"+pragmas.Encode())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := migrateUp(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateUp(db *sql.DB) error {
	source, err := iofs.New(schema.Migrations, "migrations")
	if err != nil {
		return err
	}
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Println("Migrations applied successfully!")
	return nil
}
