package comment_test

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
	"github.com/rishabh-sonic/orbit/internal/comment"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/db/mock"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newCommentRouter(q *mock.Querier) http.Handler {
	svc := comment.NewService(q)
	h := comment.NewHandler(svc)
	jwtSvc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(jwtSvc))
	r.With(middleware.RequireAuth).Post("/api/posts/{postId}/comments", h.Create)
	r.With(middleware.RequireAuth).Post("/api/comments/{id}/replies", h.Reply)
	r.Get("/api/posts/{postId}/comments", h.ListForPost)
	r.With(middleware.RequireAuth).Delete("/api/comments/{id}", h.Delete)
	return r
}

func commentToken(t *testing.T, userID uuid.UUID, role string) string {
	t.Helper()
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, _ := svc.GenerateToken(userID, "alice", role)
	return tok
}

func postComment(t *testing.T, r http.Handler, path, bearer string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// --- Create comment ---

func TestCreateComment_Success(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	q := &mock.Querier{
		CreateCommentFn: func(_ context.Context, arg db.CreateCommentParams) (db.Comment, error) {
			return db.Comment{ID: uuid.New(), Content: arg.Content, AuthorID: arg.AuthorID, PostID: arg.PostID}, nil
		},
	}
	rr := postComment(t, newCommentRouter(q),
		"/api/posts/"+postID.String()+"/comments",
		commentToken(t, authorID, "USER"),
		map[string]string{"content": "Great post!"},
	)
	if rr.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201 — body: %s", rr.Code, rr.Body)
	}
}

func TestCreateComment_Unauthenticated(t *testing.T) {
	rr := postComment(t, newCommentRouter(&mock.Querier{}),
		"/api/posts/"+uuid.New().String()+"/comments",
		"",
		map[string]string{"content": "hi"},
	)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestCreateComment_EmptyContent(t *testing.T) {
	rr := postComment(t, newCommentRouter(&mock.Querier{}),
		"/api/posts/"+uuid.New().String()+"/comments",
		commentToken(t, uuid.New(), "USER"),
		map[string]string{"content": ""},
	)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

func TestCreateComment_InvalidPostID(t *testing.T) {
	rr := postComment(t, newCommentRouter(&mock.Querier{}),
		"/api/posts/not-a-uuid/comments",
		commentToken(t, uuid.New(), "USER"),
		map[string]string{"content": "hello"},
	)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- Reply ---

func TestReply_ParentNotFound(t *testing.T) {
	rr := postComment(t, newCommentRouter(&mock.Querier{}),
		"/api/comments/"+uuid.New().String()+"/replies",
		commentToken(t, uuid.New(), "USER"),
		map[string]string{"content": "reply content"},
	)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

// --- List comments ---

func TestListComments_Success(t *testing.T) {
	postID := uuid.New()
	q := &mock.Querier{
		ListTopLevelCommentsFn: func(_ context.Context, _ db.ListTopLevelCommentsParams) ([]db.Comment, error) {
			return []db.Comment{
				{ID: uuid.New(), Content: "First comment"},
				{ID: uuid.New(), Content: "Second comment"},
			}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String()+"/comments", nil)
	rr := httptest.NewRecorder()
	newCommentRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestListComments_InvalidPostID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/posts/bad-id/comments", nil)
	rr := httptest.NewRecorder()
	newCommentRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- Delete ---

func TestDeleteComment_Success(t *testing.T) {
	authorID := uuid.New()
	commentID := uuid.New()
	q := &mock.Querier{
		GetCommentByIDFn: func(_ context.Context, _ uuid.UUID) (db.Comment, error) {
			return db.Comment{ID: commentID, AuthorID: authorID, PostID: uuid.New()}, nil
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/comments/"+commentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+commentToken(t, authorID, "USER"))
	rr := httptest.NewRecorder()
	newCommentRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204 — body: %s", rr.Code, rr.Body)
	}
}

func TestDeleteComment_Forbidden(t *testing.T) {
	commentID := uuid.New()
	authorID := uuid.New()
	otherUser := uuid.New()
	q := &mock.Querier{
		GetCommentByIDFn: func(_ context.Context, _ uuid.UUID) (db.Comment, error) {
			return db.Comment{ID: commentID, AuthorID: authorID}, nil
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/comments/"+commentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+commentToken(t, otherUser, "USER"))
	rr := httptest.NewRecorder()
	newCommentRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestDeleteComment_AdminCanDelete(t *testing.T) {
	commentID := uuid.New()
	authorID := uuid.New()
	adminID := uuid.New()
	q := &mock.Querier{
		GetCommentByIDFn: func(_ context.Context, _ uuid.UUID) (db.Comment, error) {
			return db.Comment{ID: commentID, AuthorID: authorID}, nil
		},
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/comments/"+commentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+commentToken(t, adminID, "ADMIN"))
	rr := httptest.NewRecorder()
	newCommentRouter(q).ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204", rr.Code)
	}
}

func TestDeleteComment_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/comments/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	newCommentRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}
