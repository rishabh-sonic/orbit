-- name: CreatePost :one
INSERT INTO posts (title, content, author_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetPostByIDIncludeDeleted :one
SELECT * FROM posts WHERE id = $1 LIMIT 1;

-- name: UpdatePost :one
UPDATE posts
SET title      = COALESCE(sqlc.narg(title), title),
    content    = COALESCE(sqlc.narg(content), content),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePost :exec
UPDATE posts SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: SetPostClosed :exec
UPDATE posts SET closed = $2, updated_at = NOW() WHERE id = $1;

-- name: PinPost :exec
UPDATE posts SET pinned_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: UnpinPost :exec
UPDATE posts SET pinned_at = NULL, updated_at = NOW() WHERE id = $1;

-- name: IncrementPostViews :exec
UPDATE posts SET views = views + 1 WHERE id = $1;

-- name: IncrementPostCommentCount :exec
UPDATE posts
SET comment_count = comment_count + 1,
    last_reply_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: DecrementPostCommentCount :exec
UPDATE posts
SET comment_count = GREATEST(comment_count - 1, 0),
    updated_at = NOW()
WHERE id = $1;

-- name: ListPosts :many
SELECT * FROM posts
WHERE deleted_at IS NULL
ORDER BY pinned_at DESC NULLS LAST, created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListPostsByAuthor :many
SELECT * FROM posts
WHERE author_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListRecentPosts :many
SELECT * FROM posts
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListFeaturedPosts :many
SELECT * FROM posts
WHERE deleted_at IS NULL AND pinned_at IS NOT NULL
ORDER BY pinned_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPosts :one
SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL;

-- name: CountPostsByAuthor :one
SELECT COUNT(*) FROM posts WHERE author_id = $1 AND deleted_at IS NULL;

-- name: CountPostsInRange :one
SELECT COUNT(*) FROM posts
WHERE created_at >= $1 AND created_at < $2 AND deleted_at IS NULL;

-- name: RecordPostRead :exec
INSERT INTO post_reads (post_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: HasUserReadPost :one
SELECT EXISTS(
    SELECT 1 FROM post_reads WHERE post_id = $1 AND user_id = $2
);
