package push

import (
	"encoding/json"
	"net/http"

	"github.com/rishabh-sonic/orbit/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/push/public-key
func (h *Handler) PublicKey(w http.ResponseWriter, r *http.Request) {
	middleware.Ok(w, map[string]string{"public_key": h.svc.PublicKey()})
}

// POST /api/push/subscribe
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	var body struct {
		Endpoint string `json:"endpoint"`
		P256dh   string `json:"p256dh"`
		Auth     string `json:"auth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Endpoint == "" {
		middleware.BadRequest(w, "endpoint, p256dh and auth required")
		return
	}
	if err := h.svc.Subscribe(r.Context(), claims.UserID, body.Endpoint, body.P256dh, body.Auth); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"subscribed": true})
}
