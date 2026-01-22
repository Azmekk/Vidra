-- name: GetSettings :one
SELECT * FROM settings WHERE id = 1;

-- name: UpdateSettings :one
UPDATE settings SET
    proxy_url = $1,
    default_re_encode = $2,
    default_video_codec = $3,
    default_audio_codec = $4,
    default_crf = $5,
    theme = $6,
    updated_at = NOW()
WHERE id = 1
RETURNING *;
