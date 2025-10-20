package auth_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/auth"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/db/mock"
	"github.com/rishabh-sonic/orbit/pkg/token"
	"golang.org/x/crypto/bcrypt"
)

// setupRouter builds a minimal chi router with auth routes.
func setupRouter(q *mock.Querier) http.Handler {
	jwtSvc := auth.NewJWTService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	svc := auth.NewService(q, jwtSvc, nil)
	oauthSvc := auth.NewOAuthService(nil, svc) // nil config — OAuth handlers will error, which is fine
	h := auth.NewHandler(svc, oauthSvc, jwtSvc)

	r := chi.NewRouter()
	r.Post("/api/auth/register", h.Register)
	r.Post("/api/auth/login", h.Login)
	r.Get("/api/auth/check", h.Check)
	r.Post("/api/auth/forgot/send", h.ForgotSend)
	r.Post("/api/auth/forgot/verify", h.ForgotVerify)
	r.Post("/api/auth/forgot/reset", h.ForgotReset)
	return r
}

func postJSON(t *testing.T, router http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func getWith(t *testing.T, router http.Handler, path, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// --- Register ---

func TestRegisterHandler_MissingFields(t *testing.T) {
	rr := postJSON(t, setupRouter(&mock.Querier{}), "/api/auth/register", map[string]string{
		"email": "alice@example.com",
		// missing username and password
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

func TestRegisterHandler_EmailTaken(t *testing.T) {
	q := &mock.Querier{
		GetUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: uuid.New()}, nil // found → taken
		},
	}
	rr := postJSON(t, setupRouter(q), "/api/auth/register", map[string]string{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "password123",
	})
	if rr.Code != http.StatusConflict {
		t.Errorf("status: got %d, want 409", rr.Code)
	}
}

func TestRegisterHandler_UsernameTaken(t *testing.T) {
	q := &mock.Querier{
		GetUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		GetUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: uuid.New()}, nil // found → taken
		},
	}
	rr := postJSON(t, setupRouter(q), "/api/auth/register", map[string]string{
		"username": "alice",
		"email":    "new@example.com",
		"password": "password123",
	})
	if rr.Code != http.StatusConflict {
		t.Errorf("status: got %d, want 409", rr.Code)
	}
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	setupRouter(&mock.Querier{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- Login ---

func TestLoginHandler_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	q := &mock.Querier{
		GetUserByEmailOrUsernameFn: func(_ context.Context, _ db.GetUserByEmailOrUsernameParams) (db.User, error) {
			return db.User{
				ID:           uuid.New(),
				Username:     "alice",
				Email:        "alice@example.com",
				PasswordHash: sql.NullString{String: string(hash), Valid: true},
				Verified:     true,
				Role:         db.UserRoleUSER,
			}, nil
		},
	}
	rr := postJSON(t, setupRouter(q), "/api/auth/login", map[string]string{
		"identifier": "alice@example.com",
		"password":   "password123",
	})
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200 — body: %s", rr.Code, rr.Body)
	}
	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Data.Token == "" {
		t.Error("expected token in response")
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	rr := postJSON(t, setupRouter(&mock.Querier{}), "/api/auth/login", map[string]string{
		"identifier": "nobody@example.com",
		"password":   "wrong",
	})
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestLoginHandler_MissingFields(t *testing.T) {
	rr := postJSON(t, setupRouter(&mock.Querier{}), "/api/auth/login", map[string]string{
		"identifier": "alice@example.com",
		// missing password
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- Check ---

func TestCheckHandler_Authenticated(t *testing.T) {
	jwtSvc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
	tok, _ := jwtSvc.GenerateToken(uuid.New(), "alice", "USER")

	// Build router with authenticate middleware
	svc := auth.NewService(&mock.Querier{}, jwtSvc, nil)
	oauthSvc := auth.NewOAuthService(nil, svc)
	h := auth.NewHandler(svc, oauthSvc, jwtSvc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if raw := token.ExtractBearerToken(r.Header.Get("Authorization")); raw != "" {
				if claims, err := jwtSvc.ValidateToken(raw); err == nil {
					ctx := context.WithValue(r.Context(), token.ClaimsKey, claims)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/api/auth/check", h.Check)

	rr := getWith(t, r, "/api/auth/check", tok)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	var resp struct {
		Data struct {
			Authenticated bool   `json:"authenticated"`
			Username      string `json:"username"`
		} `json:"data"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp.Data.Authenticated {
		t.Error("expected authenticated: true")
	}
	if resp.Data.Username != "alice" {
		t.Errorf("username: got %q, want alice", resp.Data.Username)
	}
}

func TestCheckHandler_Unauthenticated(t *testing.T) {
	rr := getWith(t, setupRouter(&mock.Querier{}), "/api/auth/check", "")
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	var resp struct {
		Data struct {
			Authenticated bool `json:"authenticated"`
		} `json:"data"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Data.Authenticated {
		t.Error("expected authenticated: false for unauthenticated request")
	}
}

// --- ForgotSend ---

func TestForgotSendHandler_MissingEmail(t *testing.T) {
	rr := postJSON(t, setupRouter(&mock.Querier{}), "/api/auth/forgot/send", map[string]string{})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

func TestForgotSendHandler_AlwaysOK(t *testing.T) {
	// Even for non-existent email, respond 200 to avoid email enumeration
	rr := postJSON(t, setupRouter(&mock.Querier{}), "/api/auth/forgot/send", map[string]string{
		"email": "nobody@example.com",
	})
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- ForgotVerify ---

func TestForgotVerifyHandler_MissingFields(t *testing.T) {
	rr := postJSON(t, setupRouter(&mock.Querier{}), "/api/auth/forgot/verify", map[string]string{
		"email": "alice@example.com",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

// --- ForgotReset ---

func TestForgotResetHandler_MissingFields(t *testing.T) {
	rr := postJSON(t, setupRouter(&mock.Querier{}), "/api/auth/forgot/reset", map[string]string{
		"reset_token": "sometoken",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}

func TestForgotResetHandler_InvalidToken(t *testing.T) {
	rr := postJSON(t, setupRouter(&mock.Querier{}), "/api/auth/forgot/reset", map[string]string{
		"reset_token":  "bad.token.here",
		"new_password": "newpassword123",
	})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", rr.Code)
	}
}
