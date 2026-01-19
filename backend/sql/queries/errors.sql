-- name: CreateError :one
INSERT INTO errors (
    video_id, error_message, command, output
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListRecentErrors :many
SELECT * FROM errors
WHERE (error_message ILIKE '%' || sqlc.arg('search') || '%' 
   OR command ILIKE '%' || sqlc.arg('search') || '%'
   OR CAST(video_id AS TEXT) ILIKE '%' || sqlc.arg('search') || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_val');
