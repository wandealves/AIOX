package ws

import (
	"context"
	"log/slog"
	"sync"

	"github.com/aiox-platform/aiox/internal/metrics"
)

// Hub manages WebSocket connections per user.
type Hub struct {
	mu    sync.RWMutex
	conns map[string]map[string]*Conn // userID → connID → *Conn

	registerCh   chan *Conn
	unregisterCh chan *Conn
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		conns:        make(map[string]map[string]*Conn),
		registerCh:   make(chan *Conn, 64),
		unregisterCh: make(chan *Conn, 64),
	}
}

// Run starts the hub event loop. It should be run as a goroutine.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.closeAll()
			return
		case conn := <-h.registerCh:
			h.addConn(conn)
		case conn := <-h.unregisterCh:
			h.removeConn(conn)
		}
	}
}

// Register adds a connection to the hub.
func (h *Hub) Register(conn *Conn) {
	h.registerCh <- conn
}

// Unregister removes a connection from the hub.
func (h *Hub) Unregister(conn *Conn) {
	h.unregisterCh <- conn
}

// SendToUser sends a message to all connections of a given user.
func (h *Hub) SendToUser(userID string, data []byte) {
	h.mu.RLock()
	userConns, ok := h.conns[userID]
	if !ok {
		h.mu.RUnlock()
		return
	}
	// Copy to avoid holding lock during writes
	conns := make([]*Conn, 0, len(userConns))
	for _, c := range userConns {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		select {
		case <-c.done:
			// Connection already removed, skip
		case c.send <- data:
			metrics.WSMessagesSentTotal.Inc()
		default:
			slog.Warn("ws: send buffer full, dropping message", "user_id", userID, "conn_id", c.ID)
		}
	}
}

// ConnectedCount returns the total number of active connections.
func (h *Hub) ConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, userConns := range h.conns {
		count += len(userConns)
	}
	return count
}

func (h *Hub) addConn(conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.conns[conn.UserID]; !ok {
		h.conns[conn.UserID] = make(map[string]*Conn)
	}
	h.conns[conn.UserID][conn.ID] = conn
	metrics.WSConnectionsActive.Inc()

	slog.Debug("ws: connection registered", "user_id", conn.UserID, "conn_id", conn.ID)
}

func (h *Hub) removeConn(conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if userConns, ok := h.conns[conn.UserID]; ok {
		if _, exists := userConns[conn.ID]; exists {
			delete(userConns, conn.ID)
			close(conn.done)
			metrics.WSConnectionsActive.Dec()
			if len(userConns) == 0 {
				delete(h.conns, conn.UserID)
			}
			slog.Debug("ws: connection unregistered", "user_id", conn.UserID, "conn_id", conn.ID)
		}
	}
}

func (h *Hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for userID, userConns := range h.conns {
		for connID, c := range userConns {
			close(c.done)
			delete(userConns, connID)
		}
		delete(h.conns, userID)
	}
}
