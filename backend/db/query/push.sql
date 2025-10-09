-- name: CreatePushSubscription :one
INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
VALUES ($1, $2, $3, $4)
ON CONFLICT (endpoint) DO UPDATE
SET p256dh = EXCLUDED.p256dh, auth = EXCLUDED.auth
RETURNING *;

-- name: GetPushSubscriptionsByUserID :many
SELECT * FROM push_subscriptions WHERE user_id = $1;

-- name: DeletePushSubscriptionByEndpoint :exec
DELETE FROM push_subscriptions WHERE endpoint = $1;
