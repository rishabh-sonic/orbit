package search

import (
	"net/http"

	"github.com/rishabh-sonic/orbit/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/search/posts?q=...&field=title|content
func (h *Handler) SearchPosts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	field := r.URL.Query().Get("field")
	if q == "" {
		middleware.BadRequest(w, "q parameter required")
		return
	}
	results, err := h.svc.SearchPosts(r.Context(), q, field)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, results)
}

// GET /api/search/posts/title?q=...
func (h *Handler) SearchPostTitles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		middleware.BadRequest(w, "q parameter required")
		return
	}
	results, err := h.svc.SearchPosts(r.Context(), q, "title")
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, results)
}

// GET /api/search/posts/content?q=...
func (h *Handler) SearchPostContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		middleware.BadRequest(w, "q parameter required")
		return
	}
	results, err := h.svc.SearchPosts(r.Context(), q, "content")
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, results)
}

// GET /api/search/users?q=...
func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		middleware.BadRequest(w, "q parameter required")
		return
	}
	results, err := h.svc.SearchUsers(r.Context(), q)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, results)
}

// GET /api/search/global?q=...
func (h *Handler) SearchGlobal(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		middleware.BadRequest(w, "q parameter required")
		return
	}
	results, err := h.svc.SearchGlobal(r.Context(), q)
	if err != nil {
		// Return empty results on search error rather than a 500.
		middleware.Ok(w, &GlobalResult{Posts: []PostResult{}, Users: []UserResult{}})
		return
	}
	middleware.Ok(w, results)
}
