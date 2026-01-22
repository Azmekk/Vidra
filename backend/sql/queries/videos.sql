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
WHERE (name ILIKE '%' || sqlc.arg('search') || '%' OR original_url ILIKE '%' || sqlc.arg('search') || '%')
ORDER BY 
    CASE WHEN sqlc.arg('ordering') = 'name_asc' THEN name END ASC,
    CASE WHEN sqlc.arg('ordering') = 'name_desc' THEN name END DESC,
    CASE WHEN sqlc.arg('ordering') = 'created_at_asc' THEN created_at END ASC,
    CASE WHEN sqlc.arg('ordering') = 'status_asc' THEN download_status END ASC,
    CASE WHEN sqlc.arg('ordering') = 'status_desc' THEN download_status END DESC,
    CASE WHEN sqlc.arg('ordering') = 'created_at_desc' OR sqlc.arg('ordering') = '' OR sqlc.arg('ordering') IS NULL THEN created_at END DESC
LIMIT $1 OFFSET $2;

-- name: CountVideos :one
SELECT COUNT(*) FROM videos
WHERE (name ILIKE '%' || sqlc.arg('search') || '%' OR original_url ILIKE '%' || sqlc.arg('search') || '%');

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
  file_size = $4,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateVideoName :one
UPDATE videos
  set name = $2,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteVideo :exec
DELETE FROM videos
WHERE id = $1;