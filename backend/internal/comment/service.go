package comment

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
)

var (
	ErrNotFound  = errors.New("comment not found")
	ErrForbidden = errors.New("forbidden")
)

// CommentAuthor is the author sub-object the frontend expects on every comment.
type CommentAuthor struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Avatar   *string   `json:"avatar"`
}

// CommentResponse is the JSON shape the frontend expects for every comment.
type CommentResponse struct {
	ID        uuid.UUID         `json:"id"`
	Content   string            `json:"content"`
	PinnedAt  *time.Time        `json:"pinned_at"`
	CreatedAt time.Time         `json:"created_at"`
	Author    CommentAuthor     `json:"author"`
	Replies   []CommentResponse `json:"replies"`
}

type Service struct {
	q db.Querier
}

func NewService(q db.Querier) *Service {
	return &Service{q: q}
}

// enrichComment builds a CommentResponse, fetching author info and nested
// replies recursively up to maxDepth levels.
func (s *Service) enrichComment(ctx context.Context, c db.Comment, depth int) CommentResponse {
	author, err := s.q.GetUserByID(ctx, c.AuthorID)
	if err != nil {
		author = db.User{ID: c.AuthorID, Username: "deleted"}
	}

	var avatar *string
	if author.Avatar.Valid {
		v := author.Avatar.String
		avatar = &v
	}
	var pinnedAt *time.Time
	if c.PinnedAt.Valid {
		t := c.PinnedAt.Time
		pinnedAt = &t
	}

	resp := CommentResponse{
		ID:        c.ID,
		Content:   c.Content,
		PinnedAt:  pinnedAt,
		CreatedAt: c.CreatedAt,
		Author: CommentAuthor{
			ID:       author.ID,
			Username: author.Username,
			Avatar:   avatar,
		},
		Replies: []CommentResponse{},
	}

	// Fetch nested replies up to 3 levels deep (matching frontend depth limit).
	if depth < 3 {
		rawReplies, err := s.q.ListReplies(ctx, db.ListRepliesParams{
			ParentID: uuid.NullUUID{UUID: c.ID, Valid: true},
			Limit:    100,
			Offset:   0,
		})
		if err == nil {
			for _, r := range rawReplies {
				resp.Replies = append(resp.Replies, s.enrichComment(ctx, r, depth+1))
			}
		}
	}

	return resp
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (s *Service) Create(ctx context.Context, content string, authorID, postID uuid.UUID, parentID *uuid.UUID) (CommentResponse, error) {
	var pid uuid.NullUUID
	if parentID != nil {
		pid = uuid.NullUUID{UUID: *parentID, Valid: true}
	}
	c, err := s.q.CreateComment(ctx, db.CreateCommentParams{
		Content:  content,
		AuthorID: authorID,
		PostID:   postID,
		ParentID: pid,
	})
	if err != nil {
		return CommentResponse{}, err
	}
	_ = s.q.IncrementPostCommentCount(ctx, postID)
	return s.enrichComment(ctx, c, 0), nil
}

func (s *Service) Delete(ctx context.Context, id, requestorID uuid.UUID, isAdmin bool) error {
	c, err := s.q.GetCommentByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if !isAdmin && c.AuthorID != requestorID {
		return ErrForbidden
	}
	if err := s.q.SoftDeleteComment(ctx, id); err != nil {
		return err
	}
	_ = s.q.DecrementPostCommentCount(ctx, c.PostID)
	return nil
}

// ── List queries ─────────────────────────────────────────────────────────────

func (s *Service) ListForPost(ctx context.Context, postID uuid.UUID, limit, offset int) ([]CommentResponse, error) {
	raw, err := s.q.ListTopLevelComments(ctx, db.ListTopLevelCommentsParams{
		PostID: postID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	result := make([]CommentResponse, 0, len(raw))
	for _, c := range raw {
		result = append(result, s.enrichComment(ctx, c, 0))
	}
	return result, nil
}

func (s *Service) ListReplies(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]CommentResponse, error) {
	raw, err := s.q.ListReplies(ctx, db.ListRepliesParams{
		ParentID: uuid.NullUUID{UUID: parentID, Valid: true},
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}
	result := make([]CommentResponse, 0, len(raw))
	for _, c := range raw {
		result = append(result, s.enrichComment(ctx, c, 1))
	}
	return result, nil
}

func (s *Service) Pin(ctx context.Context, id uuid.UUID) error {
	return s.q.PinComment(ctx, id)
}

func (s *Service) Unpin(ctx context.Context, id uuid.UUID) error {
	return s.q.UnpinComment(ctx, id)
}
