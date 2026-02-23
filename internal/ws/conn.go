package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"nhooyr.io/websocket"

	"github.com/aiox-platform/aiox/internal/metrics"
	inats "github.com/aiox-platform/aiox/internal/nats"
)

const (
	writeWait  = 10 * time.Second
	sendBufLen = 256
)

// ClientMessage is the JSON payload sent by WebSocket clients.
type ClientMessage struct {
	Type    string `json:"type"`
	AgentID string `json:"agent_id"`
	Body    string `json:"body"`
}

// ServerMessage is the JSON payload sent to WebSocket clients.
type ServerMessage struct {
	Type      string `json:"type"`
	AgentID   string `json:"agent_id,omitempty"`
	Body      string `json:"body,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Conn wraps a single WebSocket connection.
type Conn struct {
	ID     string
	UserID string
	ws     *websocket.Conn
	hub    *Hub
	send   chan []byte
	done   chan struct{} // closed when conn is removed from hub
}

// NewConn creates a Conn wrapper.
func NewConn(id, userID string, ws *websocket.Conn, hub *Hub) *Conn {
	return &Conn{
		ID:     id,
		UserID: userID,
		ws:     ws,
		hub:    hub,
		send:   make(chan []byte, sendBufLen),
		done:   make(chan struct{}),
	}
}

// WritePump writes messages from the send channel to the WebSocket.
func (c *Conn) WritePump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
		c.ws.Close(websocket.StatusNormalClosure, "closing")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		case msg := <-c.send:
			writeCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.ws.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				slog.Debug("ws: write error", "conn_id", c.ID, "error", err)
				return
			}
		}
	}
}

// ReadPump reads messages from the WebSocket and publishes to NATS.
func (c *Conn) ReadPump(ctx context.Context, publisher *inats.Publisher) {
	defer func() {
		c.hub.Unregister(c)
		c.ws.Close(websocket.StatusNormalClosure, "closing")
	}()

	for {
		_, data, err := c.ws.Read(ctx)
		if err != nil {
			var ce websocket.CloseError
			if errors.As(err, &ce) && ce.Code == websocket.StatusNormalClosure {
				slog.Debug("ws: connection closed normally", "conn_id", c.ID)
			} else {
				slog.Debug("ws: read error", "conn_id", c.ID, "error", err)
			}
			return
		}

		metrics.WSMessagesReceivedTotal.Inc()

		var cm ClientMessage
		if err := json.Unmarshal(data, &cm); err != nil {
			slog.Debug("ws: invalid client message", "conn_id", c.ID, "error", err)
			c.sendError("invalid message format")
			continue
		}

		if cm.Type != "message" {
			c.sendError("unsupported message type")
			continue
		}

		if cm.AgentID == "" || cm.Body == "" {
			c.sendError("agent_id and body are required")
			continue
		}

		requestID := c.ID + "-" + time.Now().Format("20060102150405.000")
		fromJID := "ws-" + c.UserID + "@ws.local"

		inbound := inats.InboundMessage{
			ID:         requestID,
			FromJID:    fromJID,
			ToJID:      "agent-" + cm.AgentID + "@agents.local",
			Body:       cm.Body,
			StanzaType: "chat",
			ReceivedAt: time.Now().UTC(),
			Channel:    inats.ChannelWS,
		}

		if err := publisher.PublishInboundMessage(ctx, inbound); err != nil {
			slog.Error("ws: publishing inbound message", "error", err)
			c.sendError("failed to process message")
		}
	}
}

func (c *Conn) sendError(msg string) {
	resp := ServerMessage{Type: "error", Error: msg}
	data, _ := json.Marshal(resp)
	select {
	case c.send <- data:
	default:
	}
}
