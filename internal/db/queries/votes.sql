-- name: CreateVote :one
INSERT INTO votes (room_id, user_id, option_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetVoteByRoomAndUser :one
SELECT * FROM votes
WHERE room_id = $1 AND user_id = $2
LIMIT 1;

-- name: GetVoteCountsByRoom :many
SELECT
    option_id,
    COUNT(*) AS vote_count
FROM votes
WHERE room_id = $1
GROUP BY option_id;

-- name: GetTotalVotesByRoom :one
SELECT COUNT(*) AS total
FROM votes
WHERE room_id = $1;