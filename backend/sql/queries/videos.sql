-- name: CreateVideo :one
INSERT INTO videos (id, name, source_title, original_url)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetVideo :one
SELECT * FROM videos
WHERE id = ?;

-- name: ListVideos :many
WITH params AS (SELECT CAST(sqlc.arg('ordering') AS TEXT) AS ordering)
SELECT videos.* FROM videos, params
WHERE CAST(sqlc.arg('search') AS TEXT) = ''
   OR name LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
   OR source_title LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
   OR original_url LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
ORDER BY
    CASE WHEN params.ordering = 'name_asc' THEN name END ASC,
    CASE WHEN params.ordering = 'name_desc' THEN name END DESC,
    CASE WHEN params.ordering = 'created_at_asc' THEN created_at END ASC,
    created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountVideos :one
SELECT COUNT(*) FROM videos
WHERE CAST(sqlc.arg('search') AS TEXT) = ''
   OR name LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
   OR source_title LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
   OR original_url LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%';

-- name: UpdateVideoName :one
UPDATE videos
SET name = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = ?
RETURNING *;

-- name: UpdateVideoSource :one
UPDATE videos
SET source_title = COALESCE(sqlc.narg('source_title'), source_title),
    thumbnail_file_name = COALESCE(sqlc.narg('thumbnail_file_name'), thumbnail_file_name),
    duration = COALESCE(sqlc.narg('duration'), duration),
    uploader = COALESCE(sqlc.narg('uploader'), uploader),
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: SetPrimaryFile :one
UPDATE videos
SET primary_file_id = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = ?
RETURNING *;

-- name: DeleteVideo :exec
DELETE FROM videos
WHERE id = ?;
