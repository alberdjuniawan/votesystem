-- name: CreateOption :one
INSERT INTO options (room_id, label, description, metadata, media_id, order_num)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListOptionsByRoom :many
SELECT * FROM options
WHERE room_id = $1
ORDER BY order_num ASC, created_at ASC;

-- name: GetOptionByID :one
SELECT * FROM options
WHERE id = $1
LIMIT 1;

-- name: DeleteOption :exec
DELETE FROM options
WHERE id = $1;