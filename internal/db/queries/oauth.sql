-- name: CreateOAuthConnection :one
INSERT INTO oauth_connections (id, user_id, provider, provider_user_id, email, name, avatar_url, access_token, refresh_token, expires_at, created_at, updated_at)
VALUES (sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(provider), sqlc.arg(provider_user_id), sqlc.narg(email), sqlc.narg(name), sqlc.narg(avatar_url), sqlc.narg(access_token), sqlc.narg(refresh_token), sqlc.narg(expires_at), strftime('%s', 'now'), strftime('%s', 'now'))
RETURNING *;

-- name: GetOAuthConnectionByProvider :one
SELECT * FROM oauth_connections
WHERE provider = sqlc.arg(provider) AND provider_user_id = sqlc.arg(provider_user_id)
LIMIT 1;

-- name: GetOAuthConnectionByUserAndProvider :one
SELECT * FROM oauth_connections
WHERE user_id = sqlc.arg(user_id) AND provider = sqlc.arg(provider)
LIMIT 1;

-- name: ListUserOAuthConnections :many
SELECT * FROM oauth_connections
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC;

-- name: UpdateOAuthConnection :exec
UPDATE oauth_connections
SET email = COALESCE(sqlc.narg(email), email),
    name = COALESCE(sqlc.narg(name), name),
    avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
    access_token = COALESCE(sqlc.narg(access_token), access_token),
    refresh_token = COALESCE(sqlc.narg(refresh_token), refresh_token),
    expires_at = COALESCE(sqlc.narg(expires_at), expires_at),
    updated_at = strftime('%s', 'now')
WHERE id = sqlc.arg(id);

-- name: DeleteOAuthConnection :exec
DELETE FROM oauth_connections WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: DeleteOAuthConnectionByProvider :exec
DELETE FROM oauth_connections WHERE user_id = sqlc.arg(user_id) AND provider = sqlc.arg(provider);

-- name: CreateUserFromOAuth :one
INSERT INTO users (id, email, password_hash, name, avatar_url, email_verified, created_at, updated_at)
VALUES (sqlc.arg(id), sqlc.arg(email), '', sqlc.arg(name), sqlc.narg(avatar_url), 1, strftime('%s', 'now'), strftime('%s', 'now'))
RETURNING *;
