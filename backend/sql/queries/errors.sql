-- name: CreateError :one
INSERT INTO errors (
    video_id, error_message, command, output
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListRecentErrors :many
SELECT * FROM errors
ORDER BY created_at DESC
LIMIT $1;
