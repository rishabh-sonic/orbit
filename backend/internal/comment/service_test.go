package comment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/comment"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/db/mock"
)

func newCommentSvc(q *mock.Querier) *comment.Service {
	return comment.NewService(q)
}

// --- Create ---

func TestCommentCreate_Success(t *testing.T) {
	authorID := uuid.New()
	postID := uuid.New()
	commentID := uuid.New()
	incrementCalled := false

	q := &mock.Querier{
		CreateCommentFn: func(_ context.Context, arg db.CreateCommentParams) (db.Comment, error) {
			return db.Comment{ID: commentID, Content: arg.Content, AuthorID: arg.AuthorID, PostID: arg.PostID}, nil
		},
		IncrementPostCommentCountFn: func(_ context.Context, id uuid.UUID) error {
			incrementCalled = true
			if id != postID {
				t.Errorf("IncrementPostCommentCount: got %v, want %v", id, postID)
			}
			return nil
		},
	}
	c, err := newCommentSvc(q).Create(context.Background(), "Great post!", authorID, postID, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.Content != "Great post!" {
		t.Errorf("Content: got %q", c.Content)
	}
	if !incrementCalled {
		t.Error("expected IncrementPostCommentCount to be called")
	}
}

func TestCommentCreate_WithParent(t *testing.T) {
	parentID := uuid.New()
	q := &mock.Querier{
		CreateCommentFn: func(_ context.Context, arg db.CreateCommentParams) (db.Comment, error) {
			if !arg.ParentID.Valid || arg.ParentID.UUID != parentID {
				t.Errorf("ParentID not set correctly")
			}
			return db.Comment{ID: uuid.New(), Content: arg.Content}, nil
		},
	}
	_, err := newCommentSvc(q).Create(context.Background(), "Reply!", uuid.New(), uuid.New(), &parentID)
	if err != nil {
		t.Fatalf("Create with parent: %v", err)
	}
}

func TestCommentCreate_DBError(t *testing.T) {
	q := &mock.Querier{
		CreateCommentFn: func(_ context.Context, _ db.CreateCommentParams) (db.Comment, error) {
			return db.Comment{}, errors.New("db error")
		},
	}
	_, err := newCommentSvc(q).Create(context.Background(), "content", uuid.New(), uuid.New(), nil)
	if err == nil {
		t.Error("expected error from DB failure")
	}
}

// --- Delete ---

func TestCommentDelete_ByAuthor(t *testing.T) {
	authorID := uuid.New()
	commentID := uuid.New()
	postID := uuid.New()
	softDeleted := false
	decremented := false

	q := &mock.Querier{
		GetCommentByIDFn: func(_ context.Context, _ uuid.UUID) (db.Comment, error) {
			return db.Comment{ID: commentID, AuthorID: authorID, PostID: postID}, nil
		},
		SoftDeleteCommentFn: func(_ context.Context, _ uuid.UUID) error {
			softDeleted = true
			return nil
		},
		DecrementPostCommentCountFn: func(_ context.Context, id uuid.UUID) error {
			decremented = true
			if id != postID {
				t.Errorf("DecrementPostCommentCount: got %v, want %v", id, postID)
			}
			return nil
		},
	}
	err := newCommentSvc(q).Delete(context.Background(), commentID, authorID, false)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !softDeleted {
		t.Error("expected SoftDeleteComment to be called")
	}
	if !decremented {
		t.Error("expected DecrementPostCommentCount to be called")
	}
}

func TestCommentDelete_ByAdmin(t *testing.T) {
	authorID := uuid.New()
	adminID := uuid.New()
	commentID := uuid.New()
	softDeleted := false

	q := &mock.Querier{
		GetCommentByIDFn: func(_ context.Context, _ uuid.UUID) (db.Comment, error) {
			return db.Comment{ID: commentID, AuthorID: authorID}, nil
		},
		SoftDeleteCommentFn: func(_ context.Context, _ uuid.UUID) error {
			softDeleted = true
			return nil
		},
	}
	if err := newCommentSvc(q).Delete(context.Background(), commentID, adminID, true); err != nil {
		t.Fatalf("Delete by admin: %v", err)
	}
	if !softDeleted {
		t.Error("expected SoftDeleteComment to be called")
	}
}

func TestCommentDelete_ForbiddenForOtherUser(t *testing.T) {
	commentID := uuid.New()
	authorID := uuid.New()
	otherUser := uuid.New()
	q := &mock.Querier{
		GetCommentByIDFn: func(_ context.Context, _ uuid.UUID) (db.Comment, error) {
			return db.Comment{ID: commentID, AuthorID: authorID}, nil
		},
	}
	err := newCommentSvc(q).Delete(context.Background(), commentID, otherUser, false)
	if !errors.Is(err, comment.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestCommentDelete_NotFound(t *testing.T) {
	err := newCommentSvc(&mock.Querier{}).Delete(context.Background(), uuid.New(), uuid.New(), false)
	if !errors.Is(err, comment.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

// --- ListForPost ---

func TestCommentListForPost(t *testing.T) {
	postID := uuid.New()
	q := &mock.Querier{
		ListTopLevelCommentsFn: func(_ context.Context, arg db.ListTopLevelCommentsParams) ([]db.Comment, error) {
			if arg.PostID != postID {
				t.Errorf("PostID mismatch")
			}
			if arg.Limit != 10 || arg.Offset != 5 {
				t.Errorf("pagination params: got %d/%d", arg.Limit, arg.Offset)
			}
			return []db.Comment{
				{ID: uuid.New(), Content: "First comment", PostID: postID},
			}, nil
		},
	}
	comments, err := newCommentSvc(q).ListForPost(context.Background(), postID, 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(comments))
	}
}

// --- ListReplies ---

func TestCommentListReplies(t *testing.T) {
	parentID := uuid.New()
	q := &mock.Querier{
		ListRepliesFn: func(_ context.Context, arg db.ListRepliesParams) ([]db.Comment, error) {
			if !arg.ParentID.Valid || arg.ParentID.UUID != parentID {
				t.Errorf("ParentID not set correctly")
			}
			return []db.Comment{
				{ID: uuid.New(), Content: "Reply", ParentID: uuid.NullUUID{UUID: parentID, Valid: true}},
			}, nil
		},
	}
	replies, err := newCommentSvc(q).ListReplies(context.Background(), parentID, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(replies) != 1 {
		t.Errorf("expected 1 reply, got %d", len(replies))
	}
}
