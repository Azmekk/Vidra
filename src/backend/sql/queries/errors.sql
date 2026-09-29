-- name: CreateError :exec
INSERT INTO errors (id, video_id, file_id, error_message, command, output)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListRecentErrors :many
SELECT * FROM errors
WHERE CAST(sqlc.arg('search') AS TEXT) = ''
   OR error_message LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
   OR command LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
   OR video_id LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountErrors :one
SELECT COUNT(*) FROM errors
WHERE CAST(sqlc.arg('search') AS TEXT) = ''
   OR error_message LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
   OR command LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%'
   OR video_id LIKE '%' || CAST(sqlc.arg('search') AS TEXT) || '%';
