package message_test

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
	"github.com/rishabh-sonic/orbit/internal/message"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newMsgRouter(q *mock.Querier) http.Handler {
	svc := message.NewService(q)
	h := message.NewHandler(svc)
	jwtSvc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc))
	r.With(middleware.RequireAuth).Get("/api/messages/conversations", h.ListConversations)
	r.With(middleware.RequireAuth).Get("/api/messages/conversations/{id}", h.GetConversation)
	r.With(middleware.RequireAuth).Post("/api/messages", h.StartConversation)
	r.With(middleware.RequireAuth).Post("/api/messages/conversations/{id}/messages", h.SendMessage)
	r.With(middleware.RequireAuth).Get("/api/messages/conversations/{id}/messages", h.ListMessages)
	r.With(middleware.RequireAuth).Post("/api/messages/conversations/{id}/read", h.MarkRead)
	r.With(middleware.RequireAuth).Get("/api/messages/unread-count", h.UnreadCount)
	return r
}

func msgToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, _ := svc.GenerateToken(userID, "alice", "USER")
	return tok
}

func msgPost(t *testing.T, r http.Handler, path, bearer string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearer)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func msgGet(t *testing.T, r http.Handler, path, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+bearer)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// --- Auth enforcement ---

func TestMessages_RequireAuth(t *testing.T) {
	endpoints := []string{
		"/api/messages/conversations",
		"/api/messages/unread-count",
	}
	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			rr := httptest.NewRecorder()
			newMsgRouter(&mock.Querier{}).ServeHTTP(rr, req)
			if rr.Code != http.StatusUnauthorized {
				t.Errorf("status: got %d, want 401", rr.Code)
			}
		})
	}
}

// --- ListConversations ---

