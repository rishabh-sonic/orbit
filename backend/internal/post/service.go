package post

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
)

// bgCtx returns a detached background context for fire-and-forget operations
// like search indexing that must outlive the HTTP request context.
func bgCtx() context.Context { return context.Background() }

var (
	ErrNotFound   = errors.New("post not found")
	ErrForbidden  = errors.New("forbidden")
	ErrPostClosed = errors.New("post is closed")
)

// PostIndexer is implemented by the search service; optional — nil disables indexing.
type PostIndexer interface {
	IndexPost(ctx context.Context, post db.Post)
	DeletePost(ctx context.Context, postID string)
}

// AuthorInfo is the author sub-object the frontend expects on every post.
type AuthorInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Avatar   *string   `json:"avatar"`
}

// PostResponse is the JSON shape the frontend expects for every post.
type PostResponse struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Status       string     `json:"status"` // "open" | "closed"
	PinnedAt     *time.Time `json:"pinned_at"`
	Views        int32      `json:"views"`
	CommentCount int32      `json:"comment_count"`
	CreatedAt    time.Time  `json:"created_at"`
	Author       AuthorInfo `json:"author"`
}

type Service struct {
	q       db.Querier
	indexer PostIndexer // optional
}

func NewService(q db.Querier, indexer PostIndexer) *Service {
	return &Service{q: q, indexer: indexer}
}

// toPostResponse converts a raw db.Post + its author into the API response shape.
func toPostResponse(p db.Post, author db.User) PostResponse {
	status := "open"
	if p.Closed {
		status = "closed"
	}
	var pinnedAt *time.Time
	if p.PinnedAt.Valid {
		t := p.PinnedAt.Time
		pinnedAt = &t
	}
	var avatar *string
	if author.Avatar.Valid {
		s := author.Avatar.String
		avatar = &s
	}
	return PostResponse{
		ID:           p.ID,
		Title:        p.Title,
		Content:      p.Content,
		Status:       status,
		PinnedAt:     pinnedAt,
		Views:        p.Views,
		CommentCount: p.CommentCount,
		CreatedAt:    p.CreatedAt,
		Author: AuthorInfo{
			ID:       author.ID,
			Username: author.Username,
			Avatar:   avatar,
		},
	}
}

// enrichPost fetches the author for a single post and builds a PostResponse.
func (s *Service) enrichPost(ctx context.Context, p db.Post) (PostResponse, error) {
	author, err := s.q.GetUserByID(ctx, p.AuthorID)
	if err != nil {
		// Return partial data rather than a hard error so a deleted author doesn't
		// break the whole feed.
		author = db.User{ID: p.AuthorID, Username: "deleted"}
	}
	return toPostResponse(p, author), nil
}

// enrichPosts batch-fetches authors (deduped) and builds PostResponse slices.
func (s *Service) enrichPosts(ctx context.Context, posts []db.Post) ([]PostResponse, error) {
	// Collect unique author IDs
	seen := make(map[uuid.UUID]bool)
	for _, p := range posts {
		seen[p.AuthorID] = true
	}
	authors := make(map[uuid.UUID]db.User, len(seen))
	for id := range seen {
		if u, err := s.q.GetUserByID(ctx, id); err == nil {
			authors[id] = u
		}
	}
	result := make([]PostResponse, 0, len(posts))
	for _, p := range posts {
		author := authors[p.AuthorID]
		result = append(result, toPostResponse(p, author))
	}
	return result, nil
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

type CreateInput struct {
	Title    string
	Content  string
	AuthorID uuid.UUID
}

func (s *Service) Create(ctx context.Context, in CreateInput) (PostResponse, error) {
	if in.Title == "" || in.Content == "" {
		return PostResponse{}, errors.New("title and content are required")
	}
	p, err := s.q.CreatePost(ctx, db.CreatePostParams{
		Title:    in.Title,
		Content:  in.Content,
		AuthorID: in.AuthorID,
	})
	if err != nil {
		return PostResponse{}, err
	}
	if s.indexer != nil {
		go s.indexer.IndexPost(bgCtx(), p)
	}
	return s.enrichPost(ctx, p)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (PostResponse, error) {
	p, err := s.q.GetPostByID(ctx, id)
	if err != nil {
		return PostResponse{}, ErrNotFound
	}
	_ = s.q.IncrementPostViews(ctx, id)
	if viewerID != nil {
		_ = s.q.RecordPostRead(ctx, db.RecordPostReadParams{PostID: id, UserID: *viewerID})
	}
	return s.enrichPost(ctx, p)
}

type UpdateInput struct {
	Title   *string
	Content *string
}

func (s *Service) Update(ctx context.Context, id, requestorID uuid.UUID, in UpdateInput) (PostResponse, error) {
	p, err := s.q.GetPostByID(ctx, id)
	if err != nil {
		return PostResponse{}, ErrNotFound
	}
	if p.AuthorID != requestorID {
		return PostResponse{}, ErrForbidden
	}
	params := db.UpdatePostParams{ID: id}
	if in.Title != nil {
		params.Title = nullableStr(*in.Title)
	}
	if in.Content != nil {
		params.Content = nullableStr(*in.Content)
	}
	updated, err := s.q.UpdatePost(ctx, params)
	if err != nil {
		return PostResponse{}, err
	}
	if s.indexer != nil {
		go s.indexer.IndexPost(bgCtx(), updated)
	}
	return s.enrichPost(ctx, updated)
}

func (s *Service) Delete(ctx context.Context, id, requestorID uuid.UUID, isAdmin bool) error {
	p, err := s.q.GetPostByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if !isAdmin && p.AuthorID != requestorID {
		return ErrForbidden
	}
	if err := s.q.SoftDeletePost(ctx, id); err != nil {
		return err
	}
	if s.indexer != nil {
		go s.indexer.DeletePost(bgCtx(), id.String())
	}
	return nil
}

func (s *Service) SetClosed(ctx context.Context, id, requestorID uuid.UUID, closed bool) error {
	p, err := s.q.GetPostByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if p.AuthorID != requestorID {
		return ErrForbidden
	}
	return s.q.SetPostClosed(ctx, db.SetPostClosedParams{ID: id, Closed: closed})
}

func (s *Service) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error {
	if pinned {
		return s.q.PinPost(ctx, id)
	}
	return s.q.UnpinPost(ctx, id)
}

// ── List queries ─────────────────────────────────────────────────────────────

func (s *Service) List(ctx context.Context, limit, offset int) ([]PostResponse, error) {
	posts, err := s.q.ListPosts(ctx, db.ListPostsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	return s.enrichPosts(ctx, posts)
}

func (s *Service) ListRecent(ctx context.Context, limit, offset int) ([]PostResponse, error) {
	posts, err := s.q.ListRecentPosts(ctx, db.ListRecentPostsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	return s.enrichPosts(ctx, posts)
}

func (s *Service) ListFeatured(ctx context.Context, limit, offset int) ([]PostResponse, error) {
	posts, err := s.q.ListFeaturedPosts(ctx, db.ListFeaturedPostsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	return s.enrichPosts(ctx, posts)
}

func (s *Service) ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]PostResponse, error) {
	posts, err := s.q.ListPostsByAuthor(ctx, db.ListPostsByAuthorParams{
		AuthorID: authorID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}
	return s.enrichPosts(ctx, posts)
}

func nullableStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}
