package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/email"
	"github.com/rishabh-sonic/orbit/internal/push"
	"github.com/rishabh-sonic/orbit/pkg/rabbitmq"
)

// WirePayload is what gets published to RabbitMQ and consumed by the WS service.
type WirePayload struct {
	NotificationID uuid.UUID `json:"notification_id"`
	Type           string    `json:"type"`
	UserID         uuid.UUID `json:"user_id"`
	Username       string    `json:"username"`
	PostID         *uuid.UUID `json:"post_id,omitempty"`
	CommentID      *uuid.UUID `json:"comment_id,omitempty"`
	FromUserID     *uuid.UUID `json:"from_user_id,omitempty"`
	Content        string    `json:"content,omitempty"`
}

type Service struct {
	q      db.Querier
	mq     *rabbitmq.Client
	email  *email.Sender
	push   *push.Service
}

func NewService(q db.Querier, mq *rabbitmq.Client, emailSvc *email.Sender, pushSvc *push.Service) *Service {
	return &Service{q: q, mq: mq, email: emailSvc, push: pushSvc}
}

// Notify creates an in-app notification, then publishes to RabbitMQ for real-time delivery.
func (s *Service) Notify(ctx context.Context, notifType string, userID uuid.UUID, username string, fromUserID, postID, commentID *uuid.UUID, content string) {
	// Check in-app preference (default: enabled)
	pref, err := s.q.GetNotificationPref(ctx, db.GetNotificationPrefParams{UserID: userID, Type: notifType})
	if err == nil && !pref.InAppEnabled {
		return
	}

	n, err := s.q.CreateNotification(ctx, db.CreateNotificationParams{
		Type:        notifType,
		UserID:      userID,
		FromUserID:  nullUUID(fromUserID),
		PostID:      nullUUID(postID),
		CommentID:   nullUUID(commentID),
		Content:     sql.NullString{String: content, Valid: content != ""},
	})
	if err != nil {
		slog.Error("create notification", "err", err)
		return
	}

	// Publish to RabbitMQ for WS delivery
	payload := WirePayload{
		NotificationID: n.ID,
		Type:           notifType,
		UserID:         userID,
		Username:       username,
		Content:        content,
	}
	if postID != nil {
		payload.PostID = postID
	}
	if commentID != nil {
		payload.CommentID = commentID
	}
	if fromUserID != nil {
		payload.FromUserID = fromUserID
	}

	body, _ := json.Marshal(payload)
	if err := s.mq.Publish(username, body); err != nil {
		slog.Error("publish notification", "err", err)
	}

	// Email for COMMENT_REPLY
	if notifType == TypeCommentReply {
		go s.sendEmailNotification(userID, notifType, content)
	}

	// Push notification
	go s.sendPushNotification(ctx, userID, notifType, content)
}

func (s *Service) sendEmailNotification(userID uuid.UUID, notifType, content string) {
	ctx := context.Background()
	pref, err := s.q.GetEmailPref(ctx, db.GetEmailPrefParams{UserID: userID, Type: notifType})
	if err == nil && !pref.Enabled {
		return
	}
	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		return
	}
	if err := s.email.Send(user.Email, "New notification: "+notifType, "<p>"+content+"</p>"); err != nil {
		slog.Error("send email notification", "err", err)
	}
}

func (s *Service) sendPushNotification(ctx context.Context, userID uuid.UUID, notifType, content string) {
	pref, err := s.q.GetNotificationPref(ctx, db.GetNotificationPrefParams{UserID: userID, Type: notifType})
	if err == nil && !pref.PushEnabled {
		return
	}
	subs, err := s.q.GetPushSubscriptionsByUserID(ctx, userID)
	if err != nil || len(subs) == 0 {
		return
	}
	for _, sub := range subs {
		if err := s.push.Send(sub, notifType, content); err != nil {
			slog.Error("send push notification", "err", err, "endpoint", sub.Endpoint)
		}
	}
}

// List returns notifications for the user.
func (s *Service) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]db.Notification, error) {
	return s.q.ListNotifications(ctx, db.ListNotificationsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
}

// UnreadCount returns the unread notification count.
func (s *Service) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.q.CountUnreadNotifications(ctx, userID)
}

// MarkRead marks all notifications as read for the user.
func (s *Service) MarkRead(ctx context.Context, userID uuid.UUID) error {
	return s.q.MarkNotificationsRead(ctx, userID)
}

func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) ([]db.NotificationPreference, error) {
	return s.q.GetNotificationPreferences(ctx, userID)
}

func (s *Service) UpdatePreference(ctx context.Context, userID uuid.UUID, notifType string, inApp, pushEnabled bool) error {
	return s.q.UpsertNotificationPreference(ctx, db.UpsertNotificationPreferenceParams{
		UserID:       userID,
		Type:         notifType,
		InAppEnabled: inApp,
		PushEnabled:  pushEnabled,
	})
}

func (s *Service) GetEmailPreferences(ctx context.Context, userID uuid.UUID) ([]db.EmailPreference, error) {
	return s.q.GetEmailPreferences(ctx, userID)
}

func (s *Service) UpdateEmailPreference(ctx context.Context, userID uuid.UUID, notifType string, enabled bool) error {
	return s.q.UpsertEmailPreference(ctx, db.UpsertEmailPreferenceParams{
		UserID:  userID,
		Type:    notifType,
		Enabled: enabled,
	})
}

func nullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}
