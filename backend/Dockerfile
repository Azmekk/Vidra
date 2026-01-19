# Build stage
FROM golang:1.25.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Install runtime dependencies: ffmpeg, python3 (for yt-dlp), and curl to download yt-dlp
RUN apk add --no-cache ffmpeg python3 curl

# Install yt-dlp
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/main .

# Create the downloads directory
RUN mkdir -p downloads

# Set environment variables (defaults)
ENV PORT=8080
ENV DATABASE_URL=postgres://postgres:postgres@db:5432/vidra?sslmode=disable

EXPOSE 8080

CMD ["./main"]
