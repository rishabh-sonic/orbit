package post_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/db/mock"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/internal/post"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newSubscriptionRouter(q *mock.Querier) http.Handler {
	h := post.NewSubscriptionHandler(q)
	jwtSvc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc))
	r.With(middleware.RequireAuth).Post("/api/subscriptions/posts/{postId}", h.Subscribe)
	r.With(middleware.RequireAuth).Delete("/api/subscriptions/posts/{postId}", h.Unsubscribe)
	return r
}

func subToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, _ := svc.GenerateToken(userID, "alice", "USER")
	return tok
}

func TestSubscribeToPost_Success(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()
	subscribed := false
	q := &mock.Querier{
		SubscribeToPostFn: func(_ context.Context, arg db.SubscribeToPostParams) error {
			subscribed = true
			if arg.PostID != postID || arg.UserID != userID {
				t.Errorf("wrong IDs: postID=%v userID=%v", arg.PostID, arg.UserID)
			}
			return nil
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions/posts/"+postID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+subToken(t, userID))
	rr := httptest.NewRecorder()
	newSubscriptionRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
	if !subscribed {
		t.Error("expected SubscribeToPost to be called")
	}
}

func TestSubscribeToPost_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions/posts/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	newSubscriptionRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestSubscribeToPost_InvalidPostID(t *testing.T) {
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions/posts/not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer "+subToken(t, userID))
	rr := httptest.NewRecorder()
	newSubscriptionRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

func TestUnsubscribeFromPost_Success(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()
	unsubscribed := false
	q := &mock.Querier{
		UnsubscribeFromPostFn: func(_ context.Context, arg db.UnsubscribeFromPostParams) error {
			unsubscribed = true
			if arg.PostID != postID || arg.UserID != userID {
				t.Errorf("wrong IDs")
			}
			return nil
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/subscriptions/posts/"+postID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+subToken(t, userID))
	rr := httptest.NewRecorder()
	newSubscriptionRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204", rr.Code)
	}
	if !unsubscribed {
		t.Error("expected UnsubscribeFromPost to be called")
	}
}

func TestUnsubscribeFromPost_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/subscriptions/posts/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	newSubscriptionRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}
