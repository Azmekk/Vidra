# Vidra

Vidra is a self-hosted, mobile-first video downloader. Paste a link from YouTube, TikTok, Instagram or any other site [yt-dlp](https://github.com/yt-dlp/yt-dlp) supports, and Vidra downloads it, keeps every version you make of it, and puts it one tap away from your phone's photo library.

It ships as one small Go binary with the web app built in, SQLite for storage and a single Docker image.

## Features

- **Quick download.** Paste a link (or a whole "shared via…" message) and go. Vidra finds the link, names the video, and fills in the real title when it lands.
- **Versions.** Keep the original and re-encode it any time to H.264, HEVC, AV1 or VP9. Pick which version is the default.
- **Smart encoding.** Choose a goal (iPhone, Balanced, Smallest, Fastest) and get a recommended profile with reasons, or set every ffmpeg option yourself. Working hardware encoders are detected automatically.
- **Save to Photos on iPhone.** One tap opens the share sheet with the video ready to save. Versions that won't play on iOS get a one-tap "Make iPhone version".
- **Installable app.** A PWA with offline shell, share target on Android and an iOS Shortcut via API tokens.
- **Backups.** Copy videos and database snapshots to S3/R2, Google Drive, MEGA or a local folder with rclone, configured from Settings.
- **Live progress** over WebSocket, cancellable jobs, infinite-scroll library, search and an error log.

## Screenshots

<div align="center">
  <table border="0">
    <tr>
      <td><img src="assets/mobile-library.png" alt="Library" width="220"></td>
      <td><img src="assets/mobile-download.png" alt="Download" width="220"></td>
      <td><img src="assets/mobile-versions.png" alt="Versions" width="220"></td>
      <td><img src="assets/mobile-settings.png" alt="Settings" width="220"></td>
    </tr>
    <tr>
      <td align="center">Library</td>
      <td align="center">Download</td>
      <td align="center">Versions</td>
      <td align="center">Settings</td>
    </tr>
  </table>
</div>

## Quick start

```bash
mkdir vidra && cd vidra
curl -L https://raw.githubusercontent.com/Azmekk/Vidra/master/docker-compose.example.yml -o docker-compose.yml
docker compose up -d
docker compose logs vidra | grep "setup code"
```

The example pins `ghcr.io/azmekk/vidra:2`, which gets every 2.x update but never a breaking major release. `:latest` is always the newest release, including future major versions.

Open Vidra, enter the setup code from the log and create your account. The code is only printed while no account exists.

Vidra is meant to run behind HTTPS (Caddy, Traefik, Cloudflare Tunnel, …). Session cookies are `Secure`, and Save to Photos needs a secure context. To try it over plain HTTP on your LAN, set `VIDRA_INSECURE_COOKIES=true`.

### Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port |
| `DB_PATH` | `/app/data/vidra.db` | SQLite database |
| `DOWNLOADS_DIR` | `/app/downloads` | Videos and thumbnails |
| `VIDRA_INSECURE_COOKIES` | `false` | Allow sessions over plain HTTP |

Everything else, including proxy, concurrency, default encoding and backups, lives in **Settings**.

Vidra keeps its own yt-dlp in the data folder (`data/bin`) and updates it on startup and daily, so updates survive image upgrades. To stay on a specific version, pick it in **Settings → System**; choose **Always latest** to unpin.

When a reverse proxy runs on the same host or Docker network, Vidra trusts its `X-Forwarded-For` header for rate limiting. Requests from public addresses are keyed by their own IP.

## Phones

Open Vidra in Safari (iPhone) or Chrome (Android) and add it to the home screen: **Share → Add to Home Screen** or **⋮ → Install app**.

### Share to Vidra on Android

Once installed, Vidra appears in the share sheet of every app. Sharing a video from TikTok, Instagram, YouTube and others opens the download page with the link already filled in.

### Share to Vidra on iPhone

iOS doesn't let web apps join the share sheet, so a Shortcut sends the link to Vidra instead. You set it up once. After that, sharing is **Share → Vidra** and the download starts on the server without opening anything.

1. In Vidra, open **Settings → API tokens**, create a token (e.g. "iPhone Shortcut") and copy it. It is only shown once.
2. In the **Shortcuts** app, tap **+**. The editor opens with an empty "New Shortcut".
3. Tap **New Shortcut ⌄** at the top and choose **Rename**. Call it "Vidra".
4. Tap the **ⓘ** button at the bottom of the editor, turn on **Show in Share Sheet** and tap **Done**. A **Receive Images and 16 more from Share Sheet** block appears at the top.
5. Optionally, tap **Images and 16 more** and leave only **URLs** and **Text** selected, so Vidra only shows up when you share a link or text. The defaults also work.
6. In **Search Actions**, find **Get Contents of URL** and add it. Tap its **URL** placeholder and enter `https://<your-vidra>/api/videos/quick`, then expand the action (the arrow or **Show More**):
   - Method: **POST**
   - Headers: **Add new header**, key `Authorization`, value `Bearer <your token>`
   - Request Body: **JSON**, then **Add new field → Text**, key `url`, value **Shortcut Input** (pick it from the variables bar above the keyboard)
7. Optionally, add **Show Notification** with "Sent to Vidra" so you get a confirmation.

In TikTok, tap **Share → More (…) → Vidra**. Apps often share text like "Check out this video! https://vm.tiktok.com/…" rather than a bare link. That's fine, because Vidra finds the link in the text.

A few things to know:

- **Encoding:** quick downloads use your default encoding from Settings.
- **Names:** shared videos get a random name. Tap the pencil in the library to rename; the card suggests the video's original title.
- **Reaching Vidra:** the phone must be able to reach your Vidra address, e.g. on your home network or over a VPN such as Tailscale.
- **Security:** a token can do everything your account can. If it leaks, delete it in Settings and it stops working immediately.
- **Opening Vidra instead:** to pick the format first, have the Shortcut run **Open URLs** with `https://<your-vidra>/download?url=<Shortcut Input>` instead. Adding `&quick=1` starts a quick download.

## Backups

In **Settings → Backups**, add one or more targets:

- **S3 / R2** (AWS, Cloudflare R2, Wasabi, MinIO, …): endpoint, keys and `bucket/folder`.
- **Google Drive**: run `rclone authorize "drive"` on any computer and paste the token.
- **MEGA**: email and password.
- **Local folder**: any absolute path, e.g. a mounted NAS share.

Finished versions and thumbnails upload as they complete, deleted ones are removed from the remote, and the database is snapshotted shortly after changes and daily (`db/vidra-latest.db` plus the last 7). **Run** copies everything again and is safe to repeat.

## Migrating from Vidra 1.x (Postgres)

Vidra 2 is a breaking release: one container and SQLite instead of Postgres, a backend, a frontend and nginx. 1.x images (`vidra-backend`, `vidra-frontend`) are no longer updated, so existing installs keep running until you migrate.

Follow [docs/MIGRATING.md](docs/MIGRATING.md). It copies your videos, errors and settings with the bundled `pg2sqlite`, hardlinks your downloads instead of copying them and leaves the old stack intact for rollback.

## Development

Requirements: Go 1.27, Bun, sqlc, swag, ffmpeg and yt-dlp (rclone and deno optional).

```bash
cd src/backend
sqlc generate && ./scripts/swag_init.sh
VIDRA_INSECURE_COOKIES=true go run .

cd src/frontend
bun install
bun run dev
```

The Vite dev server proxies `/api` to `:8080`. See [CLAUDE.md](CLAUDE.md) for layout and conventions.

## License

AGPL-3.0. See [LICENSE](LICENSE).
