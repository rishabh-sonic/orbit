package post

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

// POST /api/posts
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.BadRequest(w, "invalid request")
		return
	}

	post, err := h.svc.Create(r.Context(), CreateInput{
		Title:    body.Title,
		Content:  body.Content,
		AuthorID: claims.UserID,
	})
	if err != nil {
		middleware.BadRequest(w, err.Error())
		return
	}
	middleware.Created(w, post)
}

// GET /api/posts/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}
	claims := middleware.ClaimsFromContext(r.Context())
	var viewerID *uuid.UUID
	if claims != nil {
		viewerID = &claims.UserID
	}

	post, err := h.svc.GetByID(r.Context(), id, viewerID)
	if err != nil {
		middleware.NotFound(w, "post not found")
		return
	}
	middleware.Ok(w, post)
}

// PUT /api/posts/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}

	var body struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.BadRequest(w, "invalid request")
		return
	}

	post, err := h.svc.Update(r.Context(), id, claims.UserID, UpdateInput{
		Title:   body.Title,
		Content: body.Content,
	})
	if err != nil {
		switch err {
		case ErrNotFound:
			middleware.NotFound(w, "post not found")
		case ErrForbidden:
			middleware.Forbidden(w, "")
		default:
			middleware.InternalError(w, err)
		}
		return
	}
	middleware.Ok(w, post)
}

// DELETE /api/posts/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}
	isAdmin := claims.Role == "ADMIN"
	if err := h.svc.Delete(r.Context(), id, claims.UserID, isAdmin); err != nil {
		switch err {
		case ErrNotFound:
			middleware.NotFound(w, "post not found")
		case ErrForbidden:
			middleware.Forbidden(w, "")
		default:
			middleware.InternalError(w, err)
		}
		return
	}
	middleware.NoContent(w)
}

// POST /api/posts/{id}/close
func (h *Handler) Close(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	if err := h.svc.SetClosed(r.Context(), id, claims.UserID, true); err != nil {
		middleware.Forbidden(w, err.Error())
		return
	}
	middleware.Ok(w, map[string]bool{"closed": true})
}

// POST /api/posts/{id}/reopen
func (h *Handler) Reopen(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	if err := h.svc.SetClosed(r.Context(), id, claims.UserID, false); err != nil {
		middleware.Forbidden(w, err.Error())
		return
	}
	middleware.Ok(w, map[string]bool{"closed": false})
}

// GET /api/posts
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginate(r)

	// If user_id is provided, return only that author's posts.
	if rawUID := r.URL.Query().Get("user_id"); rawUID != "" {
		authorID, err := uuid.Parse(rawUID)
		if err != nil {
			middleware.BadRequest(w, "invalid user_id")
			return
		}
		posts, err := h.svc.ListByAuthor(r.Context(), authorID, limit, offset)
		if err != nil {
			middleware.InternalError(w, err)
			return
		}
		middleware.Ok(w, posts)
		return
	}

	posts, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, posts)
}

// GET /api/posts/recent
func (h *Handler) Recent(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginate(r)
	posts, err := h.svc.ListRecent(r.Context(), limit, offset)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, posts)
}

// GET /api/posts/featured
func (h *Handler) Featured(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginate(r)
	posts, err := h.svc.ListFeatured(r.Context(), limit, offset)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, posts)
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
	} else if v := r.URL.Query().Get("page"); v != "" {
		// Frontend sends ?page=N — convert to zero-based offset.
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			offset = (n - 1) * limit
		}
	}
	return limit, offset
}
