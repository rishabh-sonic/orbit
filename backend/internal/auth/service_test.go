package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/auth"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/db/mock"
	"github.com/rishabh-sonic/orbit/internal/email"
	"github.com/rishabh-sonic/orbit/pkg/config"
	"golang.org/x/crypto/bcrypt"
)

func newJWTSvc() *auth.JWTService {
	return auth.NewJWTService("test-secret-32-bytes-long-enough", "reset-secret-also-long-enough!!", 24*time.Hour)
}

// noopEmail returns an email.Sender that is configured but uses a fake API key so
// it won't actually send anything (and won't panic on nil receiver).
func noopEmail() *email.Sender {
	return email.New(&config.Config{ResendAPIKey: "test-key", ResendFromEmail: "no-reply@test.local"})
}

func newService(q *mock.Querier) *auth.Service {
	return auth.NewService(q, newJWTSvc(), noopEmail())
}

// --- Register ---

func TestRegister_Success(t *testing.T) {
	userID := uuid.New()
	q := &mock.Querier{
		GetUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		GetUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		CreateUserFn: func(_ context.Context, _ db.CreateUserParams) (db.User, error) {
			return db.User{ID: userID, Username: "alice", Role: db.UserRoleUSER}, nil
		},
	}
	tok, err := newService(q).Register(context.Background(), auth.RegisterInput{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if tok == "" {
		t.Error("expected non-empty token")
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	q := &mock.Querier{
		GetUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: uuid.New(), Email: "alice@example.com"}, nil
		},
	}
	svc := newService(q)
	_, err := svc.Register(context.Background(), auth.RegisterInput{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if !errors.Is(err, auth.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got: %v", err)
	}
}

func TestRegister_UsernameTaken(t *testing.T) {
	q := &mock.Querier{
		GetUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, sql.ErrNoRows
		},
		GetUserByUsernameFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{ID: uuid.New(), Username: "alice"}, nil
		},
	}
	svc := newService(q)
	_, err := svc.Register(context.Background(), auth.RegisterInput{
		Username: "alice",
		Email:    "new@example.com",
		Password: "password123",
	})
	if !errors.Is(err, auth.ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got: %v", err)
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	userID := uuid.New()
	q := &mock.Querier{
		GetUserByEmailOrUsernameFn: func(_ context.Context, _ db.GetUserByEmailOrUsernameParams) (db.User, error) {
			return db.User{
				ID:           userID,
				Username:     "alice",
				Email:        "alice@example.com",
				PasswordHash: sql.NullString{String: string(hash), Valid: true},
				Verified:     true,
				Banned:       false,
				Role:         db.UserRoleUSER,
			}, nil
		},
	}
	svc := newService(q)
	tok, err := svc.Login(context.Background(), auth.LoginInput{
		Identifier: "alice@example.com",
		Password:   "password123",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if tok == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := newService(&mock.Querier{})
	_, err := svc.Login(context.Background(), auth.LoginInput{
		Identifier: "nobody@example.com",
		Password:   "password",
	})
	if !errors.Is(err, auth.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	q := &mock.Querier{
		GetUserByEmailOrUsernameFn: func(_ context.Context, _ db.GetUserByEmailOrUsernameParams) (db.User, error) {
			return db.User{
				ID:           uuid.New(),
				Username:     "alice",
				Verified:     true,
				PasswordHash: sql.NullString{String: string(hash), Valid: true},
			}, nil
		},
	}
	_, err := newService(q).Login(context.Background(), auth.LoginInput{
		Identifier: "alice",
		Password:   "wrong",
	})
	if !errors.Is(err, auth.ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}
}

func TestLogin_Banned(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	q := &mock.Querier{
		GetUserByEmailOrUsernameFn: func(_ context.Context, _ db.GetUserByEmailOrUsernameParams) (db.User, error) {
			return db.User{
				Verified:     true,
				Banned:       true,
				PasswordHash: sql.NullString{String: string(hash), Valid: true},
			}, nil
		},
	}
	_, err := newService(q).Login(context.Background(), auth.LoginInput{
		Identifier: "alice",
		Password:   "password",
	})
	if !errors.Is(err, auth.ErrBanned) {
		t.Errorf("expected ErrBanned, got: %v", err)
	}
}

// --- ForgotReset ---

func TestForgotReset_InvalidToken(t *testing.T) {
	err := newService(&mock.Querier{}).ForgotReset(context.Background(), "bad-token", "newpass")
	if !errors.Is(err, auth.ErrInvalidCode) {
		t.Errorf("expected ErrInvalidCode, got: %v", err)
	}
}
