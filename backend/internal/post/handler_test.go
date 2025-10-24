package post_test

import (
	"bytes"
	"context"
	"database/sql"
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
	"github.com/rishabh-sonic/orbit/internal/post"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newPostRouter(q *mock.Querier) http.Handler {
	svc := post.NewService(q)
	h := post.NewHandler(svc)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(
		token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour),
	))
	r.Get("/api/posts", h.List)
	r.Get("/api/posts/recent", h.Recent)
	r.Get("/api/posts/featured", h.Featured)
	r.With(middleware.RequireAuth).Post("/api/posts", h.Create)
	r.Get("/api/posts/{id}", h.GetByID)
	r.With(middleware.RequireAuth).Put("/api/posts/{id}", h.Update)
	r.With(middleware.RequireAuth).Delete("/api/posts/{id}", h.Delete)
	r.With(middleware.RequireAuth).Post("/api/posts/{id}/close", h.Close)
	r.With(middleware.RequireAuth).Post("/api/posts/{id}/reopen", h.Reopen)
	return r
}

func userToken(t *testing.T, userID uuid.UUID, role string) string {
	t.Helper()
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, err := svc.GenerateToken(userID, "alice", role)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func doPost(t *testing.T, r http.Handler, path, bearer string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func doGet(t *testing.T, r http.Handler, path, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func doPut(t *testing.T, r http.Handler, path, bearer string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPut, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func doDelete(t *testing.T, r http.Handler, path, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// --- Create ---

func TestCreatePost_Success(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	q := &mock.Querier{
		CreatePostFn: func(_ context.Context, arg db.CreatePostParams) (db.Post, error) {
			return db.Post{ID: postID, Title: arg.Title, Content: arg.Content, AuthorID: arg.AuthorID}, nil
		},
	}
	rr := doPost(t, newPostRouter(q), "/api/posts", userToken(t, authorID, "USER"), map[string]string{
		"title":   "Hello World",
		"content": "This is my first post.",
	})
	if rr.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201 — body: %s", rr.Code, rr.Body)
	}
}

func TestCreatePost_Unauthenticated(t *testing.T) {
	rr := doPost(t, newPostRouter(&mock.Querier{}), "/api/posts", "", map[string]string{
		"title":   "Test",
		"content": "Body",
	})
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestCreatePost_EmptyTitle(t *testing.T) {
	rr := doPost(t, newPostRouter(&mock.Querier{}), "/api/posts", userToken(t, uuid.New(), "USER"), map[string]string{
		"title":   "",
		"content": "body",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- GetByID ---

func TestGetPostByID_Success(t *testing.T) {
	postID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, id uuid.UUID) (db.Post, error) {
			return db.Post{
				ID:      id,
				Title:   "Test Post",
				Content: "Content here",
			}, nil
		},
	}
	rr := doGet(t, newPostRouter(q), "/api/posts/"+postID.String(), "")
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
}

func TestGetPostByID_NotFound(t *testing.T) {
	rr := doGet(t, newPostRouter(&mock.Querier{}), "/api/posts/"+uuid.New().String(), "")
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

func TestGetPostByID_InvalidID(t *testing.T) {
	rr := doGet(t, newPostRouter(&mock.Querier{}), "/api/posts/not-a-uuid", "")
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- Update ---

func TestUpdatePost_Success(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	newTitle := "Updated Title"
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID, Title: "Old Title"}, nil
		},
		UpdatePostFn: func(_ context.Context, arg db.UpdatePostParams) (db.Post, error) {
			return db.Post{ID: arg.ID, Title: newTitle, AuthorID: authorID}, nil
		},
	}
	rr := doPut(t, newPostRouter(q), "/api/posts/"+postID.String(), userToken(t, authorID, "USER"), map[string]*string{
		"title": &newTitle,
	})
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
}

func TestUpdatePost_Forbidden(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	otherUser := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	title := "New Title"
	rr := doPut(t, newPostRouter(q), "/api/posts/"+postID.String(), userToken(t, otherUser, "USER"), map[string]*string{
		"title": &title,
	})
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestUpdatePost_NotFound(t *testing.T) {
	title := "X"
	rr := doPut(t, newPostRouter(&mock.Querier{}), "/api/posts/"+uuid.New().String(), userToken(t, uuid.New(), "USER"), map[string]*string{
		"title": &title,
	})
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

// --- Delete ---

func TestDeletePost_Success(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	rr := doDelete(t, newPostRouter(q), "/api/posts/"+postID.String(), userToken(t, authorID, "USER"))
	if rr.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204 — body: %s", rr.Code, rr.Body)
	}
}

func TestDeletePost_ForbiddenForOtherUser(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	otherUser := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	rr := doDelete(t, newPostRouter(q), "/api/posts/"+postID.String(), userToken(t, otherUser, "USER"))
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestDeletePost_AdminCanDeleteAnyPost(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	adminID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	rr := doDelete(t, newPostRouter(q), "/api/posts/"+postID.String(), userToken(t, adminID, "ADMIN"))
	if rr.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204 — body: %s", rr.Code, rr.Body)
	}
}

// --- List ---

func TestListPosts_Success(t *testing.T) {
	q := &mock.Querier{
		ListPostsFn: func(_ context.Context, _ db.ListPostsParams) ([]db.Post, error) {
			return []db.Post{
				{ID: uuid.New(), Title: "Post 1"},
				{ID: uuid.New(), Title: "Post 2"},
			}, nil
		},
	}
	rr := doGet(t, newPostRouter(q), "/api/posts", "")
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestListPosts_Empty(t *testing.T) {
	rr := doGet(t, newPostRouter(&mock.Querier{}), "/api/posts", "")
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- Close / Reopen ---

func TestClosePost_Success(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	rr := doPost(t, newPostRouter(q), "/api/posts/"+postID.String()+"/close", userToken(t, authorID, "USER"), nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
}

func TestReopenPost_Success(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{
				ID:       postID,
				AuthorID: authorID,
				Closed:   true,
				DeletedAt: sql.NullTime{},
			}, nil
		},
	}
	rr := doPost(t, newPostRouter(q), "/api/posts/"+postID.String()+"/reopen", userToken(t, authorID, "USER"), nil)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
}

// --- Featured / Recent ---

func TestFeaturedPosts(t *testing.T) {
	q := &mock.Querier{
		ListFeaturedPostsFn: func(_ context.Context, _ db.ListFeaturedPostsParams) ([]db.Post, error) {
			return []db.Post{{ID: uuid.New(), Title: "Featured"}}, nil
		},
	}
	rr := doGet(t, newPostRouter(q), "/api/posts/featured", "")
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}
