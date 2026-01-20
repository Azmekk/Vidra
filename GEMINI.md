# Vidra - Full-Stack Video Downloader

Vidra is a comprehensive video downloading platform that leverages `yt-dlp` for fetching video metadata and downloads, and `ffmpeg` for high-quality H.264 encoding. It features a RESTful Go backend and a modern Svelte 5 frontend.

## 🏗 Project Structure

The project is organized as a monorepo:

- **`backend/`**: Go-based REST API and background worker service.
- **`frontend/`**: SvelteKit-based web application.

---

## 🚀 CI/CD & Deployment

Vidra features a fully automated CI/CD pipeline using GitHub Actions, making deployment and updates seamless.

### Docker Images (GHCR)
Production-ready images are automatically built and published to the GitHub Container Registry on every push to `main`:
- **Backend**: `ghcr.io/azmekk/vidra-backend:main`
- **Frontend**: `ghcr.io/azmekk/vidra-frontend:main`

### Deployment Options
1.  **Zero-Clone (Recommended)**: Deploy by only downloading `docker-compose.yml.example` and `nginx.conf`. This is the fastest way to run Vidra without needing the source code locally.
2.  **Local Build**: Clone the repository and build the images manually using the provided Dockerfiles (primarily for development).

---

## 🚀 Backend (`backend/`)

The backend manages video format listing, background download orchestration, and progress tracking.

### Key Technologies
- **Language:** Go (1.25.5)
- **Web Framework:** Chi (v5)
- **Database:** PostgreSQL with `pgx/v5` and `sqlc` for type-safe queries.
- **Migrations:** `golang-migrate`
- **Video Tools:** `yt-dlp` (Downloading), `ffmpeg` (Encoding)
- **Documentation:** Swagger (swaggo)

### Building and Running
1. **Setup Environment:** Copy `backend/.env.example` to `backend/.env` and configure your database.
2. **Run Server:** `go run main.go` (from `backend/` directory)
3. **Generate SQL Code:** `sqlc generate`
4. **Update Swagger:** `./scripts/swag_init.sh`

---

## 🎨 Frontend (`frontend/`)

The frontend provides a reactive user interface for submitting downloads and monitoring progress in real-time.

### Key Technologies
- **Framework:** Svelte 5 (using Runes: `$state`, `$derived`, etc.)
- **Meta-Framework:** SvelteKit
- **Styling:** Tailwind CSS 4
- **API Client:** Generated TypeScript client via OpenAPI/Axios.
- **Runtime:** Bun (preferred)

### Building and Running
1. **Install Dependencies:** `bun install`
2. **Run Development Server:** `bun run dev`
3. **Build for Production:** `bun run build`
4. **Regenerate API Client:** `./scripts/openapi_generate.sh`

---

## 🛠 Shared Workflows

### API Changes
1. Modify backend handlers/models.
2. Run `backend/scripts/swag_init.sh` to update `swagger.json`.
3. Run `frontend/scripts/openapi_generate.sh` to update the TypeScript API client.

### Database Changes
1. Add a new migration in `backend/sql/migrations/` (use `backend/scripts/add_migration.sh`).
2. Define queries in `backend/sql/queries/`.
3. Run `backend/scripts/sqlc_generate.sh`.

## 📝 Development Conventions

- **State Management:** Use Svelte 5 Runes ($state, $props, $derived) in the frontend.
- **API Communication:** Use the singleton instances in `frontend/src/lib/api-client.ts`.
- **Database:** All database interactions should go through `sqlc` generated code.
- **Progress Tracking:** Background tasks use a thread-safe `sync.Map` in the backend, exposed via a dedicated progress endpoint.
