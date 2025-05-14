-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    NOW() + INTERVAL '60 days',
    null
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT 
    token, 
    created_at, 
    updated_at, 
    user_id, 
    expires_at,
    revoked_at
FROM refresh_tokens 
WHERE token = $1;


-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1;