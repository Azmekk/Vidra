-- name: ListBackupTargets :many
SELECT * FROM backup_targets
ORDER BY created_at;

-- name: GetBackupTarget :one
SELECT * FROM backup_targets
WHERE id = ?;

-- name: CreateBackupTarget :one
INSERT INTO backup_targets (id, name, provider, config, path, include_videos, include_database, enabled)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateBackupTarget :one
UPDATE backup_targets
SET name = ?, config = ?, path = ?, include_videos = ?, include_database = ?, enabled = ?,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = ?
RETURNING *;

-- name: SetBackupTargetStatus :one
UPDATE backup_targets
SET last_run_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), last_status = ?, last_error = ?
WHERE id = ?
RETURNING *;

-- name: DeleteBackupTarget :exec
DELETE FROM backup_targets
WHERE id = ?;

-- name: ListStoredFileNames :many
SELECT file_name AS name FROM video_files
WHERE status = 'completed' AND file_name IS NOT NULL
UNION
SELECT thumbnail_file_name AS name FROM videos
WHERE thumbnail_file_name IS NOT NULL;
