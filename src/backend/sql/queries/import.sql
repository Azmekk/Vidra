-- name: ImportVideo :exec
INSERT INTO videos (id, name, original_url, thumbnail_file_name, primary_file_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ImportVideoFile :exec
INSERT INTO video_files (
    id, video_id, kind, label, status, file_name, container, video_codec, audio_codec,
    width, height, fps, bitrate, file_size, ios_compatible, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ImportError :exec
INSERT INTO errors (id, video_id, error_message, command, output, created_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: CountAllVideos :one
SELECT COUNT(*) FROM videos;

-- name: DeleteAllVideos :exec
DELETE FROM videos;

-- name: DeleteAllErrors :exec
DELETE FROM errors;
