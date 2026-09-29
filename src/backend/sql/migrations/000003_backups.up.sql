CREATE TABLE backup_targets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    provider TEXT NOT NULL,
    config TEXT NOT NULL DEFAULT '{}',
    path TEXT NOT NULL DEFAULT '',
    include_videos BOOLEAN NOT NULL DEFAULT TRUE,
    include_database BOOLEAN NOT NULL DEFAULT TRUE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at TEXT,
    last_status TEXT NOT NULL DEFAULT 'never',
    last_error TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
