package notification_test

import (
	"context"
	"bytes"
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
	"github.com/rishabh-sonic/orbit/internal/notification"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newNotifRouter(q *mock.Querier) http.Handler {
	svc := notification.NewService(q, nil, nil, nil)
	h := notification.NewHandler(svc)
	jwtSvc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc))
	r.With(middleware.RequireAuth).Get("/api/notifications", h.List)
	r.With(middleware.RequireAuth).Get("/api/notifications/unread-count", h.UnreadCount)
	r.With(middleware.RequireAuth).Post("/api/notifications/read", h.MarkRead)
	r.With(middleware.RequireAuth).Get("/api/notifications/prefs", h.GetPrefs)
	r.With(middleware.RequireAuth).Post("/api/notifications/prefs", h.UpdatePrefs)
	return r
}

func notifToken(t *testing.T) (string, uuid.UUID) {
	t.Helper()
	uid := uuid.New()
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, _ := svc.GenerateToken(uid, "alice", "USER")
	return tok, uid
}

func getAuth(t *testing.T, r http.Handler, path, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func postAuth(t *testing.T, r http.Handler, path, bearer string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// --- Unauthenticated access ---

func TestNotifications_RequireAuth(t *testing.T) {
	endpoints := []string{
		"/api/notifications",
		"/api/notifications/unread-count",
		"/api/notifications/prefs",
	}
	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			rr := getAuth(t, newNotifRouter(&mock.Querier{}), ep, "")
			if rr.Code != http.StatusUnauthorized {
				t.Errorf("status: got %d, want 401", rr.Code)
			}
		})
	}
}

// --- List ---

func TestNotificationList_Success(t *testing.T) {
	tok, uid := notifToken(t)
	q := &mock.Querier{
		ListNotificationsFn: func(_ context.Context, arg db.ListNotificationsParams) ([]db.Notification, error) {
			return []db.Notification{
				{ID: uuid.New(), Type: "COMMENT_REPLY", UserID: uid, Read: false},
				{ID: uuid.New(), Type: "USER_FOLLOWED", UserID: uid, Read: true},
			}, nil
		},
	}
	rr := getAuth(t, newNotifRouter(q), "/api/notifications", tok)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	var resp struct{ Data []db.Notification `json:"data"` }
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(resp.Data))
	}
}

func TestNotificationList_Empty(t *testing.T) {
	tok, _ := notifToken(t)
	rr := getAuth(t, newNotifRouter(&mock.Querier{}), "/api/notifications", tok)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestNotificationList_Pagination(t *testing.T) {
	tok, _ := notifToken(t)
	q := &mock.Querier{
		ListNotificationsFn: func(_ context.Context, arg db.ListNotificationsParams) ([]db.Notification, error) {
			if arg.Limit != 5 || arg.Offset != 10 {
				t.Errorf("pagination: got limit=%d offset=%d, want 5/10", arg.Limit, arg.Offset)
			}
			return nil, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/notifications?limit=5&offset=10", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	newNotifRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- UnreadCount ---

func TestNotificationUnreadCount_Success(t *testing.T) {
	tok, _ := notifToken(t)
	q := &mock.Querier{
		CountUnreadNotificationsFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
			return 7, nil
		},
	}
	rr := getAuth(t, newNotifRouter(q), "/api/notifications/unread-count", tok)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	var resp struct {
		Data struct{ Count int64 `json:"count"` } `json:"data"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Data.Count != 7 {
		t.Errorf("count: got %d, want 7", resp.Data.Count)
	}
}

func TestNotificationUnreadCount_Zero(t *testing.T) {
	tok, _ := notifToken(t)
	rr := getAuth(t, newNotifRouter(&mock.Querier{}), "/api/notifications/unread-count", tok)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- MarkRead ---

func TestNotificationMarkRead_Success(t *testing.T) {
	tok, _ := notifToken(t)
	rr := postAuth(t, newNotifRouter(&mock.Querier{}), "/api/notifications/read", tok, nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestNotificationMarkRead_Unauthenticated(t *testing.T) {
	rr := postAuth(t, newNotifRouter(&mock.Querier{}), "/api/notifications/read", "", nil)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

// --- GetPrefs ---

func TestNotificationGetPrefs_Success(t *testing.T) {
	tok, uid := notifToken(t)
	q := &mock.Querier{
		GetNotificationPreferencesFn: func(_ context.Context, _ uuid.UUID) ([]db.NotificationPreference, error) {
			return []db.NotificationPreference{
				{UserID: uid, Type: "COMMENT_REPLY", InAppEnabled: true, PushEnabled: false},
			}, nil
		},
	}
	rr := getAuth(t, newNotifRouter(q), "/api/notifications/prefs", tok)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- UpdatePrefs ---

func TestNotificationUpdatePrefs_Success(t *testing.T) {
	tok, _ := notifToken(t)
	rr := postAuth(t, newNotifRouter(&mock.Querier{}), "/api/notifications/prefs", tok, map[string]any{
		"type":             "COMMENT_REPLY",
		"in_app_enabled":   true,
		"push_enabled":     false,
	})
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestNotificationUpdatePrefs_MissingType(t *testing.T) {
	tok, _ := notifToken(t)
	rr := postAuth(t, newNotifRouter(&mock.Querier{}), "/api/notifications/prefs", tok, map[string]any{
		"in_app_enabled": true,
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}
