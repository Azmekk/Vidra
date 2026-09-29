-- name: GetSettings :one
SELECT * FROM settings
WHERE id = 1;

-- name: UpdateSettings :one
UPDATE settings
SET proxy_url = ?,
    theme = ?,
    prefer_compatible_formats = ?,
    default_encoding = ?,
    keep_original = ?,
    cache_size = ?,
    max_concurrent_downloads = ?,
    max_concurrent_encodes = ?,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = 1
RETURNING *;
