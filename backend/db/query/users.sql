-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, verified, avatar, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: GetUserByEmailOrUsername :one
SELECT * FROM users WHERE email = $1 OR username = $2 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET username = COALESCE(sqlc.narg(username), username),
    email = COALESCE(sqlc.narg(email), email),
    introduction = COALESCE(sqlc.narg(introduction), introduction),
    avatar = COALESCE(sqlc.narg(avatar), avatar),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetUserVerified :exec
UPDATE users SET verified = TRUE, updated_at = NOW() WHERE id = $1;

-- name: SetUserPasswordHash :exec
UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1;

-- name: SetUserBanned :exec
UPDATE users SET banned = $2, updated_at = NOW() WHERE id = $1;

-- name: SetUserRole :exec
UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: SearchUsers :many
SELECT * FROM users
WHERE username ILIKE $1 OR email ILIKE $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountNewUsersInRange :one
SELECT COUNT(*) FROM users
WHERE created_at >= $1 AND created_at < $2;

-- name: GetFollowerIDs :many
SELECT follower_id FROM user_subscriptions WHERE following_id = $1;

-- name: GetFollowingIDs :many
SELECT following_id FROM user_subscriptions WHERE follower_id = $1;

-- name: GetFollowers :many
SELECT u.* FROM users u
JOIN user_subscriptions us ON us.follower_id = u.id
WHERE us.following_id = $1
ORDER BY us.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetFollowing :many
SELECT u.* FROM users u
JOIN user_subscriptions us ON us.following_id = u.id
WHERE us.follower_id = $1
ORDER BY us.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetFollowerCount :one
SELECT COUNT(*) FROM user_subscriptions WHERE following_id = $1;

-- name: GetFollowingCount :one
SELECT COUNT(*) FROM user_subscriptions WHERE follower_id = $1;

-- name: IsFollowing :one
SELECT EXISTS(
    SELECT 1 FROM user_subscriptions WHERE follower_id = $1 AND following_id = $2
);

-- name: FollowUser :exec
INSERT INTO user_subscriptions (follower_id, following_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnfollowUser :exec
DELETE FROM user_subscriptions WHERE follower_id = $1 AND following_id = $2;

-- name: CreateOAuthAccount :exec
INSERT INTO oauth_accounts (user_id, provider, provider_id)
VALUES ($1, $2, $3)
ON CONFLICT (provider, provider_id) DO NOTHING;

-- name: GetOAuthAccount :one
SELECT * FROM oauth_accounts WHERE provider = $1 AND provider_id = $2 LIMIT 1;

-- name: RecordUserVisit :exec
INSERT INTO user_visits (visit_date, user_id)
VALUES (CURRENT_DATE, $1)
ON CONFLICT DO NOTHING;

-- name: CountDailyActiveUsers :one
SELECT COUNT(DISTINCT user_id) FROM user_visits WHERE visit_date = $1;

-- name: CountDAUInRange :many
SELECT visit_date, COUNT(DISTINCT user_id) AS count
FROM user_visits
WHERE visit_date >= $1 AND visit_date <= $2
GROUP BY visit_date
ORDER BY visit_date;

-- name: GetAdminUsers :many
SELECT * FROM users WHERE role = 'ADMIN' ORDER BY created_at;
