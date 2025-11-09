package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/comment"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/internal/post"
)

type Handler struct {
	q        db.Querier
	postSvc  *post.Service
	commentSvc *comment.Service
}

func NewHandler(q db.Querier, postSvc *post.Service, commentSvc *comment.Service) *Handler {
	return &Handler{q: q, postSvc: postSvc, commentSvc: commentSvc}
}

// GET /api/admin/config
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListConfig(r.Context())
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	result := make(map[string]string, len(rows))
	for _, row := range rows {
		result[row.Key] = row.Value
	}
	middleware.Ok(w, result)
}

// POST /api/admin/config
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.BadRequest(w, "invalid request body")
		return
	}
	for k, v := range body {
		if err := h.q.UpsertConfigValue(r.Context(), db.UpsertConfigValueParams{Key: k, Value: v}); err != nil {
			middleware.InternalError(w, err)
			return
		}
	}
	middleware.Ok(w, map[string]bool{"ok": true})
}

// GET /api/admin/users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}
	users, err := h.q.ListUsers(r.Context(), db.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, users)
}

// POST /api/admin/users/{id}/ban
func (h *Handler) BanUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid user id")
		return
	}
	if err := h.q.SetUserBanned(r.Context(), db.SetUserBannedParams{ID: id, Banned: true}); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"banned": true})
}

// POST /api/admin/users/{id}/unban
func (h *Handler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid user id")
		return
	}
	if err := h.q.SetUserBanned(r.Context(), db.SetUserBannedParams{ID: id, Banned: false}); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"banned": false})
}

// DELETE /api/admin/posts/{id}
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}
	if err := h.postSvc.Delete(r.Context(), id, claims.UserID, true); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.NoContent(w)
}

// POST /api/admin/posts/{id}/pin
func (h *Handler) PinPost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}
	if err := h.postSvc.SetPinned(r.Context(), id, true); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"pinned": true})
}

// POST /api/admin/posts/{id}/unpin
func (h *Handler) UnpinPost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}
	if err := h.postSvc.SetPinned(r.Context(), id, false); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"pinned": false})
}

// DELETE /api/admin/comments/{id}
func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid comment id")
		return
	}
	if err := h.commentSvc.Delete(r.Context(), id, claims.UserID, true); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.NoContent(w)
}

// POST /api/admin/comments/{id}/pin
func (h *Handler) PinComment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid comment id")
		return
	}
	if err := h.commentSvc.Pin(r.Context(), id); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"pinned": true})
}

// POST /api/admin/comments/{id}/unpin
func (h *Handler) UnpinComment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid comment id")
		return
	}
	if err := h.commentSvc.Unpin(r.Context(), id); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"pinned": false})
}

// GET /api/admin/stats/dau
func (h *Handler) StatDAU(w http.ResponseWriter, r *http.Request) {
	count, err := h.q.CountDailyActiveUsers(r.Context(), time.Now())
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]int64{"dau": count})
}

// GET /api/admin/stats/new-users-range?from=2024-01-01&to=2024-01-31
func (h *Handler) StatNewUsers(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseDateRange(r)
	if err != nil {
		middleware.BadRequest(w, "invalid date range; use from=YYYY-MM-DD&to=YYYY-MM-DD")
		return
	}
	count, err := h.q.CountNewUsersInRange(r.Context(), db.CountNewUsersInRangeParams{
		CreatedAt:   from,
		CreatedAt_2: to,
	})
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]int64{"count": count})
}

// GET /api/admin/stats/posts-range
func (h *Handler) StatPosts(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseDateRange(r)
	if err != nil {
		middleware.BadRequest(w, err.Error())
		return
	}
	count, err := h.q.CountPostsInRange(r.Context(), db.CountPostsInRangeParams{
		CreatedAt:   from,
		CreatedAt_2: to,
	})
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]int64{"count": count})
}

// GET /api/admin/stats/dau-range
func (h *Handler) StatDAURange(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseDateRange(r)
	if err != nil {
		middleware.BadRequest(w, err.Error())
		return
	}
	rows, err := h.q.CountDAUInRange(r.Context(), db.CountDAUInRangeParams{
		VisitDate:   from,
		VisitDate_2: to,
	})
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, rows)
}

func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("from and to required")
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to.Add(24 * time.Hour), nil
}

