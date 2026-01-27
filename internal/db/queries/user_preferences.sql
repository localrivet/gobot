-- name: GetUserPreferences :one
SELECT user_id, email_notifications, marketing_emails, timezone, language, theme, updated_at
FROM user_preferences
WHERE user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: CreateUserPreferences :one
INSERT INTO user_preferences (
    user_id, email_notifications, marketing_emails, timezone, language, theme, updated_at
) VALUES (
    sqlc.arg(user_id), 1, 0, 'UTC', 'en', 'system', strftime('%s', 'now')
)
RETURNING user_id, email_notifications, marketing_emails, timezone, language, theme, updated_at;

-- name: UpdateUserPreferences :exec
UPDATE user_preferences
SET email_notifications = COALESCE(sqlc.narg(email_notifications), email_notifications),
    marketing_emails = COALESCE(sqlc.narg(marketing_emails), marketing_emails),
    timezone = COALESCE(sqlc.narg(timezone), timezone),
    language = COALESCE(sqlc.narg(language), language),
    theme = COALESCE(sqlc.narg(theme), theme),
    updated_at = strftime('%s', 'now')
WHERE user_id = sqlc.arg(user_id);

-- name: DeleteUserPreferences :exec
DELETE FROM user_preferences WHERE user_id = sqlc.arg(user_id);
