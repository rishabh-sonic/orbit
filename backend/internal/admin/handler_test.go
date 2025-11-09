package admin_test

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
	"github.com/rishabh-sonic/orbit/internal/admin"
	"github.com/rishabh-sonic/orbit/internal/comment"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/db/mock"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/internal/post"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newAdminRouter(q *mock.Querier) http.Handler {
	postSvc := post.NewService(q)
	commentSvc := comment.NewService(q)
	h := admin.NewHandler(q, postSvc, commentSvc)
	jwtSvc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc))
	r.With(middleware.RequireAdmin).Get("/api/admin/users", h.ListUsers)
	r.With(middleware.RequireAdmin).Post("/api/admin/users/{id}/ban", h.BanUser)
	r.With(middleware.RequireAdmin).Post("/api/admin/users/{id}/unban", h.UnbanUser)
	r.With(middleware.RequireAdmin).Delete("/api/admin/posts/{id}", h.DeletePost)
	r.With(middleware.RequireAdmin).Post("/api/admin/posts/{id}/pin", h.PinPost)
	r.With(middleware.RequireAdmin).Post("/api/admin/posts/{id}/unpin", h.UnpinPost)
	r.With(middleware.RequireAdmin).Get("/api/admin/config", h.GetConfig)
	r.With(middleware.RequireAdmin).Post("/api/admin/config", h.UpdateConfig)
	r.With(middleware.RequireAdmin).Get("/api/admin/stats/dau", h.StatDAU)
	r.With(middleware.RequireAdmin).Get("/api/admin/stats/new-users-range", h.StatNewUsers)
	r.With(middleware.RequireAdmin).Get("/api/admin/stats/posts-range", h.StatPosts)
	r.With(middleware.RequireAdmin).Get("/api/admin/stats/dau-range", h.StatDAURange)
	return r
}

func adminToken(t *testing.T) string {
	t.Helper()
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, _ := svc.GenerateToken(uuid.New(), "admin", "ADMIN")
	return tok
}

func userToken(t *testing.T) string {
	t.Helper()
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, _ := svc.GenerateToken(uuid.New(), "user", "USER")
	return tok
}

// --- Access control ---

func TestAdminEndpoints_RequireAdminRole(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/users"},
		{http.MethodPost, "/api/admin/users/" + uuid.New().String() + "/ban"},
		{http.MethodDelete, "/api/admin/posts/" + uuid.New().String()},
		{http.MethodGet, "/api/admin/config"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Authorization", "Bearer "+userToken(t))
			rr := httptest.NewRecorder()
			newAdminRouter(&mock.Querier{}).ServeHTTP(rr, req)
			if rr.Code != http.StatusForbidden {
				t.Errorf("status: got %d, want 403", rr.Code)
			}
		})
	}
}

func TestAdminEndpoints_RequireAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	rr := httptest.NewRecorder()
	newAdminRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

// --- ListUsers ---

func TestAdminListUsers_Success(t *testing.T) {
	q := &mock.Querier{
		ListUsersFn: func(_ context.Context, _ db.ListUsersParams) ([]db.User, error) {
			return []db.User{
				{ID: uuid.New(), Username: "alice", Role: db.UserRoleUSER},
				{ID: uuid.New(), Username: "bob", Role: db.UserRoleUSER},
			}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- BanUser ---

func TestAdminBanUser_Success(t *testing.T) {
	targetID := uuid.New()
	q := &mock.Querier{
		GetUserByIDFn: func(_ context.Context, _ uuid.UUID) (db.User, error) {
			return db.User{ID: targetID, Username: "alice", Role: db.UserRoleUSER}, nil
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+targetID.String()+"/ban", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
}

func TestAdminBanUser_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/not-a-uuid/ban", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- UnbanUser ---

func TestAdminUnbanUser_Success(t *testing.T) {
	targetID := uuid.New()
	q := &mock.Querier{
		GetUserByIDFn: func(_ context.Context, _ uuid.UUID) (db.User, error) {
			return db.User{ID: targetID, Username: "alice", Role: db.UserRoleUSER, Banned: true}, nil
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+targetID.String()+"/unban", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- DeletePost ---

func TestAdminDeletePost_Success(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/posts/"+postID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204 — body: %s", rr.Code, rr.Body)
	}
}

// --- PinPost ---

func TestAdminPinPost_Success(t *testing.T) {
	postID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/posts/"+postID.String()+"/pin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
}

// --- GetConfig ---

func TestAdminGetConfig_Success(t *testing.T) {
	q := &mock.Querier{
		ListConfigFn: func(_ context.Context) ([]db.SiteConfig, error) {
			return []db.SiteConfig{
				{Key: "site_name", Value: "Orbit"},
			}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- UpdateConfig ---

func TestAdminUpdateConfig_Success(t *testing.T) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(map[string]string{"site_name": "My Orbit"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/config", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- Stats ---

func TestAdminStatDAU_Success(t *testing.T) {
	q := &mock.Querier{
		CountDailyActiveUsersFn: func(_ context.Context, _ time.Time) (int64, error) {
			return 42, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats/dau", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	var resp struct {
		Data struct{ Dau int64 `json:"dau"` } `json:"data"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Data.Dau != 42 {
		t.Errorf("dau: got %d, want 42", resp.Data.Dau)
	}
}

func TestAdminStatDAU_RequiresAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats/dau", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestAdminStatNewUsers_Success(t *testing.T) {
	q := &mock.Querier{
		CountNewUsersInRangeFn: func(_ context.Context, _ db.CountNewUsersInRangeParams) (int64, error) {
			return 15, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats/new-users-range?from=2025-01-01&to=2025-01-31", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestAdminStatNewUsers_MissingDateRange(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats/new-users-range", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

func TestAdminStatPosts_Success(t *testing.T) {
	q := &mock.Querier{
		CountPostsInRangeFn: func(_ context.Context, _ db.CountPostsInRangeParams) (int64, error) {
			return 8, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats/posts-range?from=2025-01-01&to=2025-01-31", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestAdminStatDAURange_Success(t *testing.T) {
	q := &mock.Querier{
		CountDAUInRangeFn: func(_ context.Context, _ db.CountDAUInRangeParams) ([]db.CountDAUInRangeRow, error) {
			return []db.CountDAUInRangeRow{}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats/dau-range?from=2025-01-01&to=2025-01-31", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rr := httptest.NewRecorder()
	newAdminRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}
