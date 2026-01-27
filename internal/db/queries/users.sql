-- name: GetUserByID :one
SELECT id, email, password_hash, name, avatar_url, email_verified,
       email_verify_token, email_verify_expires, password_reset_token,
       password_reset_expires, created_at, updated_at
FROM users
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, avatar_url, email_verified,
       email_verify_token, email_verify_expires, password_reset_token,
       password_reset_expires, created_at, updated_at
FROM users
WHERE LOWER(email) = LOWER(sqlc.arg(email))
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    id, email, password_hash, name, created_at, updated_at
) VALUES (
    sqlc.arg(id), sqlc.arg(email), sqlc.arg(password_hash), sqlc.arg(name),
    strftime('%s', 'now'), strftime('%s', 'now')
)
RETURNING id, email, password_hash, name, avatar_url, email_verified,
          email_verify_token, email_verify_expires, password_reset_token,
          password_reset_expires, created_at, updated_at;

-- name: UpdateUser :exec
UPDATE users
SET name = COALESCE(sqlc.narg(name), name),
    avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
    updated_at = strftime('%s', 'now')
WHERE id = sqlc.arg(id);

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = sqlc.arg(password_hash),
    password_reset_token = NULL,
    password_reset_expires = NULL,
    updated_at = strftime('%s', 'now')
WHERE id = sqlc.arg(id);

-- name: DeleteUser :exec
DELETE FROM users WHERE id = sqlc.arg(id);

-- name: SetEmailVerified :exec
UPDATE users
SET email_verified = 1,
    email_verify_token = NULL,
    email_verify_expires = NULL,
    updated_at = strftime('%s', 'now')
WHERE id = sqlc.arg(id);

-- name: SetEmailVerifyToken :exec
UPDATE users
SET email_verify_token = sqlc.arg(token),
    email_verify_expires = sqlc.arg(expires),
    updated_at = strftime('%s', 'now')
WHERE id = sqlc.arg(id);

-- name: GetUserByEmailVerifyToken :one
SELECT id, email, password_hash, name, avatar_url, email_verified,
       email_verify_token, email_verify_expires, password_reset_token,
       password_reset_expires, created_at, updated_at
FROM users
WHERE email_verify_token = sqlc.arg(token)
  AND email_verify_expires > strftime('%s', 'now')
LIMIT 1;

-- name: SetPasswordResetToken :exec
UPDATE users
SET password_reset_token = sqlc.arg(token),
    password_reset_expires = sqlc.arg(expires),
    updated_at = strftime('%s', 'now')
WHERE id = sqlc.arg(id);

-- name: GetUserByPasswordResetToken :one
SELECT id, email, password_hash, name, avatar_url, email_verified,
       email_verify_token, email_verify_expires, password_reset_token,
       password_reset_expires, created_at, updated_at
FROM users
WHERE password_reset_token = sqlc.arg(token)
  AND password_reset_expires > strftime('%s', 'now')
LIMIT 1;

-- name: CheckEmailExists :one
SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END as found
FROM users WHERE LOWER(email) = LOWER(sqlc.arg(email));

-- Admin queries

-- name: CountUsers :one
SELECT COUNT(*) as total FROM users;

-- name: CountUsersCreatedAfter :one
SELECT COUNT(*) as total FROM users WHERE created_at >= sqlc.arg(after);

-- name: ListUsersPaginated :many
SELECT id, email, name, avatar_url, email_verified, created_at, updated_at
FROM users
WHERE (sqlc.arg(search) = '' OR LOWER(email) LIKE '%' || LOWER(sqlc.arg(search)) || '%' OR LOWER(name) LIKE '%' || LOWER(sqlc.arg(search)) || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);
