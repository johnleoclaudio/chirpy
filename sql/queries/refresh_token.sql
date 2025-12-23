-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (created_at, updated_at, token, user_id, expires_at)
VALUES (
  NOW(), NOW(), $1, $2, (NOW() + INTERVAL '60 days')
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1 AND revoked_at IS NULL;

-- name: RevokeRefreskToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(),
updated_at = NOW()
WHERE token = $1
RETURNING *;
