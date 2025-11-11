package wsHub

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

// Hub manages per-user WebSocket connections.
type Hub struct {
	mu    sync.RWMutex
	conns map[uuid.UUID][]*websocket.Conn
}

func New() *Hub {
	return &Hub{conns: make(map[uuid.UUID][]*websocket.Conn)}
}

func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = append(h.conns[userID], conn)
}

func (h *Hub) Unregister(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.conns[userID]
	for i, c := range conns {
		if c == conn {
			h.conns[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

// Send writes a JSON message to all connections for the given user.
func (h *Hub) Send(userID uuid.UUID, payload []byte) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, len(h.conns[userID]))
	copy(conns, h.conns[userID])
	h.mu.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			slog.Error("ws write", "user", userID, "err", err)
		}
	}
}
