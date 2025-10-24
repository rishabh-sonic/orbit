package post

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/middleware"
)

type SubscriptionHandler struct {
	q db.Querier
}

func NewSubscriptionHandler(q db.Querier) *SubscriptionHandler {
	return &SubscriptionHandler{q: q}
}

// POST /api/subscriptions/posts/{postId}
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}
	if err := h.q.SubscribeToPost(r.Context(), db.SubscribeToPostParams{
		PostID: postID, UserID: claims.UserID,
	}); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"subscribed": true})
}

// DELETE /api/subscriptions/posts/{postId}
func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		middleware.BadRequest(w, "invalid post id")
		return
	}
	if err := h.q.UnsubscribeFromPost(r.Context(), db.UnsubscribeFromPostParams{
		PostID: postID, UserID: claims.UserID,
	}); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.NoContent(w)
}
