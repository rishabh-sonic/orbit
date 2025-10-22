package user_test

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
	"github.com/redis/go-redis/v9"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/db/mock"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/internal/user"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newUserRouter(q *mock.Querier) http.Handler {
	// nil redis client — tests don't exercise Redis-dependent code paths
	svc := user.NewService(q, (*redis.Client)(nil))
	h := user.NewHandler(svc)
	jwtSvc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc))

	r.With(middleware.RequireAuth).Get("/api/users/me", h.GetMe)
	r.With(middleware.RequireAuth).Put("/api/users/me", h.UpdateMe)
	r.Get("/api/users/{identifier}", h.GetUser)
	r.Get("/api/users/{identifier}/followers", h.GetFollowers)
	r.Get("/api/users/{identifier}/following", h.GetFollowing)
	r.With(middleware.RequireAuth).Post("/api/subscriptions/users/{username}", h.Follow)
	r.With(middleware.RequireAuth).Delete("/api/subscriptions/users/{username}", h.Unfollow)
	return r
}

func userTok(t *testing.T, userID uuid.UUID, username string) string {
	t.Helper()
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, _ := svc.GenerateToken(userID, username, "USER")
	return tok
}

// --- GetMe ---

func TestGetMe_Success(t *testing.T) {
	userID := uuid.New()
	q := &mock.Querier{
		GetUserByIDFn: func(_ context.Context, id uuid.UUID) (db.User, error) {
			return db.User{ID: id, Username: "alice", Email: "alice@example.com", Role: db.UserRoleUSER}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+userTok(t, userID, "alice"))
	rr := httptest.NewRecorder()
	newUserRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
}

func TestGetMe_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	rr := httptest.NewRecorder()
	newUserRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestGetMe_UserDeleted(t *testing.T) {
	userID := uuid.New()
	// DB returns not found even though token is valid (edge case)
	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+userTok(t, userID, "alice"))
	rr := httptest.NewRecorder()
	newUserRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

// --- UpdateMe ---

func TestUpdateMe_Success(t *testing.T) {
	userID := uuid.New()
	bio := "Hello, world"
	q := &mock.Querier{
		UpdateUserFn: func(_ context.Context, arg db.UpdateUserParams) (db.User, error) {
			return db.User{ID: arg.ID, Username: "alice"}, nil
		},
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(map[string]*string{"introduction": &bio})
	req := httptest.NewRequest(http.MethodPut, "/api/users/me", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userTok(t, userID, "alice"))
	rr := httptest.NewRecorder()
	newUserRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
}

// --- GetUser ---

func TestGetUser_ByUsername_Success(t *testing.T) {
	q := &mock.Querier{
		GetUserByUsernameFn: func(_ context.Context, username string) (db.User, error) {
			return db.User{ID: uuid.New(), Username: username, Role: db.UserRoleUSER}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/users/alice", nil)
	rr := httptest.NewRecorder()
	newUserRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/nobody", nil)
	rr := httptest.NewRecorder()
	newUserRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

// --- GetFollowers / GetFollowing ---

func TestGetFollowers_Success(t *testing.T) {
	q := &mock.Querier{
		GetUserByUsernameFn: func(_ context.Context, u string) (db.User, error) {
			return db.User{ID: uuid.New(), Username: u}, nil
		},
		GetFollowersFn: func(_ context.Context, _ db.GetFollowersParams) ([]db.User, error) {
			return []db.User{{ID: uuid.New(), Username: "bob"}}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/users/alice/followers", nil)
	rr := httptest.NewRecorder()
	newUserRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestGetFollowing_Success(t *testing.T) {
	q := &mock.Querier{
		GetUserByUsernameFn: func(_ context.Context, u string) (db.User, error) {
			return db.User{ID: uuid.New(), Username: u}, nil
		},
		GetFollowingFn: func(_ context.Context, _ db.GetFollowingParams) ([]db.User, error) {
			return []db.User{{ID: uuid.New(), Username: "carol"}}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/users/alice/following", nil)
	rr := httptest.NewRecorder()
	newUserRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- Follow / Unfollow ---

func TestFollow_Success(t *testing.T) {
	followerID := uuid.New()
	targetID := uuid.New()
	q := &mock.Querier{
		GetUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: targetID, Username: "bob"}, nil
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions/users/bob", nil)
	req.Header.Set("Authorization", "Bearer "+userTok(t, followerID, "alice"))
	rr := httptest.NewRecorder()
	newUserRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
}

func TestFollow_TargetNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions/users/nobody", nil)
	req.Header.Set("Authorization", "Bearer "+userTok(t, uuid.New(), "alice"))
	rr := httptest.NewRecorder()
	newUserRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

func TestUnfollow_Success(t *testing.T) {
	followerID := uuid.New()
	q := &mock.Querier{
		GetUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: uuid.New(), Username: "bob"}, nil
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/subscriptions/users/bob", nil)
	req.Header.Set("Authorization", "Bearer "+userTok(t, followerID, "alice"))
	rr := httptest.NewRecorder()
	newUserRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204", rr.Code)
	}
}

func TestFollow_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions/users/bob", nil)
	rr := httptest.NewRecorder()
	newUserRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}
