CREATE TABLE videos (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    source_title TEXT,
    original_url TEXT NOT NULL,
    thumbnail_file_name TEXT,
    duration REAL,
    uploader TEXT,
    primary_file_id TEXT REFERENCES video_files (id) ON DELETE SET NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_videos_created_at ON videos (created_at);

CREATE TABLE video_files (
    id TEXT PRIMARY KEY,
    video_id TEXT NOT NULL REFERENCES videos (id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    source_file_id TEXT REFERENCES video_files (id) ON DELETE SET NULL,
    label TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    file_name TEXT,
    container TEXT,
    video_codec TEXT,
    audio_codec TEXT,
    width INTEGER,
    height INTEGER,
    fps REAL,
    bitrate INTEGER,
    file_size INTEGER,
    ios_compatible BOOLEAN NOT NULL DEFAULT FALSE,
    encoding_profile TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_video_files_video_id ON video_files (video_id, created_at);

CREATE TABLE errors (
    id TEXT PRIMARY KEY,
    video_id TEXT REFERENCES videos (id) ON DELETE CASCADE,
    file_id TEXT REFERENCES video_files (id) ON DELETE SET NULL,
    error_message TEXT NOT NULL,
    command TEXT NOT NULL,
    output TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_errors_created_at ON errors (created_at);

CREATE TABLE settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    proxy_url TEXT NOT NULL DEFAULT '',
    theme TEXT NOT NULL DEFAULT 'system',
    prefer_compatible_formats BOOLEAN NOT NULL DEFAULT TRUE,
    default_encoding TEXT NOT NULL DEFAULT '{"goal":"compatible"}',
    keep_original BOOLEAN NOT NULL DEFAULT TRUE,
    cache_size INTEGER NOT NULL DEFAULT 100,
    max_concurrent_downloads INTEGER NOT NULL DEFAULT 3,
    max_concurrent_encodes INTEGER NOT NULL DEFAULT 1,
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO settings (id) VALUES (1);
