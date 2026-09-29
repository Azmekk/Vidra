-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CreateUser :one
INSERT INTO users (id, username, password_hash)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ?;

-- name: GetUser :one
SELECT * FROM users
WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = ?
WHERE id = ?;

-- name: CreateSession :exec
INSERT INTO sessions (token_hash, user_id, expires_at, last_seen_at, remember, user_agent)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetSession :one
SELECT sqlc.embed(sessions), sqlc.embed(users)
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.token_hash = ? AND sessions.expires_at > sqlc.arg('now');

-- name: TouchSession :exec
UPDATE sessions
SET last_seen_at = ?, expires_at = ?
WHERE token_hash = ?;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token_hash = ?;

-- name: DeleteOtherSessions :exec
DELETE FROM sessions
WHERE user_id = ? AND token_hash != ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= ?;

-- name: CreateAPIToken :one
INSERT INTO api_tokens (id, user_id, name, token_hash)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListAPITokens :many
SELECT * FROM api_tokens
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: GetUserByAPIToken :one
SELECT users.* FROM api_tokens
JOIN users ON users.id = api_tokens.user_id
WHERE api_tokens.token_hash = ?;

-- name: TouchAPIToken :exec
UPDATE api_tokens
SET last_used_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE token_hash = ?;

-- name: DeleteAPIToken :exec
DELETE FROM api_tokens
WHERE id = ? AND user_id = ?;
