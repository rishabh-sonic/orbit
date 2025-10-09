-- name: CreateConversation :one
INSERT INTO message_conversations DEFAULT VALUES
RETURNING *;

-- name: GetConversationByID :one
SELECT * FROM message_conversations WHERE id = $1 LIMIT 1;

-- name: GetConversationBetweenUsers :one
SELECT mc.* FROM message_conversations mc
JOIN message_participants mp1 ON mp1.conversation_id = mc.id AND mp1.user_id = $1
JOIN message_participants mp2 ON mp2.conversation_id = mc.id AND mp2.user_id = $2
LIMIT 1;

-- name: ListConversationsForUser :many
SELECT mc.* FROM message_conversations mc
JOIN message_participants mp ON mp.conversation_id = mc.id
WHERE mp.user_id = $1
ORDER BY mc.last_message_at DESC NULLS LAST
LIMIT $2 OFFSET $3;

-- name: AddParticipant :exec
INSERT INTO message_participants (conversation_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetParticipant :one
SELECT * FROM message_participants WHERE conversation_id = $1 AND user_id = $2 LIMIT 1;

-- name: GetParticipants :many
SELECT * FROM message_participants WHERE conversation_id = $1;

-- name: GetOtherParticipantUserID :one
SELECT user_id FROM message_participants
WHERE conversation_id = $1 AND user_id != $2
LIMIT 1;

-- name: IncrementUnreadCounts :exec
UPDATE message_participants
SET unread_count = unread_count + 1
WHERE conversation_id = $1 AND user_id != $2;

-- name: MarkConversationRead :exec
UPDATE message_participants
SET unread_count = 0, last_read_at = NOW()
WHERE conversation_id = $1 AND user_id = $2;

-- name: GetTotalUnreadCount :one
SELECT COALESCE(SUM(unread_count), 0)::BIGINT FROM message_participants WHERE user_id = $1;

-- name: UpdateConversationLastMessage :exec
UPDATE message_conversations SET last_message_at = NOW() WHERE id = $1;

-- name: CreateMessage :one
INSERT INTO messages (conversation_id, sender_id, content, reply_to_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMessageByID :one
SELECT * FROM messages WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: ListMessages :many
SELECT * FROM messages
WHERE conversation_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: SoftDeleteMessage :exec
UPDATE messages SET deleted_at = NOW() WHERE id = $1;
