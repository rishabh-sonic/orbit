package post_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/db/mock"
	"github.com/rishabh-sonic/orbit/internal/post"
)

func newPostSvc(q *mock.Querier) *post.Service {
	return post.NewService(q)
}

// --- Create ---

func TestPostCreate_Success(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	q := &mock.Querier{
		CreatePostFn: func(_ context.Context, arg db.CreatePostParams) (db.Post, error) {
			return db.Post{ID: postID, Title: arg.Title, Content: arg.Content, AuthorID: arg.AuthorID}, nil
		},
	}
	p, err := newPostSvc(q).Create(context.Background(), post.CreateInput{
		Title:    "Hello World",
		Content:  "Post body",
		AuthorID: authorID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.Title != "Hello World" {
		t.Errorf("Title: got %q, want 'Hello World'", p.Title)
	}
	if p.AuthorID != authorID {
		t.Errorf("AuthorID mismatch")
	}
}

func TestPostCreate_EmptyTitle(t *testing.T) {
	_, err := newPostSvc(&mock.Querier{}).Create(context.Background(), post.CreateInput{
		Title:    "",
		Content:  "body",
		AuthorID: uuid.New(),
	})
	if err == nil {
		t.Error("expected error for empty title")
	}
}

func TestPostCreate_EmptyContent(t *testing.T) {
	_, err := newPostSvc(&mock.Querier{}).Create(context.Background(), post.CreateInput{
		Title:    "Title",
		Content:  "",
		AuthorID: uuid.New(),
	})
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestPostCreate_DBError(t *testing.T) {
	q := &mock.Querier{
		CreatePostFn: func(_ context.Context, _ db.CreatePostParams) (db.Post, error) {
			return db.Post{}, errors.New("db error")
		},
	}
	_, err := newPostSvc(q).Create(context.Background(), post.CreateInput{
		Title:    "Title",
		Content:  "Body",
		AuthorID: uuid.New(),
	})
	if err == nil {
		t.Error("expected error from DB failure")
	}
}

// --- GetByID ---

func TestPostGetByID_Success(t *testing.T) {
	postID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, id uuid.UUID) (db.Post, error) {
			return db.Post{ID: id, Title: "My Post"}, nil
		},
	}
	p, err := newPostSvc(q).GetByID(context.Background(), postID, nil)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if p.ID != postID {
		t.Errorf("ID mismatch")
	}
}

func TestPostGetByID_NotFound(t *testing.T) {
	_, err := newPostSvc(&mock.Querier{}).GetByID(context.Background(), uuid.New(), nil)
	if !errors.Is(err, post.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestPostGetByID_RecordsViewAndRead(t *testing.T) {
	viewsIncremented := false
	readRecorded := false
	postID := uuid.New()
	viewerID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, id uuid.UUID) (db.Post, error) {
			return db.Post{ID: id}, nil
		},
		IncrementPostViewsFn: func(_ context.Context, id uuid.UUID) error {
			viewsIncremented = true
			return nil
		},
		RecordPostReadFn: func(_ context.Context, arg db.RecordPostReadParams) error {
			readRecorded = true
			if arg.PostID != postID || arg.UserID != viewerID {
				t.Errorf("RecordPostRead: wrong IDs")
			}
			return nil
		},
	}
	newPostSvc(q).GetByID(context.Background(), postID, &viewerID)
	if !viewsIncremented {
		t.Error("expected IncrementPostViews to be called")
	}
	if !readRecorded {
		t.Error("expected RecordPostRead to be called for authenticated viewer")
	}
}

func TestPostGetByID_NoReadForAnonymous(t *testing.T) {
	readRecorded := false
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, id uuid.UUID) (db.Post, error) {
			return db.Post{ID: id}, nil
		},
		RecordPostReadFn: func(_ context.Context, _ db.RecordPostReadParams) error {
			readRecorded = true
			return nil
		},
	}
	newPostSvc(q).GetByID(context.Background(), uuid.New(), nil)
	if readRecorded {
		t.Error("should NOT record post read for anonymous viewer")
	}
}

// --- Update ---

func TestPostUpdate_Success(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	newTitle := "Updated Title"
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID, Title: "Old Title"}, nil
		},
		UpdatePostFn: func(_ context.Context, arg db.UpdatePostParams) (db.Post, error) {
			return db.Post{ID: arg.ID, Title: newTitle}, nil
		},
	}
	p, err := newPostSvc(q).Update(context.Background(), postID, authorID, post.UpdateInput{Title: &newTitle})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if p.Title != newTitle {
		t.Errorf("Title: got %q, want %q", p.Title, newTitle)
	}
}

func TestPostUpdate_NotFound(t *testing.T) {
	_, err := newPostSvc(&mock.Querier{}).Update(context.Background(), uuid.New(), uuid.New(), post.UpdateInput{})
	if !errors.Is(err, post.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestPostUpdate_ForbiddenForOtherUser(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	otherUser := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	_, err := newPostSvc(q).Update(context.Background(), postID, otherUser, post.UpdateInput{})
	if !errors.Is(err, post.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

// --- Delete ---

func TestPostDelete_ByAuthor(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	deleted := false
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
		SoftDeletePostFn: func(_ context.Context, _ uuid.UUID) error {
			deleted = true
			return nil
		},
	}
	err := newPostSvc(q).Delete(context.Background(), postID, authorID, false)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Error("expected SoftDeletePost to be called")
	}
}

func TestPostDelete_ByAdmin(t *testing.T) {
	authorID := uuid.New()
	adminID := uuid.New()
	postID := uuid.New()
	deleted := false
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
		SoftDeletePostFn: func(_ context.Context, _ uuid.UUID) error {
			deleted = true
			return nil
		},
	}
	err := newPostSvc(q).Delete(context.Background(), postID, adminID, true)
	if err != nil {
		t.Fatalf("Delete by admin: %v", err)
	}
	if !deleted {
		t.Error("expected SoftDeletePost to be called for admin")
	}
}

func TestPostDelete_ForbiddenForOtherUser(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	otherUser := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	err := newPostSvc(q).Delete(context.Background(), postID, otherUser, false)
	if !errors.Is(err, post.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestPostDelete_NotFound(t *testing.T) {
	err := newPostSvc(&mock.Querier{}).Delete(context.Background(), uuid.New(), uuid.New(), false)
	if !errors.Is(err, post.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

// --- SetClosed ---

func TestSetClosed_ByAuthor(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	if err := newPostSvc(q).SetClosed(context.Background(), postID, authorID, true); err != nil {
		t.Fatalf("SetClosed: %v", err)
	}
}

func TestSetClosed_ForbiddenForOtherUser(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	q := &mock.Querier{
		GetPostByIDFn: func(_ context.Context, _ uuid.UUID) (db.Post, error) {
			return db.Post{ID: postID, AuthorID: authorID}, nil
		},
	}
	err := newPostSvc(q).SetClosed(context.Background(), postID, uuid.New(), true)
	if !errors.Is(err, post.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

// --- List variants ---

func TestPostList_ReturnsAll(t *testing.T) {
	q := &mock.Querier{
		ListPostsFn: func(_ context.Context, arg db.ListPostsParams) ([]db.Post, error) {
			if arg.Limit != 10 || arg.Offset != 5 {
				t.Errorf("pagination params: got %d/%d", arg.Limit, arg.Offset)
			}
			return []db.Post{{ID: uuid.New(), Title: "A"}, {ID: uuid.New(), Title: "B"}}, nil
		},
	}
	posts, err := newPostSvc(q).List(context.Background(), 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}
}

// Suppress unused import warnings — sql is used transitively
var _ = sql.NullString{}
