package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimit returns a middleware that limits requests to `limit` per `window` per IP.
// It uses Redis sliding-window counters.
func RateLimit(rdb *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			key := fmt.Sprintf("rl:%s:%s", r.URL.Path, ip)

			ctx := context.Background()
			count, err := rdb.Incr(ctx, key).Result()
			if err == nil && count == 1 {
				rdb.Expire(ctx, key, window)
			}

			if err == nil && count > int64(limit) {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				JSON(w, http.StatusTooManyRequests, APIResponse{Error: "rate limit exceeded"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
