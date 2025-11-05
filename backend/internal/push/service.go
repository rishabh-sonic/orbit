package push

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/pkg/config"
)

type Service struct {
	q          db.Querier
	publicKey  string
	privateKey string
	subscriber string
}

func NewService(q db.Querier, cfg *config.Config) *Service {
	return &Service{
		q:          q,
		publicKey:  cfg.WebPushPublicKey,
		privateKey: cfg.WebPushPrivateKey,
		subscriber: cfg.WebPushSubscriber,
	}
}

func (s *Service) PublicKey() string {
	return s.publicKey
}

func (s *Service) Subscribe(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth string) error {
	_, err := s.q.CreatePushSubscription(ctx, db.CreatePushSubscriptionParams{
		UserID:   userID,
		Endpoint: endpoint,
		P256dh:   p256dh,
		Auth:     auth,
	})
	return err
}

type PushMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (s *Service) Send(sub db.PushSubscription, notifType, content string) error {
	if s.publicKey == "" || s.privateKey == "" {
		slog.Debug("web push not configured, skipping")
		return nil
	}

	payload, _ := json.Marshal(PushMessage{Type: notifType, Content: content})

	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}, &webpush.Options{
		Subscriber:      s.subscriber,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             30,
	})
	if err != nil {
		return fmt.Errorf("web push send: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
