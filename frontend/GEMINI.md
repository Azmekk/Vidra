# Vidra Frontend - Project Context

This is the frontend for **Vidra**, a video downloader application. It is built using **Svelte 5** and **SvelteKit**, styled with **Tailwind CSS 4**, and interfaces with a backend API (likely Go-based) for video processing and downloading.

## 🛠 Tech Stack

- **Framework:** [Svelte 5](https://svelte.dev/) (using Runes: `$state`, `$props`, `$derived`, etc.)
- **Meta-Framework:** [SvelteKit](https://kit.svelte.dev/)
- **Styling:** [Tailwind CSS 4](https://tailwindcss.com/blog/tailwindcss-v4-alpha) (using `@tailwindcss/vite`)
- **API Client:** [Axios](https://axios-http.com/) with TypeScript client generated via [OpenAPI Generator](https://openapi-generator.tech/)
- **Runtime:** [Bun](https://bun.sh/) (preferred for dependency management and running scripts)

## 📁 Project Structure

- `src/api/`: Contains the generated API client. **Do not edit manually.** Use `scripts/openapi_generate.sh` to update.
- `src/lib/`:
  - `api-client.ts`: Singleton instances of the generated API classes (`videosApi`, `errorsApi`, `ytdlpApi`).
  - `components/ui/`: UI components (Shadcn-style structure).
- `src/routes/`:
  - `/`: Main dashboard listing videos and their download progress.
  - `/download`: Interface for submitting new video download requests.
  - `/errors`: View for system or download errors.
- `scripts/`:
  - `openapi_generate.sh`: Bash script to regenerate the API client from the backend's Swagger definition.

## 🚀 Building and Running

### Development
```bash
bun install
bun run dev
```

### Production Build
```bash
bun run build
bun run preview
```

### API Client Generation
If the backend API changes, regenerate the client:
```bash
# Ensure backend is at ../backend/gen/docs/swagger/swagger.json
./scripts/openapi_generate.sh
```

## 📝 Development Conventions

- **Tooling:** Always use `bun` for package management and running scripts. Avoid `npm` or `yarn`.
- **Backend Types:** When working with backend-provided UUIDs (from `pgx/v5`), prefer using the `.String()` method for `pgtype.UUID` if a string is needed instead of manual scanning.
- **Svelte 5 Runes:** Use `$state`, `$derived`, and `$props` instead of Svelte 4's `let`, `$:`, and `export let`.
- **API Communication:** Use the exported API instances from `$lib/api-client`.
- **Styling:** Use Tailwind CSS 4 utility classes.
- **Path Aliases:**
  - `@/*` -> `src/*`
  - `$lib/*` -> `src/lib/*`
  - `$api/*` -> `src/api/*` (Note: Ensure this alias is correctly resolved in your environment).
- **Icons:** Use `@lucide/svelte` for iconography.

## 🔗 Backend Integration
The frontend proxies requests to the backend:
- `/api` -> `http://localhost:8080` (or `VITE_BACKEND_URL`)
- `/downloads` -> `http://localhost:8080/downloads` (Serving downloaded video files and thumbnails)
- `/swagger` -> `http://localhost:8080/swagger`
