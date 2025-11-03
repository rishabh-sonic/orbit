package notification

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rishabh-sonic/orbit/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/notifications
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	limit, offset := paginate(r)
	notifs, err := h.svc.List(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, notifs)
}

// GET /api/notifications/unread-count
func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	count, err := h.svc.UnreadCount(r.Context(), claims.UserID)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]int64{"count": count})
}

// POST /api/notifications/read
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if err := h.svc.MarkRead(r.Context(), claims.UserID); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"ok": true})
}

// GET /api/notifications/prefs
func (h *Handler) GetPrefs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	prefs, err := h.svc.GetPreferences(r.Context(), claims.UserID)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, prefs)
}

// POST /api/notifications/prefs
func (h *Handler) UpdatePrefs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	var body struct {
		Type        string `json:"type"`
		InApp       bool   `json:"in_app_enabled"`
		PushEnabled bool   `json:"push_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Type == "" {
		middleware.BadRequest(w, "type required")
		return
	}
	if err := h.svc.UpdatePreference(r.Context(), claims.UserID, body.Type, body.InApp, body.PushEnabled); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"ok": true})
}

// GET /api/notifications/email-prefs
func (h *Handler) GetEmailPrefs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	prefs, err := h.svc.GetEmailPreferences(r.Context(), claims.UserID)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, prefs)
}

// POST /api/notifications/email-prefs
func (h *Handler) UpdateEmailPrefs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	var body struct {
		Type    string `json:"type"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Type == "" {
		middleware.BadRequest(w, "type required")
		return
	}
	if err := h.svc.UpdateEmailPreference(r.Context(), claims.UserID, body.Type, body.Enabled); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"ok": true})
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
