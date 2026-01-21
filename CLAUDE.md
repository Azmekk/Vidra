# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Vidra is a full-stack video downloader that uses `yt-dlp` for fetching video metadata/downloads and `ffmpeg` for H.264 encoding. It features a Go backend REST API and a Svelte 5 frontend.

## Common Commands

### Backend (run from `backend/` directory)

```bash
go run main.go                          # Run the server
go mod tidy                             # Tidy dependencies
go vet ./...                            # Lint/check code
sqlc generate                           # Generate database code from SQL
./scripts/swag_init.sh                  # Update Swagger docs
./scripts/add_migration.sh <name>       # Create new database migration
./scripts/sqlc_generate.sh              # Generate sqlc code
```

### Frontend (run from `frontend/` directory)

```bash
bun install                             # Install dependencies
bun run dev                             # Development server (proxies to backend :8080)
bun run build                           # Production build
bun run preview                         # Preview production build
bun run check                           # Type checking
./scripts/openapi_generate.sh           # Regenerate API client from Swagger
```

### Docker

```bash
docker compose up -d                    # Start full stack (from root with docker-compose.yml)
```

## Architecture

### Monorepo Structure

- `backend/` - Go REST API with Chi router, PostgreSQL via sqlc
- `frontend/` - SvelteKit web app with generated TypeScript API client
- `nginx.conf` - Reverse proxy serving both services

### Backend Flow

```
HTTP Request → Chi Router (routers/) → Handler (handlers/) → Service (services/) → Database (gen/database/)
```

Key services:

- `services/bootstrap.go` - DB initialization, migrations
- `services/downloader.go` - yt-dlp/ffmpeg orchestration with background goroutines
- `services/websocket.go` - Real-time progress broadcasting

### Frontend Flow

- Server-side data loading via `+page.ts` load functions
- Client-side mutations via generated API client (`src/api/`)
- Real-time updates via WebSocket
- Singleton API instances in `src/lib/api-client.ts`

### Database Layer

- Migrations in `backend/sql/migrations/`
- Query definitions in `backend/sql/queries/`
- Generated code in `backend/gen/database/`

### API Client Generation

The frontend API client is auto-generated from the backend's Swagger spec. After backend API changes:

1. Run `backend/scripts/swag_init.sh`
2. Run `frontend/scripts/openapi_generate.sh`

## Key Conventions

### Backend

- Use DTOs for API responses (avoid exposing `pgtype.*` types directly)
- Swagger operation IDs in camelCase
- All database interactions through sqlc-generated code
- Background tasks use `sync.Map` for thread-safe progress tracking
- Video processing: download with GUID temp filename → encode to H.264 → save as .mp4

### Frontend

- Use Svelte 5 Runes: `$state`, `$derived`, `$props` (not Svelte 4 syntax)
- Use `bun` for package management (not npm/yarn)
- API calls through singleton instances from `$lib/api-client`
- Check `browser` from `$app/environment` for client-side code
- UI components in `src/lib/components/ui/` (shadcn-style)
- Icons from `@lucide/svelte`
- Path aliases: `$lib/*` → `src/lib/*`, `$api/*` → `src/api/*`
- Biome for formatting/linting

```bash
# Format all files
bunx biome format --write

# Format specific files
bunx biome format --write <files>

# Lint and apply safe fixes to all files
bunx biome lint --write

# Lint files and apply safe fixes to specific files
bunx biome lint --write <files>

# Format, lint, and organize imports of all files
bunx biome check --write

# Format, lint, and organize imports of specific files
bunx biome check --write <files>
```

### Environment Variables

- Backend: `backend/.env` (see `.env.example`) - DATABASE_URL, PORT
- Frontend: `frontend/.env` - VITE_BACKEND_URL

## API Endpoints

```
POST   /api/videos              - Create video download
GET    /api/videos              - List videos (paginated, searchable)
POST   /api/videos/metadata     - Get metadata + format options for URL
GET    /api/videos/{id}         - Get single video
PUT    /api/videos/{id}         - Update video name
DELETE /api/videos/{id}         - Delete video
GET    /api/videos/{id}/progress - Get download progress
GET    /api/ws                  - WebSocket for real-time updates
GET    /swagger/*               - Swagger UI & JSON
```

## Database Schema

Main tables:

- `videos` - id (UUID), name, file_name, thumbnail_file_name, original_url, download_status, timestamps
- `errors` - error logging with video_id foreign key
