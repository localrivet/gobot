-- name: CreateNotification :one
INSERT INTO notifications (id, user_id, type, title, body, action_url, icon, created_at)
VALUES (sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(type), sqlc.arg(title), sqlc.narg(body), sqlc.narg(action_url), sqlc.narg(icon), strftime('%s', 'now'))
RETURNING *;

-- name: GetNotification :one
SELECT * FROM notifications WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) LIMIT 1;

-- name: ListUserNotifications :many
SELECT * FROM notifications
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

-- name: ListUnreadNotifications :many
SELECT * FROM notifications
WHERE user_id = sqlc.arg(user_id) AND read_at IS NULL
ORDER BY created_at DESC
LIMIT sqlc.arg(page_size);

-- name: CountUnreadNotifications :one
SELECT COUNT(*) as count FROM notifications
WHERE user_id = sqlc.arg(user_id) AND read_at IS NULL;

-- name: MarkNotificationRead :exec
UPDATE notifications
SET read_at = strftime('%s', 'now')
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = strftime('%s', 'now')
WHERE user_id = sqlc.arg(user_id) AND read_at IS NULL;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: DeleteOldNotifications :exec
DELETE FROM notifications WHERE created_at < sqlc.arg(before);

-- name: GetInappNotificationsSetting :one
SELECT inapp_notifications FROM user_preferences WHERE user_id = sqlc.arg(user_id) LIMIT 1;

-- name: SetInappNotificationsSetting :exec
UPDATE user_preferences
SET inapp_notifications = sqlc.arg(enabled),
    updated_at = strftime('%s', 'now')
WHERE user_id = sqlc.arg(user_id);
