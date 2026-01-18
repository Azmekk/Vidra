-- name: CreateVideo :one
INSERT INTO videos (
    name, file_name, thumbnail_file_name, original_url, download_status
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetVideo :one
SELECT * FROM videos
WHERE id = $1 LIMIT 1;

-- name: ListVideos :many
SELECT * FROM videos
ORDER BY created_at DESC;

-- name: UpdateVideoStatus :one
UPDATE videos
  set download_status = $2,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateVideoFiles :one
UPDATE videos
  set file_name = $2,
  thumbnail_file_name = $3,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteVideo :exec
DELETE FROM videos
WHERE id = $1;