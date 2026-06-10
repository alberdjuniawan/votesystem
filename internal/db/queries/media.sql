-- name: CreateMedia :one
INSERT INTO media (uploader_id, filename, original_name, mime_type, size_bytes, storage_path)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMediaByID :one
SELECT * FROM media
WHERE id = $1
LIMIT 1;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE id = $1;