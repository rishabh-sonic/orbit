package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

func newService() *token.Service {
	return token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
}

func TestGenerateAndValidateToken(t *testing.T) {
	svc := newService()
	userID := uuid.New()

	tok, err := svc.GenerateToken(userID, "alice", "USER")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if tok == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID: got %v, want %v", claims.UserID, userID)
	}
	if claims.Username != "alice" {
		t.Errorf("Username: got %v, want alice", claims.Username)
	}
	if claims.Role != "USER" {
		t.Errorf("Role: got %v, want USER", claims.Role)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := newService()
	svc2 := token.NewService("completely-different-secret!!!!!", "reset-secret-also-long-enough!!", time.Hour)

	tok, _ := svc1.GenerateToken(uuid.New(), "alice", "USER")
	if _, err := svc2.ValidateToken(tok); err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidateToken_Tampered(t *testing.T) {
	svc := newService()
	tok, _ := svc.GenerateToken(uuid.New(), "alice", "USER")
	// Corrupt the last character
	tampered := tok[:len(tok)-1] + "X"
	if _, err := svc.ValidateToken(tampered); err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	svc := token.NewService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", -1*time.Minute)
	tok, err := svc.GenerateToken(uuid.New(), "alice", "USER")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if _, err := svc.ValidateToken(tok); err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateToken_Empty(t *testing.T) {
	svc := newService()
	if _, err := svc.ValidateToken(""); err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestGenerateAndValidateResetToken(t *testing.T) {
	svc := newService()
	email := "user@example.com"

	tok, err := svc.GenerateResetToken(email)
	if err != nil {
		t.Fatalf("GenerateResetToken: %v", err)
	}

	got, err := svc.ValidateResetToken(tok)
	if err != nil {
		t.Fatalf("ValidateResetToken: %v", err)
	}
	if got != email {
		t.Errorf("email: got %q, want %q", got, email)
	}
}

func TestValidateResetToken_WrongSecret(t *testing.T) {
	svc1 := newService()
	svc2 := token.NewService("test-secret-32-bytes-long-enough", "totally-different-reset-secret!!", time.Hour)

	tok, _ := svc1.GenerateResetToken("user@example.com")
	if _, err := svc2.ValidateResetToken(tok); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"valid bearer", "Bearer abc123", "abc123"},
		{"no prefix", "abc123", ""},
		{"empty", "", ""},
		{"bearer only", "Bearer ", ""},
		{"lowercase bearer", "bearer abc123", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := token.ExtractBearerToken(tc.header)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClaimsFromContext_Set(t *testing.T) {
	claims := &token.Claims{Username: "alice", Role: "USER"}
	ctx := context.WithValue(context.Background(), token.ClaimsKey, claims)
	got := token.ClaimsFromContext(ctx)
	if got == nil {
		t.Fatal("expected claims, got nil")
	}
	if got.Username != "alice" {
		t.Errorf("Username: got %q, want alice", got.Username)
	}
}

func TestClaimsFromContext_Missing(t *testing.T) {
	got := token.ClaimsFromContext(context.Background())
	if got != nil {
		t.Fatalf("expected nil claims, got %+v", got)
	}
}

func TestAdminRolePreserved(t *testing.T) {
	svc := newService()
	id := uuid.New()
	tok, _ := svc.GenerateToken(id, "admin_user", "ADMIN")
	claims, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Role != "ADMIN" {
		t.Errorf("expected ADMIN role, got %q", claims.Role)
	}
}
