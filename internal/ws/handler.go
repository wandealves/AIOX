package ws

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"nhooyr.io/websocket"

	"github.com/aiox-platform/aiox/internal/auth"
	inats "github.com/aiox-platform/aiox/internal/nats"
)

// Handler handles WebSocket HTTP upgrade requests.
type Handler struct {
	hub       *Hub
	jwtMgr    *auth.JWTManager
	publisher *inats.Publisher
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub, jwtMgr *auth.JWTManager, publisher *inats.Publisher) *Handler {
	return &Handler{hub: hub, jwtMgr: jwtMgr, publisher: publisher}
}

// Upgrade handles GET /api/v1/ws?token=<JWT>.
// Auth is done via query param since browsers can't set headers on WebSocket handshake.
func (h *Handler) Upgrade(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwtMgr.ValidateAccessToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // CORS handled by middleware
	})
	if err != nil {
		slog.Error("ws: accept error", "error", err)
		return
	}

	connID := uuid.New().String()
	conn := NewConn(connID, claims.UserID, wsConn, h.hub)

	h.hub.Register(conn)

	slog.Info("ws: new connection", "user_id", claims.UserID, "conn_id", connID)

	// Start read/write pumps — both will call Unregister on exit
	go conn.WritePump(r.Context())
	conn.ReadPump(r.Context(), h.publisher)
}
