package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rishabh-sonic/orbit/internal/middleware"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	rdb := newTestRedis(t)
	h := middleware.RateLimit(rdb, 5, time.Minute)(okHandler())

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: got %d, want 200", i+1, rr.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	rdb := newTestRedis(t)
	h := middleware.RateLimit(rdb, 3, time.Minute)(okHandler())

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d should pass: got %d", i+1, rr.Code)
		}
	}
	// 4th request should be blocked
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("4th request: got %d, want 429", rr.Code)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429")
	}
}

func TestRateLimit_DifferentIPsAreIndependent(t *testing.T) {
	rdb := newTestRedis(t)
	h := middleware.RateLimit(rdb, 2, time.Minute)(okHandler())

	// IP 1: hit limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "1.1.1.1:1234"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "1.1.1.1:1234"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("IP 1 over limit: got %d, want 429", rr.Code)
	}

	// IP 2: fresh window — should still pass
	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.RemoteAddr = "2.2.2.2:1234"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("IP 2 first request: got %d, want 200", rr2.Code)
	}
}

func TestRateLimit_XRealIPTakesPriority(t *testing.T) {
	rdb := newTestRedis(t)
	h := middleware.RateLimit(rdb, 1, time.Minute)(okHandler())

	// First request from X-Real-IP passes
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req1.Header.Set("X-Real-IP", "5.5.5.5")
	req1.RemoteAddr = "proxy:1234"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("first request: got %d, want 200", rr1.Code)
	}

	// Second request from same X-Real-IP (different RemoteAddr) is rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.Header.Set("X-Real-IP", "5.5.5.5")
	req2.RemoteAddr = "proxy2:1234"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request same X-Real-IP: got %d, want 429", rr2.Code)
	}
}
