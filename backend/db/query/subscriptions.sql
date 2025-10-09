-- name: SubscribeToPost :exec
INSERT INTO post_subscriptions (post_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnsubscribeFromPost :exec
DELETE FROM post_subscriptions WHERE post_id = $1 AND user_id = $2;

-- name: IsSubscribedToPost :one
SELECT EXISTS(
    SELECT 1 FROM post_subscriptions WHERE post_id = $1 AND user_id = $2
);

-- name: GetPostSubscribers :many
SELECT user_id FROM post_subscriptions WHERE post_id = $1;

-- name: GetSubscribedPostIDs :many
SELECT post_id FROM post_subscriptions WHERE user_id = $1;
