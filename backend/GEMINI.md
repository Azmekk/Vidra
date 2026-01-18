# Vidra Backend - Project Context

This project is a Go-based backend for a video downloading platform, utilizing `yt-dlp` for fetching video metadata and downloads, and `ffmpeg` for encoding processed videos to H.264.

## Project Overview
- **Purpose:** Provide a RESTful API for listing available video formats, initiating background downloads, and tracking progress.
- **Main Technologies:**
    - **Language:** Go (1.25.5)
    - **Web Framework:** [Chi](https://github.com/go-chi/chi) (v5)
    - **Database:** PostgreSQL
    - **Database Driver:** [pgx/v5](https://github.com/jackc/pgx)
    - **SQL Generation:** [sqlc](https://sqlc.dev/)
    - **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)
    - **API Documentation:** [Swagger/swaggo](https://github.com/swaggo/swag)
    - **CLI Tools:** `yt-dlp` (Video download), `ffmpeg` (H.264 encoding)

## Architecture
- `main.go`: Application entry point. Initializes dependencies via `services.Bootstrap` and starts the HTTP server on the configured `PORT`.
- `handlers/`: HTTP request handlers (e.g., `VideoHandler`).
- `routers/`: Route definitions using Chi.
- `services/`:
    - `bootstrap.go`: Handles environment loading, database creation, and running migrations.
    - `downloader.go`: Manages `yt-dlp` interactions and background download/encoding tasks with thread-safe progress tracking.
- `sql/`:
    - `migrations/`: Versioned SQL migration files (`.up.sql` and `.down.sql`).
    - `queries/`: SQL query definitions used by `sqlc`.
- `gen/`:
    - `database/`: Generated Go code from `sqlc`.
    - `docs/swagger/`: Generated Swagger documentation.

## Building and Running
- **Run the server:** `go run main.go`
- **Generate database code:** `sqlc generate` (or run `./scripts/sqlc_generate.sh`)
- **Initialize/Update Swagger docs:** `swag init -g main.go --output ./gen/docs/swagger` (or run `./scripts/swag_init.sh`)
- **Create a new migration:** `./scripts/add_migration.sh <migration_name>`

## Development Conventions
- **Environment Variables:** Configuration is managed via a `.env` file (see `.env.example`).
- **Database Initialization:** The application automatically ensures the database exists and runs migrations on startup.
- **API Design:** 
    - Uses **DTOs** (Data Transfer Objects) for API responses to avoid exposing database-specific types (e.g., `pgtype.Text`).
    - Swagger operation IDs are in `camelCase`.
- **Background Tasks:** 
    - Video downloads are initiated in background goroutines.
    - Real-time progress (percentage, speed, ETA) is stored in a `sync.Map` and exposed via `/api/videos/{id}/progress`.
- **File Processing:**
    1. Download video using its GUID as a temporary filename (to avoid conflicts).
    2. Encode the video to H.264 using `ffmpeg` and save it as an `.mp4` with the final name in the `downloads/` directory.
    3. Clean up the temporary file.
