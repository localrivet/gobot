-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    id, user_id, token_hash, expires_at, created_at
) VALUES (
    sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(expires_at),
    strftime('%s', 'now')
)
RETURNING id, user_id, token_hash, expires_at, created_at;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, created_at
FROM refresh_tokens
WHERE token_hash = sqlc.arg(token_hash)
  AND expires_at > strftime('%s', 'now')
LIMIT 1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token_hash = sqlc.arg(token_hash);

-- name: DeleteRefreshTokensByUserID :exec
DELETE FROM refresh_tokens WHERE user_id = sqlc.arg(user_id);

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens WHERE expires_at <= strftime('%s', 'now');
