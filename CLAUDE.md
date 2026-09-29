# CLAUDE.md

Guidance for Claude Code (claude.ai/code) in this repository.

## Project overview

Vidra is a self-hosted, mobile-first video downloader. A single Go binary serves a REST API, a WebSocket and an embedded React SPA, stores data in SQLite and shells out to `yt-dlp`, `ffmpeg`/`ffprobe` and `rclone`. It ships as one Docker image and always runs behind HTTPS in production.

## Layout

- `src/backend`: Go module `github.com/Azmekk/Vidra/backend` (chi, sqlc on SQLite via `modernc.org/sqlite`, golang-migrate, swag).
- `src/frontend`: React 19 SPA (Vite 8, TypeScript 7, TanStack Router + Query, zod 4, Tailwind 4, shadcn/ui, Biome). Builds into `src/backend/web/build/app`, which the binary embeds.
- `Dockerfile`: Bun build, Go build with the embedded SPA, Alpine runtime with ffmpeg, python3, yt-dlp, deno and rclone.
- `docker-compose.example.yml`: single-service example. A local `docker-compose.yml` is gitignored.

## Commands

### Backend (`src/backend`)

```bash
sqlc generate                  # gen/database from sql/queries (gen/ is gitignored)
./scripts/swag_init.sh         # gen/docs/swagger
./scripts/add_migration.sh <name>
./scripts/check.sh             # gofmt, go vet and staticcheck for windows and linux
VIDRA_INSECURE_COOKIES=true go run .
go run ./cmd/pg2sqlite --pg <postgres url> --out data/vidra.db
```

### Frontend (`src/frontend`)

```bash
bun install
bun run dev                    # proxies /api and the WebSocket to :8080
bun run check                  # tsc + Biome, must report zero diagnostics
bun run build                  # outputs to ../backend/web/build/app
bun run api                    # regenerate src/api/gen with orval from the swagger spec
bun run icons                  # regenerate PWA icons from public/favicon.svg
bunx biome check --write       # format, lint and sort imports
```

After changing API handlers or DTOs: `./scripts/swag_init.sh` in the backend, then `bun run api` in the frontend. The generated client in `src/api/gen` is committed and excluded from Biome.

## Backend architecture

`main.go` wires everything: config → SQLite (`services.OpenDatabase` applies embedded migrations from `sql/migrations`) → services → chi routes. All `/api` routes except `/api/auth/{status,login,setup,logout}` go through `middleware.RequireAuth`.

Key services:

- `services/videos.go`: `VideoStore`, the only code touching video/file queries. Keeps an in-memory cache of the newest N videos and broadcasts `video_*` events.
- `services/downloader.go` and `jobs.go`: download and encode pools, cancellable jobs, per-file progress (`file_progress`), `OnFileCompleted` and `OnFilesDeleted` hooks.
- `services/encoding/`: ffmpeg capability detection, curated encoder catalog, `Profile.BuildArgs` and the goal-based recommender.
- `services/auth.go` and `middleware/auth.go`: argon2id, sliding session cookie `vidra_session`, bearer API tokens, same-origin check for cookie-authenticated writes.
- `middleware/client_ip.go`: client IP resolution (X-Forwarded-For trusted only from private peers) and per-IP rate limiting.
- `services/backup.go`: rclone targets configured in Settings. Remotes are passed to rclone as `RCLONE_CONFIG_VIDRA_*` env vars per process. Secrets are masked in responses.
- `services/websocket.go`: per-client buffered queues.

Data model: `videos` (logical item, renamable, `primary_file_id`) → `video_files` (versions: `original` or `encode`, status, media info, `ios_compatible`). Also `settings` (single row), `errors`, `users`, `sessions`, `api_tokens`, `backup_targets`.

## Conventions

### Backend

- Every query goes through sqlc. The SQLite engine ignores args in `ORDER BY` (see the `params` CTE in `ListVideos`), and text args are wrapped as `CAST(sqlc.arg(x) AS TEXT)`.
- IDs are UUIDv7 strings (`services.NewID`), timestamps are millisecond RFC 3339 text.
- Responses use DTOs, never raw sqlc rows. Swagger operation IDs are camelCase. Run swag with `--requiredByDefault`, so optional fields must be `omitempty`.
- Env vars are limited to `PORT`, `DB_PATH`, `DOWNLOADS_DIR` and `VIDRA_INSECURE_COOKIES`. Everything else belongs in Settings.
- `./scripts/check.sh` must pass, and deprecations (staticcheck SA1019) count as failures.

### Frontend

- Use Bun, never npm or yarn. Install packages with `@latest`.
- API calls go through the generated orval hooks and `src/api/fetcher.ts`. A 401 clears the cache and redirects to login.
- Live updates arrive in `src/lib/ws.ts`, which parses with zod and patches the Query cache (`src/lib/videos.ts`).
- Code-based routes live in `src/router.tsx`, with search params validated by zod.
- Never let a page flicker or show a blank state. Every loading state gets a shape-matched skeleton (`src/components/skeletons.tsx`), refetches keep previous data, and the theme is set before first paint.
- Inputs are at least 16px (no iOS zoom), layouts respect safe areas, and the design uses large rounded cards (`rounded-[2.5rem]`), bold type and `max-w-2xl`.
- Use `@/` imports, shadcn components in `src/components/ui`, icons from `lucide-react` and toasts from `sonner`.
- Keep comments to a minimum and don't add tests unless asked.
