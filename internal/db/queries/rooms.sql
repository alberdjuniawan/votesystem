-- name: CreateRoom :one
INSERT INTO rooms (owner_id, title, description, type, show_realtime, max_votes, starts_at, ends_at, share_code)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetRoomByID :one
SELECT * FROM rooms
WHERE id = $1
LIMIT 1;

-- name: GetRoomByShareCode :one
SELECT * FROM rooms
WHERE share_code = $1
LIMIT 1;

-- name: ListRoomsByOwner :many
SELECT * FROM rooms
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: UpdateRoomStatus :one
UPDATE rooms
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRoom :exec
DELETE FROM rooms
WHERE id = $1 AND owner_id = $2;