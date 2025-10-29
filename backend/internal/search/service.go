package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/pkg/config"
)

type Service struct {
	client      *opensearchapi.Client
	cfg         *config.Config
	q           db.Querier
	indexPrefix string
	enabled     bool
}

func NewService(client *opensearchapi.Client, q db.Querier, cfg *config.Config) *Service {
	return &Service{
		client:      client,
		cfg:         cfg,
		q:           q,
		indexPrefix: cfg.SearchIndexPrefix,
		enabled:     cfg.SearchEnabled,
	}
}

func (s *Service) postsIndex() string { return s.indexPrefix + "_posts" }
func (s *Service) usersIndex() string { return s.indexPrefix + "_users" }

// ── Index document types ─────────────────────────────────────────────────────

type PostDocument struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	AuthorID  string `json:"author_id"`
	CreatedAt string `json:"created_at"`
}

type UserDocument struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar,omitempty"`
}

// ── Response types returned to the API layer ─────────────────────────────────

type PostAuthor struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
}

type PostResult struct {
	ID           string      `json:"id"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	Status       string      `json:"status"`
	PinnedAt     *time.Time  `json:"pinned_at"`
	Views        int32       `json:"views"`
	CommentCount int32       `json:"comment_count"`
	CreatedAt    time.Time   `json:"created_at"`
	Author       PostAuthor  `json:"author"`
}

type UserResult struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	Avatar       *string `json:"avatar"`
	Introduction *string `json:"introduction"`
}

type GlobalResult struct {
	Posts []PostResult `json:"posts"`
	Users []UserResult `json:"users"`
}

// ── Indexing ─────────────────────────────────────────────────────────────────

func (s *Service) IndexPost(ctx context.Context, post db.Post) {
	if !s.enabled {
		return
	}
	doc := PostDocument{
		ID:        post.ID.String(),
		Title:     post.Title,
		Content:   post.Content,
		AuthorID:  post.AuthorID.String(),
		CreatedAt: post.CreatedAt.String(),
	}
	body, _ := json.Marshal(doc)
	// Use Index (PUT /_doc/{id}) so this is an upsert — safe to call for
	// both new and existing documents, unlike Document.Create which errors
	// with 409 when the document already exists.
	_, err := s.client.Index(ctx, opensearchapi.IndexReq{
		Index:      s.postsIndex(),
		DocumentID: post.ID.String(),
		Body:       bytes.NewReader(body),
	})
	if err != nil {
		slog.Error("index post", "err", err)
	}
}

func (s *Service) IndexUser(ctx context.Context, user db.User) {
	if !s.enabled {
		return
	}
	avatar := ""
	if user.Avatar.Valid {
		avatar = user.Avatar.String
	}
	doc := UserDocument{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Avatar:   avatar,
	}
	body, _ := json.Marshal(doc)
	_, err := s.client.Index(ctx, opensearchapi.IndexReq{
		Index:      s.usersIndex(),
		DocumentID: user.ID.String(),
		Body:       bytes.NewReader(body),
	})
	if err != nil {
		slog.Error("index user", "err", err)
	}
}

func (s *Service) DeletePost(ctx context.Context, postID string) {
	if !s.enabled {
		return
	}
	_, err := s.client.Document.Delete(ctx, opensearchapi.DocumentDeleteReq{
		Index:      s.postsIndex(),
		DocumentID: postID,
	})
	if err != nil {
		slog.Error("delete post from index", "err", err)
	}
}

// ── Search ───────────────────────────────────────────────────────────────────

func (s *Service) SearchPosts(ctx context.Context, query string, field string) ([]json.RawMessage, error) {
	if !s.enabled {
		return nil, nil
	}
	if field == "" {
		field = "title,content"
	}
	fields := strings.Split(field, ",")
	return s.search(ctx, s.postsIndex(), query, fields)
}

func (s *Service) SearchUsers(ctx context.Context, query string) ([]json.RawMessage, error) {
	if !s.enabled {
		return nil, nil
	}
	return s.search(ctx, s.usersIndex(), query, []string{"username", "email"})
}

// SearchGlobal returns enriched posts and users fetched from the DB after
// resolving IDs from OpenSearch. This ensures the response matches exactly
// what the frontend PostCard and user-list components expect.
func (s *Service) SearchGlobal(ctx context.Context, query string) (*GlobalResult, error) {
	result := &GlobalResult{
		Posts: []PostResult{},
		Users: []UserResult{},
	}
	if !s.enabled {
		return result, nil
	}

	// ── Posts ────────────────────────────────────────────────────────────────
	postHits, err := s.search(ctx, s.postsIndex(), query, []string{"title", "content"})
	if err == nil {
		for _, raw := range postHits {
			var doc PostDocument
			if json.Unmarshal(raw, &doc) != nil {
				continue
			}
			id, err := uuid.Parse(doc.ID)
			if err != nil {
				continue
			}
			p, err := s.q.GetPostByID(ctx, id)
			if err != nil {
				continue
			}
			author, err := s.q.GetUserByID(ctx, p.AuthorID)
			if err != nil {
				author = db.User{ID: p.AuthorID, Username: "deleted"}
			}
			var avatar *string
			if author.Avatar.Valid {
				v := author.Avatar.String
				avatar = &v
			}
			status := "open"
			if p.Closed {
				status = "closed"
			}
			var pinnedAt *time.Time
			if p.PinnedAt.Valid {
				t := p.PinnedAt.Time
				pinnedAt = &t
			}
			result.Posts = append(result.Posts, PostResult{
				ID:           p.ID.String(),
				Title:        p.Title,
				Content:      p.Content,
				Status:       status,
				PinnedAt:     pinnedAt,
				Views:        p.Views,
				CommentCount: p.CommentCount,
				CreatedAt:    p.CreatedAt,
				Author: PostAuthor{
					ID:       author.ID.String(),
					Username: author.Username,
					Avatar:   avatar,
				},
			})
		}
	}

	// ── Users ────────────────────────────────────────────────────────────────
	userHits, err := s.search(ctx, s.usersIndex(), query, []string{"username", "email"})
	if err == nil {
		for _, raw := range userHits {
			var doc UserDocument
			if json.Unmarshal(raw, &doc) != nil {
				continue
			}
			id, err := uuid.Parse(doc.ID)
			if err != nil {
				continue
			}
			u, err := s.q.GetUserByID(ctx, id)
			if err != nil {
				continue
			}
			var avatar *string
			if u.Avatar.Valid {
				v := u.Avatar.String
				avatar = &v
			}
			var intro *string
			if u.Introduction.Valid {
				v := u.Introduction.String
				intro = &v
			}
			result.Users = append(result.Users, UserResult{
				ID:           u.ID.String(),
				Username:     u.Username,
				Avatar:       avatar,
				Introduction: intro,
			})
		}
	}

	return result, nil
}

// ── Internal helpers ─────────────────────────────────────────────────────────

func (s *Service) search(ctx context.Context, index, query string, fields []string) ([]json.RawMessage, error) {
	queryBody := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":     query,
				"fields":    fields,
				"type":      "best_fields",
				"fuzziness": "AUTO",
			},
		},
		"size": 20,
	}

	body, _ := json.Marshal(queryBody)
	resp, err := s.client.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    bytes.NewReader(body),
	})
	if err != nil {
		return nil, fmt.Errorf("opensearch search: %w", err)
	}

	hits := make([]json.RawMessage, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		hits = append(hits, h.Source)
	}
	return hits, nil
}

// ReindexAll pages through all posts and users in the DB and indexes them.
// Call this on startup when SEARCH_REINDEX_ON_STARTUP=true.
func (s *Service) ReindexAll(ctx context.Context) error {
	if !s.enabled {
		return nil
	}
	const batch = 100

	// Reindex posts
	for offset := 0; ; offset += batch {
		posts, err := s.q.ListPosts(ctx, db.ListPostsParams{Limit: batch, Offset: int32(offset)})
		if err != nil || len(posts) == 0 {
			break
		}
		for _, p := range posts {
			s.IndexPost(ctx, p)
		}
		if len(posts) < batch {
			break
		}
	}

	// Reindex users
	for offset := 0; ; offset += batch {
		users, err := s.q.ListUsers(ctx, db.ListUsersParams{Limit: batch, Offset: int32(offset)})
		if err != nil || len(users) == 0 {
			break
		}
		for _, u := range users {
			s.IndexUser(ctx, u)
		}
		if len(users) < batch {
			break
		}
	}

	return nil
}

func (s *Service) EnsureIndices(ctx context.Context) error {
	if !s.enabled {
		return nil
	}
	for _, index := range []string{s.postsIndex(), s.usersIndex()} {
		if err := s.createIndex(ctx, index); err != nil {
			slog.Warn("ensure index", "index", index, "err", err)
		}
	}
	return nil
}

func (s *Service) createIndex(ctx context.Context, index string) error {
	_, err := s.client.Indices.Exists(ctx, opensearchapi.IndicesExistsReq{
		Indices: []string{index},
	})
	if err == nil {
		return nil // index already exists
	}

	mappingBody := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"title":    map[string]string{"type": "text"},
				"content":  map[string]string{"type": "text"},
				"username": map[string]string{"type": "text"},
				"email":    map[string]string{"type": "keyword"},
			},
		},
	}
	body, _ := json.Marshal(mappingBody)
	_, err = s.client.Indices.Create(ctx, opensearchapi.IndicesCreateReq{
		Index: index,
		Body:  bytes.NewReader(body),
	})
	return err
}
