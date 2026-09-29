-- name: CreateVideoFile :one
INSERT INTO video_files (id, video_id, kind, source_file_id, label, status, encoding_profile)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetVideoFile :one
SELECT * FROM video_files
WHERE id = ?;

-- name: ListFilesByVideoIDs :many
SELECT * FROM video_files
WHERE video_id IN (sqlc.slice('video_ids'))
ORDER BY created_at;

-- name: UpdateFileStatus :one
UPDATE video_files
SET status = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = ?
RETURNING *;

-- name: UpdateFileMedia :one
UPDATE video_files
SET file_name = sqlc.arg('file_name'),
    label = sqlc.arg('label'),
    status = sqlc.arg('status'),
    container = sqlc.narg('container'),
    video_codec = sqlc.narg('video_codec'),
    audio_codec = sqlc.narg('audio_codec'),
    width = sqlc.narg('width'),
    height = sqlc.narg('height'),
    fps = sqlc.narg('fps'),
    bitrate = sqlc.narg('bitrate'),
    file_size = sqlc.narg('file_size'),
    ios_compatible = sqlc.arg('ios_compatible'),
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: SetFileName :exec
UPDATE video_files
SET file_name = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = ?;

-- name: MarkInterruptedFiles :many
UPDATE video_files
SET status = 'error', updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE status IN ('queued', 'downloading', 'encoding')
RETURNING *;

-- name: DeleteVideoFile :exec
DELETE FROM video_files
WHERE id = ?;
