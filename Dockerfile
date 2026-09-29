# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM oven/bun:1.4-slim AS web
WORKDIR /src/frontend
COPY src/frontend/package.json src/frontend/bun.lock ./
RUN bun install --frozen-lockfile
COPY src/frontend ./
RUN bun run build

FROM --platform=$BUILDPLATFORM sqlc/sqlc:1.30.0 AS sqlc

FROM --platform=$BUILDPLATFORM golang:1.27-alpine3.24 AS api
ARG TARGETOS TARGETARCH
WORKDIR /src/backend
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4
COPY --from=sqlc /workspace/sqlc /usr/local/bin/sqlc
COPY src/backend/go.mod src/backend/go.sum ./
RUN go mod download
COPY src/backend ./
COPY --from=web /src/backend/web/build/app ./web/build/app
RUN sqlc generate \
    && swag init --output ./gen/docs/swagger --parseInternal --requiredByDefault --quiet \
    && export CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    && go build -trimpath -ldflags="-s -w" -o /out/vidra . \
    && go build -trimpath -ldflags="-s -w" -o /out/pg2sqlite ./cmd/pg2sqlite

FROM alpine:3.24
RUN apk add --no-cache ca-certificates tzdata ffmpeg python3 deno rclone
ADD --chmod=755 https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp /usr/local/bin/yt-dlp
WORKDIR /app
COPY --from=api /out/vidra /out/pg2sqlite ./
ENV PORT=8080 \
    DB_PATH=/app/data/vidra.db \
    DOWNLOADS_DIR=/app/downloads
VOLUME ["/app/data", "/app/downloads"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s \
    CMD wget -qO- "http://127.0.0.1:${PORT}/api/auth/status" >/dev/null || exit 1
CMD ["./vidra"]
