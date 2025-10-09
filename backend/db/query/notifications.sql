-- name: CreateNotification :one
INSERT INTO notifications (type, user_id, from_user_id, post_id, comment_id, content)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id = $1 LIMIT 1;

-- name: ListNotifications :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = FALSE;

-- name: MarkNotificationsRead :exec
UPDATE notifications SET read = TRUE WHERE user_id = $1 AND read = FALSE;

-- name: MarkNotificationReadByID :exec
UPDATE notifications SET read = TRUE WHERE id = $1;

-- name: GetNotificationPreferences :many
SELECT * FROM notification_preferences WHERE user_id = $1;

-- name: UpsertNotificationPreference :exec
INSERT INTO notification_preferences (user_id, type, in_app_enabled, push_enabled)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, type) DO UPDATE
SET in_app_enabled = EXCLUDED.in_app_enabled,
    push_enabled   = EXCLUDED.push_enabled;

-- name: GetEmailPreferences :many
SELECT * FROM email_preferences WHERE user_id = $1;

-- name: UpsertEmailPreference :exec
INSERT INTO email_preferences (user_id, type, enabled)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, type) DO UPDATE
SET enabled = EXCLUDED.enabled;

-- name: GetNotificationPref :one
SELECT * FROM notification_preferences WHERE user_id = $1 AND type = $2 LIMIT 1;

-- name: GetEmailPref :one
SELECT * FROM email_preferences WHERE user_id = $1 AND type = $2 LIMIT 1;
