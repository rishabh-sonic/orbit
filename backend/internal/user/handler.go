package user

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

// GET /api/users/me
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	user, err := h.svc.GetByID(r.Context(), claims.UserID)
	if err != nil {
		middleware.NotFound(w, "user not found")
		return
	}
	middleware.Ok(w, user)
}

// PUT /api/users/me
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())

	var body struct {
		Username     *string `json:"username"`
		Introduction *string `json:"introduction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.BadRequest(w, "invalid request")
		return
	}

	user, err := h.svc.UpdateProfile(r.Context(), claims.UserID, UpdateProfileInput{
		Username:     body.Username,
		Introduction: body.Introduction,
	})
	if err != nil {
		middleware.Conflict(w, err.Error())
		return
	}
	middleware.Ok(w, user)
}

// POST /api/users/me/avatar — avatar URL is set after upload
func (h *Handler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	var body struct {
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AvatarURL == "" {
		middleware.BadRequest(w, "avatar_url required")
		return
	}

	user, err := h.svc.UpdateProfile(r.Context(), claims.UserID, UpdateProfileInput{
		Avatar: &body.AvatarURL,
	})
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, user)
}

// GET /api/users/{identifier}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")
	claims := middleware.ClaimsFromContext(r.Context())
	var viewerID *uuid.UUID
	if claims != nil {
		viewerID = &claims.UserID
	}
	user, err := h.svc.GetByIdentifier(r.Context(), identifier, viewerID)
	if err != nil {
		middleware.NotFound(w, "user not found")
		return
	}
	middleware.Ok(w, user)
}

// GET /api/users/{identifier}/following
func (h *Handler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")
	target, err := h.svc.GetByIdentifier(r.Context(), identifier, nil)
	if err != nil {
		middleware.NotFound(w, "user not found")
		return
	}
	limit, offset := paginate(r)
	users, err := h.svc.GetFollowing(r.Context(), target.ID, limit, offset)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, users)
}

// GET /api/users/{identifier}/followers
func (h *Handler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "identifier")
	target, err := h.svc.GetByIdentifier(r.Context(), identifier, nil)
	if err != nil {
		middleware.NotFound(w, "user not found")
		return
	}
	limit, offset := paginate(r)
	users, err := h.svc.GetFollowers(r.Context(), target.ID, limit, offset)
	if err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, users)
}

// POST /api/subscriptions/users/{username}
func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	username := chi.URLParam(r, "username")
	target, err := h.svc.GetByIdentifier(r.Context(), username, nil)
	if err != nil {
		middleware.NotFound(w, "user not found")
		return
	}
	if err := h.svc.Follow(r.Context(), claims.UserID, target.ID); err != nil {
		middleware.InternalError(w, err)
		return
	}
	middleware.Ok(w, map[string]bool{"following": true})
}

// DELETE /api/subscriptions/users/{username}
func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	username := chi.URLParam(r, "username")
	target, err := h.svc.GetByIdentifier(r.Context(), username, nil)
	if err != nil {
		middleware.NotFound(w, "user not found")
		return
	}
	if err := h.svc.Unfollow(r.Context(), claims.UserID, target.ID); err != nil {
		middleware.InternalError(w, err)
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
