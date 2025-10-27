package comment

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// POST /api/posts/{postId}/comments
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}

	var body struct {
		Content  string  `json:"content"`
		ParentID *string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		middleware.BadRequest(w, "content required")
		return
	}

	var parentID *uuid.UUID
	if body.ParentID != nil && *body.ParentID != "" {
		pid, err := uuid.Parse(*body.ParentID)
		if err != nil {
			middleware.BadRequest(w, "invalid parent_id")
			return
		}
		parentID = &pid
	}

	c, err := h.svc.Create(r.Context(), body.Content, claims.UserID, postID, parentID)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Created(w, c)
}

// POST /api/comments/{id}/replies
func (h *Handler) Reply(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	parentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid comment id")
		return
	}

	// Get parent comment to find postID
	parent, err := h.svc.q.GetCommentByID(r.Context(), parentID)
	if err != nil {
		middleware.NotFound(w, "comment not found")
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		middleware.BadRequest(w, "content required")
		return
	}

	c, err := h.svc.Create(r.Context(), body.Content, claims.UserID, parent.PostID, &parentID)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Created(w, c)
}

// GET /api/posts/{postId}/comments
func (h *Handler) ListForPost(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}
	limit, offset := paginate(r)
	comments, err := h.svc.ListForPost(r.Context(), postID, limit, offset)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, comments)
}

// GET /api/comments/{id}/replies
func (h *Handler) ListReplies(w http.ResponseWriter, r *http.Request) {
	parentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid comment id")
		return
	}
	limit, offset := paginate(r)
	replies, err := h.svc.ListReplies(r.Context(), parentID, limit, offset)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, replies)
}

// DELETE /api/comments/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid comment id")
		return
	}
	isAdmin := claims.Role == "ADMIN"
	if err := h.svc.Delete(r.Context(), id, claims.UserID, isAdmin); err != nil {
		switch err {
		case ErrNotFound:
			middleware.NotFound(w, "comment not found")
		case ErrForbidden:
			middleware.Forbidden(w, "")
		default:
			middleware.InternalError(w, err)
		}
		return
	}
	middleware.NoContent(w)
}

func paginate(r *http.Request) (int, int) {
	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}
