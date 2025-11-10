package misc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/pkg/config"
)

type Handler struct {
	q   db.Querier
	rdb *redis.Client
	cfg *config.Config
}

func NewHandler(q db.Querier, rdb *redis.Client, cfg *config.Config) *Handler {
	return &Handler{q: q, rdb: rdb, cfg: cfg}
}

// GET /api/hello
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	middleware.Ok(w, map[string]string{"status": "ok", "service": "orbit"})
}

// GET /api/config
func (h *Handler) PublicConfig(w http.ResponseWriter, r *http.Request) {
	middleware.Ok(w, map[string]any{
		"site_name":        h.cfg.SiteName,
		"site_description": h.cfg.SiteDescription,
	})
}

// POST /api/online/heartbeat
func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims != nil {
		// Track online users in Redis with 5-minute TTL
		key := fmt.Sprintf("online:%s", claims.UserID)
		h.rdb.Set(context.Background(), key, 1, 5*time.Minute)

		// Record daily visit
		_ = h.q.RecordUserVisit(r.Context(), claims.UserID)
	}
	// Also count anonymous heartbeats using a sorted set
	h.rdb.ZAdd(context.Background(), "online:set", redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: r.RemoteAddr,
	})
	// Remove stale entries (older than 5 minutes)
	cutoff := float64(time.Now().Add(-5 * time.Minute).Unix())
	h.rdb.ZRemRangeByScore(context.Background(), "online:set", "-inf", fmt.Sprintf("%f", cutoff))

	middleware.Ok(w, map[string]bool{"ok": true})
}

// GET /api/online/count
func (h *Handler) OnlineCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.rdb.ZCard(context.Background(), "online:set").Result()
	if err != nil {
		count = 0
	}
	middleware.Ok(w, map[string]int64{"count": count})
}
