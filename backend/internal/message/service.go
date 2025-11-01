package message

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
)

type Service struct {
	q db.Querier
}

func NewService(q db.Querier) *Service {
	return &Service{q: q}
}

// ── Response types ────────────────────────────────────────────────────────────

type ConversationUser struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Avatar   *string   `json:"avatar"`
}

type ConversationResponse struct {
	ID            uuid.UUID        `json:"id"`
	OtherUser     ConversationUser `json:"other_user"`
	LastMessage   *string          `json:"last_message"`
	LastMessageAt *time.Time       `json:"last_message_at"`
	UnreadCount   int32            `json:"unread_count"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *Service) enrichConversation(ctx context.Context, conv db.MessageConversation, myUserID uuid.UUID) (ConversationResponse, error) {
	otherID, err := s.q.GetOtherParticipantUserID(ctx, db.GetOtherParticipantUserIDParams{
		ConversationID: conv.ID,
		UserID:         myUserID,
	})
	if err != nil {
		return ConversationResponse{}, fmt.Errorf("get other participant: %w", err)
	}

	other, err := s.q.GetUserByID(ctx, otherID)
	if err != nil {
		return ConversationResponse{}, fmt.Errorf("get other user: %w", err)
	}

	var avatar *string
	if other.Avatar.Valid && other.Avatar.String != "" {
		v := other.Avatar.String
		avatar = &v
	}

	// Unread count for the current user
	participant, _ := s.q.GetParticipant(ctx, db.GetParticipantParams{
		ConversationID: conv.ID,
		UserID:         myUserID,
	})

	// Last message content
	var lastMsg *string
	msgs, err := s.q.ListMessages(ctx, db.ListMessagesParams{
		ConversationID: conv.ID,
		Limit:          1,
		Offset:         0,
	})
	if err == nil && len(msgs) > 0 {
		v := msgs[0].Content
		lastMsg = &v
	}

	var lastMsgAt *time.Time
	if conv.LastMessageAt.Valid {
		t := conv.LastMessageAt.Time
		lastMsgAt = &t
	}

	return ConversationResponse{
		ID:            conv.ID,
		OtherUser:     ConversationUser{ID: other.ID, Username: other.Username, Avatar: avatar},
		LastMessage:   lastMsg,
		LastMessageAt: lastMsgAt,
		UnreadCount:   participant.UnreadCount,
	}, nil
}

// ── Service methods ───────────────────────────────────────────────────────────

// GetOrCreateConversation finds or creates a DM conversation between two users
// and returns an enriched ConversationResponse for the requesting user.
func (s *Service) GetOrCreateConversation(ctx context.Context, myUserID, otherUserID uuid.UUID) (ConversationResponse, error) {
	conv, err := s.q.GetConversationBetweenUsers(ctx, db.GetConversationBetweenUsersParams{
		UserID:   myUserID,
		UserID_2: otherUserID,
	})
	if err != nil {
		// Create new conversation
		conv, err = s.q.CreateConversation(ctx)
		if err != nil {
			return ConversationResponse{}, fmt.Errorf("create conversation: %w", err)
		}
		if err := s.q.AddParticipant(ctx, db.AddParticipantParams{ConversationID: conv.ID, UserID: myUserID}); err != nil {
			return ConversationResponse{}, err
		}
		if err := s.q.AddParticipant(ctx, db.AddParticipantParams{ConversationID: conv.ID, UserID: otherUserID}); err != nil {
			return ConversationResponse{}, err
		}
	}
	return s.enrichConversation(ctx, conv, myUserID)
}

func (s *Service) ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]ConversationResponse, error) {
	convs, err := s.q.ListConversationsForUser(ctx, db.ListConversationsForUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	result := make([]ConversationResponse, 0, len(convs))
	for _, c := range convs {
		enriched, err := s.enrichConversation(ctx, c, userID)
		if err != nil {
			continue // skip conversations with broken data
		}
		result = append(result, enriched)
	}
	return result, nil
}

func (s *Service) GetConversation(ctx context.Context, convID, userID uuid.UUID) (ConversationResponse, error) {
	conv, err := s.q.GetConversationByID(ctx, convID)
	if err != nil {
		return ConversationResponse{}, ErrNotFound
	}
	if _, err := s.q.GetParticipant(ctx, db.GetParticipantParams{
		ConversationID: convID, UserID: userID,
	}); err != nil {
		return ConversationResponse{}, ErrForbidden
	}
	return s.enrichConversation(ctx, conv, userID)
}

func (s *Service) SendMessage(ctx context.Context, convID, senderID uuid.UUID, content string, replyToID *uuid.UUID) (db.Message, error) {
	if _, err := s.q.GetParticipant(ctx, db.GetParticipantParams{
		ConversationID: convID, UserID: senderID,
	}); err != nil {
		return db.Message{}, ErrForbidden
	}

	var replyTo uuid.NullUUID
	if replyToID != nil {
		replyTo = uuid.NullUUID{UUID: *replyToID, Valid: true}
	}

	msg, err := s.q.CreateMessage(ctx, db.CreateMessageParams{
		ConversationID: convID,
		SenderID:       senderID,
		Content:        content,
		ReplyToID:      replyTo,
	})
	if err != nil {
		return db.Message{}, err
	}

	_ = s.q.UpdateConversationLastMessage(ctx, convID)
	_ = s.q.IncrementUnreadCounts(ctx, db.IncrementUnreadCountsParams{
		ConversationID: convID, UserID: senderID,
	})
	return msg, nil
}

func (s *Service) ListMessages(ctx context.Context, convID, userID uuid.UUID, limit, offset int) ([]db.Message, error) {
	if _, err := s.q.GetParticipant(ctx, db.GetParticipantParams{
		ConversationID: convID, UserID: userID,
	}); err != nil {
		return nil, ErrForbidden
	}
	return s.q.ListMessages(ctx, db.ListMessagesParams{
		ConversationID: convID,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
}

func (s *Service) MarkRead(ctx context.Context, convID, userID uuid.UUID) error {
	return s.q.MarkConversationRead(ctx, db.MarkConversationReadParams{
		ConversationID: convID, UserID: userID,
	})
}

func (s *Service) TotalUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.q.GetTotalUnreadCount(ctx, userID)
}

// keep for backward compat (used nowhere now but keeps the build clean)
var _ = sql.NullTime{}
