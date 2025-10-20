package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	mw "github.com/rishabh-sonic/orbit/internal/middleware"
	tokenpkg "github.com/rishabh-sonic/orbit/pkg/token"
)

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type Handler struct {
	svc   *Service
	oauth *OAuthService
	jwt   *JWTService
}

func NewHandler(svc *Service, oauth *OAuthService, jwt *JWTService) *Handler {
	return &Handler{svc: svc, oauth: oauth, jwt: jwt}
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}

// POST /api/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	body, err := decode[struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}](r)
	if err != nil {
		mw.BadRequest(w, "invalid request body")
		return
	}
	if body.Username == "" || body.Email == "" || body.Password == "" {
		mw.BadRequest(w, "username, email and password are required")
		return
	}

	token, err := h.svc.Register(r.Context(), RegisterInput{
		Username: body.Username,
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			mw.Conflict(w, "email already registered")
		case errors.Is(err, ErrUsernameTaken):
			mw.Conflict(w, "username already taken")
		default:
			mw.InternalError(w, err)
		}
		return
	}

	mw.Ok(w, map[string]string{"token": token})
}

// POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := decode[struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}](r)
	if err != nil || body.Identifier == "" || body.Password == "" {
		mw.BadRequest(w, "identifier and password are required")
		return
	}

	token, err := h.svc.Login(r.Context(), LoginInput{
		Identifier: body.Identifier,
		Password:   body.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrInvalidPassword):
			mw.Unauthorized(w, "invalid credentials")
		case errors.Is(err, ErrBanned):
			mw.Forbidden(w, "account is banned")
		default:
			mw.InternalError(w, err)
		}
		return
	}
	mw.Ok(w, map[string]string{"token": token})
}

// GET /api/auth/check
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	claims := tokenpkg.ClaimsFromContext(r.Context())
	if claims == nil {
		mw.Ok(w, map[string]bool{"authenticated": false})
		return
	}
	mw.Ok(w, map[string]any{
		"authenticated": true,
		"user_id":       claims.UserID,
		"username":      claims.Username,
		"role":          claims.Role,
	})
}

// POST /api/auth/forgot/send
func (h *Handler) ForgotSend(w http.ResponseWriter, r *http.Request) {
	body, err := decode[struct {
		Email string `json:"email"`
	}](r)
	if err != nil || body.Email == "" {
		mw.BadRequest(w, "email is required")
		return
	}
	_ = h.svc.ForgotSend(r.Context(), body.Email)
	mw.Ok(w, map[string]string{"message": "if that email exists, a reset code was sent"})
}

// POST /api/auth/forgot/verify
func (h *Handler) ForgotVerify(w http.ResponseWriter, r *http.Request) {
	body, err := decode[struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}](r)
	if err != nil || body.Email == "" || body.Code == "" {
		mw.BadRequest(w, "email and code are required")
		return
	}
	token, err := h.svc.ForgotVerify(r.Context(), body.Email, body.Code)
	if err != nil {
		mw.BadRequest(w, "invalid or expired code")
		return
	}
	mw.Ok(w, map[string]string{"reset_token": token})
}

// POST /api/auth/forgot/reset
func (h *Handler) ForgotReset(w http.ResponseWriter, r *http.Request) {
	body, err := decode[struct {
		ResetToken  string `json:"reset_token"`
		NewPassword string `json:"new_password"`
	}](r)
	if err != nil || body.ResetToken == "" || body.NewPassword == "" {
		mw.BadRequest(w, "reset_token and new_password are required")
		return
	}
	if err := h.svc.ForgotReset(r.Context(), body.ResetToken, body.NewPassword); err != nil {
		mw.BadRequest(w, err.Error())
		return
	}
	mw.Ok(w, map[string]string{"message": "password updated"})
}

// GET /oauth/google – initiates Google OAuth flow
func (h *Handler) OAuthGoogleRedirect(w http.ResponseWriter, r *http.Request) {
	if h.oauth.cfg.GoogleClientID == "" {
		http.Error(w, "Google OAuth not configured", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, h.oauth.GoogleAuthURL(generateState()), http.StatusFound)
}

// GET /oauth/github – initiates GitHub OAuth flow
func (h *Handler) OAuthGitHubRedirect(w http.ResponseWriter, r *http.Request) {
	if h.oauth.cfg.GitHubClientID == "" {
		http.Error(w, "GitHub OAuth not configured", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, h.oauth.GitHubAuthURL(generateState()), http.StatusFound)
}

// POST /api/auth/google
func (h *Handler) Google(w http.ResponseWriter, r *http.Request) {
	body, err := decode[struct {
		Code string `json:"code"`
	}](r)
	if err != nil || body.Code == "" {
		mw.BadRequest(w, "code required")
		return
	}
	token, err := h.oauth.HandleGoogle(r.Context(), body.Code)
	if err != nil {
		mw.Unauthorized(w, err.Error())
		return
	}
	mw.Ok(w, map[string]string{"token": token})
}

// POST /api/auth/github
func (h *Handler) GitHub(w http.ResponseWriter, r *http.Request) {
	body, err := decode[struct {
		Code string `json:"code"`
	}](r)
	if err != nil || body.Code == "" {
		mw.BadRequest(w, "code required")
		return
	}
	token, err := h.oauth.HandleGitHub(r.Context(), body.Code)
	if err != nil {
		mw.Unauthorized(w, err.Error())
		return
	}
	mw.Ok(w, map[string]string{"token": token})
}

