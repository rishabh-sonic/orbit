package push_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/db/mock"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/internal/push"
	"github.com/rishabh-sonic/orbit/pkg/config"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newPushRouter(q *mock.Querier) http.Handler {
	svc := push.NewService(q, &config.Config{
		WebPushPublicKey:  "BNcRdreALRFXTkOOUHK1EtK2wtZ5MQp5",
		WebPushPrivateKey: "test-private",
		WebPushSubscriber: "mailto:test@test.com",
	})
	h := push.NewHandler(svc)
	jwtSvc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc))
	r.Get("/api/push/public-key", h.PublicKey)
	r.With(middleware.RequireAuth).Post("/api/push/subscribe", h.Subscribe)
	return r
}

func pushToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, _ := svc.GenerateToken(userID, "alice", "USER")
	return tok
}

// --- PublicKey ---

func TestPushPublicKey_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/push/public-key", nil)
	rr := httptest.NewRecorder()
	newPushRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	var resp struct {
		Data struct {
			PublicKey string `json:"public_key"`
		} `json:"data"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Data.PublicKey == "" {
		t.Error("expected non-empty public_key in response")
	}
}

func TestPushPublicKey_NoAuthRequired(t *testing.T) {
	// Public key endpoint should not require authentication
	req := httptest.NewRequest(http.MethodGet, "/api/push/public-key", nil)
	rr := httptest.NewRecorder()
	newPushRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code == http.StatusUnauthorized {
		t.Error("public key endpoint should not require auth")
	}
}

// --- Subscribe ---

func TestPushSubscribe_Success(t *testing.T) {
	userID := uuid.New()
	subscribed := false
	q := &mock.Querier{
		CreatePushSubscriptionFn: func(_ context.Context, arg db.CreatePushSubscriptionParams) (db.PushSubscription, error) {
			subscribed = true
			return db.PushSubscription{ID: uuid.New(), UserID: arg.UserID, Endpoint: arg.Endpoint}, nil
		},
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(map[string]string{
		"endpoint": "https://fcm.googleapis.com/fcm/send/test-endpoint",
		"p256dh":   "test-p256dh",
		"auth":     "test-auth",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pushToken(t, userID))
	rr := httptest.NewRecorder()
	newPushRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
	if !subscribed {
		t.Error("expected CreatePushSubscription to be called")
	}
}

func TestPushSubscribe_MissingEndpoint(t *testing.T) {
	userID := uuid.New()
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(map[string]string{
		"p256dh": "test-p256dh",
		"auth":   "test-auth",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pushToken(t, userID))
	rr := httptest.NewRecorder()
	newPushRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

func TestPushSubscribe_Unauthenticated(t *testing.T) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(map[string]string{"endpoint": "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/push/subscribe", &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	newPushRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}