func TestListConversations_Success(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()
	q := &mock.Querier{
		ListConversationsForUserFn: func(_ context.Context, arg db.ListConversationsForUserParams) ([]db.MessageConversation, error) {
			return []db.MessageConversation{
				{ID: convID},
			}, nil
		},
	}
	rr := msgGet(t, newMsgRouter(q), "/api/messages/conversations", msgToken(t, userID))
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestListConversations_Empty(t *testing.T) {
	userID := uuid.New()
	rr := msgGet(t, newMsgRouter(&mock.Querier{}), "/api/messages/conversations", msgToken(t, userID))
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- GetConversation ---

func TestGetConversation_Success(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()
	q := &mock.Querier{
		GetConversationByIDFn: func(_ context.Context, id uuid.UUID) (db.MessageConversation, error) {
			return db.MessageConversation{ID: id}, nil
		},
		GetParticipantFn: func(_ context.Context, arg db.GetParticipantParams) (db.MessageParticipant, error) {
			return db.MessageParticipant{ConversationID: arg.ConversationID, UserID: arg.UserID}, nil
		},
	}
	rr := msgGet(t, newMsgRouter(q), "/api/messages/conversations/"+convID.String(), msgToken(t, userID))
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestGetConversation_NotFound(t *testing.T) {
	userID := uuid.New()
	rr := msgGet(t, newMsgRouter(&mock.Querier{}), "/api/messages/conversations/"+uuid.New().String(), msgToken(t, userID))
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

func TestGetConversation_NotParticipant(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()
	q := &mock.Querier{
		GetConversationByIDFn: func(_ context.Context, id uuid.UUID) (db.MessageConversation, error) {
			return db.MessageConversation{ID: id}, nil
		},
		// GetParticipantFn is nil → returns ErrNoRows → forbidden
	}
	rr := msgGet(t, newMsgRouter(q), "/api/messages/conversations/"+convID.String(), msgToken(t, userID))
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404 (service converts forbidden to not-found here)", rr.Code)
	}
}

func TestGetConversation_InvalidID(t *testing.T) {
	userID := uuid.New()
	rr := msgGet(t, newMsgRouter(&mock.Querier{}), "/api/messages/conversations/not-a-uuid", msgToken(t, userID))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- StartConversation ---

func TestStartConversation_NewConversation(t *testing.T) {
	senderID := uuid.New()
	recipientID := uuid.New()
	convID := uuid.New()
	q := &mock.Querier{
		// No existing conversation
		CreateConversationFn: func(_ context.Context) (db.MessageConversation, error) {
			return db.MessageConversation{ID: convID}, nil
		},
	}
	rr := msgPost(t, newMsgRouter(q), "/api/messages", msgToken(t, senderID), map[string]string{
		"recipient_id": recipientID.String(),
	})
	if rr.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201 — body: %s", rr.Code, rr.Body)
	}
}

func TestStartConversation_WithInitialMessage(t *testing.T) {
	senderID := uuid.New()
	recipientID := uuid.New()
	convID := uuid.New()
	q := &mock.Querier{
		CreateConversationFn: func(_ context.Context) (db.MessageConversation, error) {
			return db.MessageConversation{ID: convID}, nil
		},
		GetParticipantFn: func(_ context.Context, _ db.GetParticipantParams) (db.MessageParticipant, error) {
			return db.MessageParticipant{}, nil
		},
	}
	rr := msgPost(t, newMsgRouter(q), "/api/messages", msgToken(t, senderID), map[string]string{
		"recipient_id": recipientID.String(),
		"content":      "Hello there!",
	})
	if rr.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201 — body: %s", rr.Code, rr.Body)
	}
}

func TestStartConversation_MissingRecipient(t *testing.T) {
	senderID := uuid.New()
	rr := msgPost(t, newMsgRouter(&mock.Querier{}), "/api/messages", msgToken(t, senderID), map[string]string{})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

func TestStartConversation_InvalidRecipientID(t *testing.T) {
	senderID := uuid.New()
	rr := msgPost(t, newMsgRouter(&mock.Querier{}), "/api/messages", msgToken(t, senderID), map[string]string{
		"recipient_id": "not-a-uuid",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- SendMessage ---

func TestSendMessage_Success(t *testing.T) {
	senderID := uuid.New()
	convID := uuid.New()
	q := &mock.Querier{
		GetParticipantFn: func(_ context.Context, _ db.GetParticipantParams) (db.MessageParticipant, error) {
			return db.MessageParticipant{}, nil
		},
	}
	rr := msgPost(t, newMsgRouter(q), "/api/messages/conversations/"+convID.String()+"/messages", msgToken(t, senderID), map[string]string{
		"content": "Hello!",
	})
	if rr.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201 — body: %s", rr.Code, rr.Body)
	}
}

func TestSendMessage_NotParticipant(t *testing.T) {
	senderID := uuid.New()
	convID := uuid.New()
	// GetParticipant returns error (not a participant)
	rr := msgPost(t, newMsgRouter(&mock.Querier{}), "/api/messages/conversations/"+convID.String()+"/messages", msgToken(t, senderID), map[string]string{
		"content": "Hello!",
	})
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestSendMessage_EmptyContent(t *testing.T) {
	senderID := uuid.New()
	convID := uuid.New()
	rr := msgPost(t, newMsgRouter(&mock.Querier{}), "/api/messages/conversations/"+convID.String()+"/messages", msgToken(t, senderID), map[string]string{
		"content": "",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

func TestSendMessage_InvalidConvID(t *testing.T) {
	rr := msgPost(t, newMsgRouter(&mock.Querier{}), "/api/messages/conversations/bad-id/messages", msgToken(t, uuid.New()), map[string]string{
		"content": "hello",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- ListMessages ---

func TestListMessages_Success(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()
	q := &mock.Querier{
		GetParticipantFn: func(_ context.Context, _ db.GetParticipantParams) (db.MessageParticipant, error) {
			return db.MessageParticipant{}, nil
		},
		ListMessagesFn: func(_ context.Context, _ db.ListMessagesParams) ([]db.Message, error) {
			return []db.Message{
				{ID: uuid.New(), Content: "Hi!", SenderID: userID, ConversationID: convID},
			}, nil
		},
	}
	rr := msgGet(t, newMsgRouter(q), "/api/messages/conversations/"+convID.String()+"/messages", msgToken(t, userID))
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestListMessages_NotParticipant(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()
	rr := msgGet(t, newMsgRouter(&mock.Querier{}), "/api/messages/conversations/"+convID.String()+"/messages", msgToken(t, userID))
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

// --- MarkRead ---

func TestMarkConversationRead_Success(t *testing.T) {
	userID := uuid.New()
	convID := uuid.New()
	rr := msgPost(t, newMsgRouter(&mock.Querier{}), "/api/messages/conversations/"+convID.String()+"/read", msgToken(t, userID), nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- UnreadCount ---

func TestMsgUnreadCount_Success(t *testing.T) {
	userID := uuid.New()
	q := &mock.Querier{
		GetTotalUnreadCountFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
			return 3, nil
		},
	}
	rr := msgGet(t, newMsgRouter(q), "/api/messages/unread-count", msgToken(t, userID))
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	var resp struct {
		Data struct{ Count int64 `json:"count"` } `json:"data"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Data.Count != 3 {
		t.Errorf("count: got %d, want 3", resp.Data.Count)
	}
}
