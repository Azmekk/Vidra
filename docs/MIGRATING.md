# Migrating from Vidra 1.x to 2

Vidra 2 replaces the Postgres + backend + frontend + nginx stack with one container and a SQLite database. The image ships a `pg2sqlite` tool that copies your videos, errors and settings over.

- **Nothing old is touched.** Postgres keeps its data, and your videos are hardlinked into the new folder instead of copied, so they take no extra space. You can roll back at any time.
- **Downtime** is about a minute: from stopping the old app until the new one starts.
- **Accounts:** 1.x had none. You create one with a setup code after the switch.

## Before you start

- **Shell:** Linux with Docker Compose v2 (`docker compose`) and `sudo`.
- **Old stack:** the compose project from 1.x, with services named `db`, `backend`, `frontend` and `proxy` as in the old example. If you renamed them, adjust the names below.
- **Terminal:** run everything in one terminal session. The steps set shell variables that later steps use.

## 1. Collect what the old stack uses

Run this from the folder with your old `docker-compose.yml`, while the old stack is **still running**:

```bash
OLD="$PWD"
NEW="$(dirname "$OLD")/vidra-2"
BACKEND="$(docker compose ps -aq backend)"
DOWNLOADS="$(docker inspect "$BACKEND" --format '{{range .Mounts}}{{if eq .Destination "/app/downloads"}}{{.Source}}{{end}}{{end}}')"
NETWORK="$(docker inspect "$BACKEND" --format '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}{{end}}')"
PORT="$(docker compose port proxy 80 | sed 's/.*://')"
printf 'new folder: %s\ndownloads:  %s\nnetwork:    %s\nport:       %s\n' "$NEW" "$DOWNLOADS" "$NETWORK" "$PORT"
```

All four values must be filled in before you continue. `NEW` is where Vidra 2 will live. Change it if you want it somewhere else.

## 2. Pull the new image

```bash
docker pull ghcr.io/azmekk/vidra:2
```

## 3. Stop the old app and keep Postgres running

Downtime starts here.

```bash
docker compose stop $(docker compose config --services | grep -vx db)
```

## 4. Link the videos into the new folder

```bash
mkdir -p "$NEW/data"
sudo cp -al "$DOWNLOADS" "$NEW/downloads"
```

`cp -al` creates hardlinks: the same files appear in both folders without using extra space. `sudo` is needed because the old container wrote the files as root.

If it fails with `Invalid cross-device link`, the new folder is on a different disk. Either put `NEW` on the same disk as the downloads, or copy instead with `sudo cp -a "$DOWNLOADS" "$NEW/downloads"`, which needs free space for a full copy.

## 5. Migrate the database

```bash
docker run --rm --network "$NETWORK" \
  -v "$NEW/data:/app/data" -v "$NEW/downloads:/app/downloads" \
  -e DATABASE_URL="$(docker inspect "$BACKEND" --format '{{range .Config.Env}}{{println .}}{{end}}' | sed -n 's/^DATABASE_URL=//p')" \
  ghcr.io/azmekk/vidra:2 ./pg2sqlite -out /app/data/vidra.db -downloads /app/downloads
```

This reads the Postgres URL from the old backend, so you never type the password. It should end with `✅ Migrated N videos and M errors`.

- **Migrated as-is:** every video becomes a video with one original version, and files keep their names.
- **Migrated with changes:** settings (proxy, theme, default encoding) are converted to the new format.
- **Warnings:** `could not probe` only means that file's media info is missing. The video still works.

## 6. Create the new compose file

```bash
cat > "$NEW/docker-compose.yml" <<EOF
services:
  vidra:
    image: ghcr.io/azmekk/vidra:2
    restart: unless-stopped
    ports:
      - "$PORT:8080"
    volumes:
      - ./data:/app/data
      - ./downloads:/app/downloads
    # Only for plain http:// access. Remove it when Vidra sits behind HTTPS.
    # environment:
    #   VIDRA_INSECURE_COOKIES: "true"
EOF
```

Vidra 2 uses `Secure` session cookies, so logging in only works over HTTPS. If you open Vidra as `http://<server>:<port>`, uncomment the two `environment` lines.

## 7. Switch over

```bash
docker compose down
cd "$NEW" && docker compose up -d
```

`docker compose down` removes the old containers only. The Postgres data and your files stay where they are.

## 8. Create your account

```bash
docker compose logs vidra | grep "setup code"
```

Open Vidra on the same address as before, enter the code and choose a username and password. Your library should show all your old videos.

## Rolling back

```bash
cd "$NEW" && docker compose down
cd "$OLD" && docker compose up -d
```

Everything in the old folder is exactly as you left it.

## Cleaning up

Once you're happy with Vidra 2:

- **Why clean up:** deleting a video in Vidra 2 only frees disk space after the old hardlink is gone too.
- **Old stack:** delete the old folder, including its Postgres data and the old compose file, which holds the database password. If your old compose used named volumes instead of folders, also run `docker compose down -v` in the old folder *before* deleting it.
- **Old images:** `docker image rm ghcr.io/azmekk/vidra-backend:master ghcr.io/azmekk/vidra-frontend:master postgres:18-alpine nginx:1.29.4-alpine` (skip any image another stack still uses).
- **Renaming:** to rename `vidra-2` back to `vidra`, run `docker compose down` first and `docker compose up -d` again in the renamed folder, since Compose derives the project name from the folder.

## Staying on a major version

`ghcr.io/azmekk/vidra:2` follows every 2.x release and never jumps to a future 3.0 with breaking changes. `:latest` follows `master` and may. Pin `:2` (or an exact version such as `:2.0.0`) unless you want to be on the edge.
