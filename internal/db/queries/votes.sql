-- name: CreateVote :one
INSERT INTO votes (room_id, user_id, option_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetVoteByRoomUserOption :one
SELECT * FROM votes
WHERE room_id = $1 AND user_id = $2 AND option_id = $3
LIMIT 1;

-- name: GetVoteCountByRoomAndUser :one
SELECT COUNT(*) AS total
FROM votes
WHERE room_id = $1 AND user_id = $2;

-- name: GetTotalVotesByRoom :one
SELECT COUNT(*) AS total
FROM votes
WHERE room_id = $1;

-- name: GetVotesByRoomAndUser :many
SELECT * FROM votes
WHERE room_id = $1 AND user_id = $2;