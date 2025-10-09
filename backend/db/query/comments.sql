-- name: CreateComment :one
INSERT INTO comments (content, author_id, post_id, parent_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM comments WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetCommentByIDIncludeDeleted :one
SELECT * FROM comments WHERE id = $1 LIMIT 1;

-- name: UpdateComment :one
UPDATE comments
SET content = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteComment :exec
UPDATE comments SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: PinComment :exec
UPDATE comments SET pinned_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: UnpinComment :exec
UPDATE comments SET pinned_at = NULL, updated_at = NOW() WHERE id = $1;

-- name: ListTopLevelComments :many
SELECT * FROM comments
WHERE post_id = $1 AND parent_id IS NULL AND deleted_at IS NULL
ORDER BY pinned_at DESC NULLS LAST, created_at ASC
LIMIT $2 OFFSET $3;

-- name: ListReplies :many
SELECT * FROM comments
WHERE parent_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: CountTopLevelComments :one
SELECT COUNT(*) FROM comments
WHERE post_id = $1 AND parent_id IS NULL AND deleted_at IS NULL;

-- name: CountCommentsByAuthor :one
SELECT COUNT(*) FROM comments WHERE author_id = $1 AND deleted_at IS NULL;

-- name: ListCommentsByAuthor :many
SELECT * FROM comments
WHERE author_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCommentsInRange :one
SELECT COUNT(*) FROM comments
WHERE created_at >= $1 AND created_at < $2 AND deleted_at IS NULL;
