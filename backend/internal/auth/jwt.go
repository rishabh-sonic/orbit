package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/rishabh-sonic/orbit/pkg/token"
)

// JWTService is an alias for token.Service so auth code can use auth.JWTService.
type JWTService = token.Service

func NewJWTService(secret, resetSecret string, expiration time.Duration) *JWTService {
	return token.NewService(secret, resetSecret, expiration)
}

// Claims is re-exported from pkg/token.
type Claims = token.Claims

// GenerateToken is a convenience wrapper.
func GenerateToken(svc *JWTService, userID uuid.UUID, username, role string) (string, error) {
	return svc.GenerateToken(userID, username, role)
}
