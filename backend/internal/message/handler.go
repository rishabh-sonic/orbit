package message

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/middleware"
)

type Handler struct {
	svc *Service
	q   db.Querier
}

func NewHandler(svc *Service, q db.Querier) *Handler {
	return &Handler{svc: svc, q: q}
}

// GET /api/messages/conversations
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	limit, offset := paginate(r)
	convs, err := h.svc.ListConversations(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, convs)
}

// GET /api/messages/conversations/{id}
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid conversation id")
		return
	}
	conv, err := h.svc.GetConversation(r.Context(), convID, claims.UserID)
	if err != nil {
		middleware.NotFound(w, "conversation not found")
		return
	}
	middleware.Ok(w, conv)
}

// POST /api/messages/conversations — start or reopen a conversation by username
func (h *Handler) StartConversation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())

	var body struct {
		Username string `json:"username"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" {
		middleware.BadRequest(w, "username required")
		return
	}

	// Look up recipient by username
	recipient, err := h.q.GetUserByUsername(r.Context(), body.Username)
	if err != nil {
		middleware.NotFound(w, "user not found")
		return
	}
	if recipient.ID == claims.UserID {
		middleware.BadRequest(w, "cannot message yourself")
		return
	}

	conv, err := h.svc.GetOrCreateConversation(r.Context(), claims.UserID, recipient.ID)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}

	if body.Content != "" {
		msg, err := h.svc.SendMessage(r.Context(), conv.ID, claims.UserID, body.Content, nil)
		if err != nil {
			middleware.InternalError(w, err)
			return
		}
		middleware.Created(w, map[string]any{"conversation": conv, "message": msg})
		return
	}
	middleware.Created(w, conv)
}

// POST /api/messages/conversations/{id}/messages
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid conversation id")
		return
	}

	var body struct {
		Content   string  `json:"content"`
		ReplyToID *string `json:"reply_to_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		middleware.BadRequest(w, "content required")
		return
	}

	var replyToID *uuid.UUID
	if body.ReplyToID != nil {
		id, err := uuid.Parse(*body.ReplyToID)
		if err == nil {
			replyToID = &id
		}
	}

	msg, err := h.svc.SendMessage(r.Context(), convID, claims.UserID, body.Content, replyToID)
	if err != nil {
		switch err {
		case ErrForbidden:
			middleware.Forbidden(w, "not a participant")
		default:
			middleware.InternalError(w, err)
		}
		return
	}
	middleware.Created(w, msg)
}

// GET /api/messages/conversations/{id}/messages
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid conversation id")
		return
	}
	limit, offset := paginate(r)
	msgs, err := h.svc.ListMessages(r.Context(), convID, claims.UserID, limit, offset)
	if err != nil {
		switch err {
		case ErrForbidden:
			middleware.Forbidden(w, "not a participant")
		default:
			middleware.InternalError(w, err)
		}
		return
	}
	middleware.Ok(w, msgs)
}

// POST /api/messages/conversations/{id}/read
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	convID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.BadRequest(w, "invalid conversation id")
		return
	}
	if err := h.svc.MarkRead(r.Context(), convID, claims.UserID); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"ok": true})
}

// GET /api/messages/unread-count
func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	count, err := h.svc.TotalUnreadCount(r.Context(), claims.UserID)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]int64{"count": count})
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
