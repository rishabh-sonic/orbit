package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newJWT() *token.Service {
	return token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
}

func makeToken(t *testing.T, svc *token.Service, role string) string {
	t.Helper()
	tok, err := svc.GenerateToken(uuid.New(), "alice", role)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	return tok
}

// --- Authenticate middleware ---

func TestAuthenticate_ValidToken(t *testing.T) {
	svc := newJWT()
	tok := makeToken(t, svc, "USER")

	var gotClaims *token.Claims
	handler := middleware.Authenticate(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims = middleware.ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if gotClaims == nil {
		t.Fatal("expected claims in context, got nil")
	}
	if gotClaims.Username != "alice" {
		t.Errorf("Username: got %q, want alice", gotClaims.Username)
	}
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	svc := newJWT()
	var gotClaims *token.Claims
	handler := middleware.Authenticate(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims = middleware.ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should still call next, just without claims
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	if gotClaims != nil {
		t.Error("expected nil claims for invalid token")
	}
}

func TestAuthenticate_NoToken(t *testing.T) {
	svc := newJWT()
	called := false
	handler := middleware.Authenticate(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler should be called even without token")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// --- RequireAuth middleware ---

func TestRequireAuth_WithValidClaims(t *testing.T) {
	claims := &token.Claims{Username: "alice", Role: "USER"}
	ctx := context.WithValue(context.Background(), token.ClaimsKey, claims)

	called := false
	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler should have been called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestRequireAuth_WithoutClaims(t *testing.T) {
	called := false
	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if called {
		t.Fatal("handler should NOT have been called for unauthenticated request")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

// --- RequireAdmin middleware ---

func TestRequireAdmin_AsAdmin(t *testing.T) {
	claims := &token.Claims{Username: "admin", Role: "ADMIN"}
	ctx := context.WithValue(context.Background(), token.ClaimsKey, claims)

	called := false
	handler := middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler should have been called for admin")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestRequireAdmin_AsRegularUser(t *testing.T) {
	claims := &token.Claims{Username: "alice", Role: "USER"}
	ctx := context.WithValue(context.Background(), token.ClaimsKey, claims)

	called := false
	handler := middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if called {
		t.Fatal("handler should NOT have been called for regular user")
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestRequireAdmin_Unauthenticated(t *testing.T) {
	handler := middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestClaimsFromContext_Present(t *testing.T) {
	claims := &token.Claims{Username: "bob", Role: "USER"}
	ctx := context.WithValue(context.Background(), token.ClaimsKey, claims)
	got := middleware.ClaimsFromContext(ctx)
	if got == nil || got.Username != "bob" {
		t.Errorf("expected bob's claims, got %+v", got)
	}
}

func TestClaimsFromContext_Absent(t *testing.T) {
	got := middleware.ClaimsFromContext(context.Background())
	if got != nil {
		t.Errorf("expected nil claims, got %+v", got)
	}
}
