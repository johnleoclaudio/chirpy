-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
  gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING *;

-- name: ListChirps :many
SELECT * FROM chirps
ORDER BY 
  CASE WHEN sqlc.arg('sort_order') = 'desc' THEN created_at END DESC,
  CASE WHEN sqlc.arg('sort_order') = 'asc' THEN created_at END ASC;

-- name: ListChirpsByAuthorID :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY
  CASE WHEN sqlc.arg('sort_order') = 'desc' THEN created_at END DESC,
  CASE WHEN sqlc.arg('sort_order') = 'asc' THEN created_at END ASC;

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps where id = $1 AND user_id = $2;
