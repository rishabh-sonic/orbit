package middleware

import (
	"context"
	"net/http"

	"github.com/rishabh-sonic/orbit/pkg/token"
)

// Authenticate validates the Bearer token and injects Claims into the context.
// It does NOT reject unauthenticated requests — use RequireAuth for that.
func Authenticate(jwt *token.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := token.ExtractBearerToken(r.Header.Get("Authorization"))
			if raw != "" {
				if claims, err := jwt.ValidateToken(raw); err == nil {
					ctx := context.WithValue(r.Context(), token.ClaimsKey, claims)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth rejects requests without a valid JWT.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ClaimsFromContext(r.Context()) == nil {
			Unauthorized(w, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin rejects requests whose JWT role is not ADMIN.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil {
			Unauthorized(w, "authentication required")
			return
		}
		if claims.Role != "ADMIN" {
			Forbidden(w, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ClaimsFromContext retrieves JWT claims from the request context.
func ClaimsFromContext(ctx context.Context) *token.Claims {
	return token.ClaimsFromContext(ctx)
}
