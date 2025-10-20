package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/email"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrEmailTaken        = errors.New("email already registered")
	ErrUsernameTaken     = errors.New("username already taken")
	ErrBanned            = errors.New("account is banned")
	ErrInvalidCode       = errors.New("invalid or expired code")
)

// UserIndexer is implemented by the search service; optional — nil disables indexing.
type UserIndexer interface {
	IndexUser(ctx context.Context, user db.User)
}

type Service struct {
	q       db.Querier
	jwt     *JWTService
	email   *email.Sender
	indexer UserIndexer // optional
}

func NewService(q db.Querier, jwt *JWTService, email *email.Sender, indexer UserIndexer) *Service {
	return &Service{q: q, jwt: jwt, email: email, indexer: indexer}
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

// Register creates a verified user and returns a JWT immediately.
func (s *Service) Register(ctx context.Context, in RegisterInput) (string, error) {
	// Check uniqueness
	if _, err := s.q.GetUserByEmail(ctx, in.Email); err == nil {
		return "", ErrEmailTaken
	}
	if _, err := s.q.GetUserByUsername(ctx, in.Username); err == nil {
		return "", ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.q.CreateUser(ctx, db.CreateUserParams{
		Username:     in.Username,
		Email:        in.Email,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
		Verified:     true,
		Avatar:       sql.NullString{},
		Role:         db.UserRoleUSER,
	})
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	if s.indexer != nil {
		go s.indexer.IndexUser(context.Background(), user)
	}
	return s.jwt.GenerateToken(user.ID, user.Username, string(user.Role))
}

type LoginInput struct {
	Identifier string // email or username
	Password   string
}

// Login authenticates via password and returns a JWT.
func (s *Service) Login(ctx context.Context, in LoginInput) (string, error) {
	user, err := s.q.GetUserByEmailOrUsername(ctx, db.GetUserByEmailOrUsernameParams{
		Email:    in.Identifier,
		Username: in.Identifier,
	})
	if err != nil {
		return "", ErrUserNotFound
	}

	if user.Banned {
		return "", ErrBanned
	}
	if !user.PasswordHash.Valid {
		return "", ErrInvalidPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(in.Password)); err != nil {
		return "", ErrInvalidPassword
	}

	return s.jwt.GenerateToken(user.ID, user.Username, string(user.Role))
}

// ForgotSend sends a password reset code.
func (s *Service) ForgotSend(ctx context.Context, emailAddr string) error {
	if _, err := s.q.GetUserByEmail(ctx, emailAddr); err != nil {
		// Don't reveal whether email exists
		return nil
	}
	return s.sendVerificationCode(ctx, emailAddr, "RESET")
}

// ForgotVerify checks the reset code and returns a short-lived reset token.
func (s *Service) ForgotVerify(ctx context.Context, emailAddr, code string) (string, error) {
	vc, err := s.q.GetVerificationCode(ctx, db.GetVerificationCodeParams{
		Email: emailAddr,
		Type:  "RESET",
	})
	if err != nil || vc.Code != code {
		return "", ErrInvalidCode
	}
	if err := s.q.MarkVerificationCodeUsed(ctx, vc.ID); err != nil {
		return "", err
	}
	return s.jwt.GenerateResetToken(emailAddr)
}

// ForgotReset sets a new password using the reset token.
func (s *Service) ForgotReset(ctx context.Context, resetToken, newPassword string) error {
	emailAddr, err := s.jwt.ValidateResetToken(resetToken)
	if err != nil {
		return ErrInvalidCode
	}

	user, err := s.q.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		return ErrUserNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.q.SetUserPasswordHash(ctx, db.SetUserPasswordHashParams{
		ID:           user.ID,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
	})
}

// GetOrCreateOAuthUser finds or creates a user for an OAuth login.
func (s *Service) GetOrCreateOAuthUser(ctx context.Context, provider, providerID, emailAddr, username, avatar string) (string, error) {
	// Try to find by OAuth account first
	oa, err := s.q.GetOAuthAccount(ctx, db.GetOAuthAccountParams{
		Provider:   provider,
		ProviderID: providerID,
	})
	if err == nil {
		user, err := s.q.GetUserByID(ctx, oa.UserID)
		if err != nil {
			return "", err
		}
		if user.Banned {
			return "", ErrBanned
		}
		return s.jwt.GenerateToken(user.ID, user.Username, string(user.Role))
	}

	// Try to find existing user by email
	var userID uuid.UUID
	existingUser, err := s.q.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		// Create new user
		safeUsername := s.ensureUniqueUsername(ctx, username)
		newUser, err := s.q.CreateUser(ctx, db.CreateUserParams{
			Username:     safeUsername,
			Email:        emailAddr,
			PasswordHash: sql.NullString{},
			Verified:     true,
			Avatar:       sql.NullString{String: avatar, Valid: avatar != ""},
			Role:         db.UserRoleUSER,
		})
		if err != nil {
			return "", fmt.Errorf("create oauth user: %w", err)
		}
		userID = newUser.ID
		if s.indexer != nil {
			go s.indexer.IndexUser(context.Background(), newUser)
		}
	} else {
		userID = existingUser.ID
	}

	if err := s.q.CreateOAuthAccount(ctx, db.CreateOAuthAccountParams{
		UserID:     userID,
		Provider:   provider,
		ProviderID: providerID,
	}); err != nil {
		return "", err
	}

	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return s.jwt.GenerateToken(user.ID, user.Username, string(user.Role))
}

func (s *Service) sendVerificationCode(ctx context.Context, emailAddr, codeType string) error {
	code := generateCode(6)
	expires := time.Now().Add(15 * time.Minute)

	_ = s.q.DeleteVerificationCodes(ctx, db.DeleteVerificationCodesParams{
		Email: emailAddr,
		Type:  codeType,
	})

	_, err := s.q.CreateVerificationCode(ctx, db.CreateVerificationCodeParams{
		Email:     emailAddr,
		Code:      code,
		Type:      codeType,
		ExpiresAt: expires,
	})
	if err != nil {
		return fmt.Errorf("store code: %w", err)
	}

	return s.email.SendPasswordReset(emailAddr, code)
}

func (s *Service) ensureUniqueUsername(ctx context.Context, base string) string {
	candidate := base
	for i := 1; i <= 99; i++ {
		if _, err := s.q.GetUserByUsername(ctx, candidate); err != nil {
			return candidate
		}
		candidate = fmt.Sprintf("%s%d", base, i)
	}
	return fmt.Sprintf("%s_%s", base, generateCode(4))
}

func generateCode(n int) string {
	const digits = "0123456789"
	result := make([]byte, n)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		result[i] = digits[num.Int64()]
	}
	return string(result)
}

func nullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
