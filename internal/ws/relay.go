package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	inats "github.com/aiox-platform/aiox/internal/nats"
)

// Relay consumes outbound WebSocket messages from NATS and delivers them via the Hub.
type Relay struct {
	hub         *Hub
	consumerMgr *inats.ConsumerManager
}

// NewRelay creates a new WebSocket relay.
func NewRelay(hub *Hub, consumerMgr *inats.ConsumerManager) *Relay {
	return &Relay{hub: hub, consumerMgr: consumerMgr}
}

// Start begins the relay consumer loop.
func (r *Relay) Start(ctx context.Context) error {
	consumer, err := r.consumerMgr.EnsureConsumer(ctx, inats.StreamMessages, "ws-outbound-relay", inats.SubjectOutboundMessageWS)
	if err != nil {
		return err
	}

	slog.Info("ws outbound relay started", "subject", inats.SubjectOutboundMessageWS)

	for {
		msgs, err := consumer.Fetch(10, jetstream.FetchMaxWait(inats.FetchTimeout))
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Debug("ws relay: fetching messages", "error", err)
			continue
		}

		for msg := range msgs.Messages() {
			r.handleMessage(msg)
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}

func (r *Relay) handleMessage(msg jetstream.Msg) {
	var outbound inats.OutboundMessage
	if err := json.Unmarshal(msg.Data(), &outbound); err != nil {
		slog.Error("ws relay: unmarshaling outbound message", "error", err)
		_ = msg.Nak()
		return
	}

	// Extract userID from ToJID format "ws-<userID>@ws.local"
	userID := extractUserIDFromJID(outbound.ToJID)
	if userID == "" {
		slog.Warn("ws relay: could not extract user ID from JID", "to_jid", outbound.ToJID)
		_ = msg.Ack()
		return
	}

	// Extract agent ID from FromJID format "agent-<uuid>@agents.local"
	agentID := extractAgentIDFromJID(outbound.FromJID)

	payload := ServerMessage{
		Type:      "message",
		AgentID:   agentID,
		Body:      outbound.Body,
		RequestID: outbound.InReplyTo,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("ws relay: marshaling server message", "error", err)
		_ = msg.Nak()
		return
	}

	r.hub.SendToUser(userID, data)

	_ = msg.Ack()
}

// extractUserIDFromJID parses "ws-<userID>@ws.local" to extract userID.
func extractUserIDFromJID(jid string) string {
	// Remove resource if present
	if idx := strings.Index(jid, "/"); idx >= 0 {
		jid = jid[:idx]
	}

	atIdx := strings.Index(jid, "@")
	if atIdx < 0 {
		return ""
	}

	local := jid[:atIdx]
	if !strings.HasPrefix(local, "ws-") {
		return ""
	}

	return local[3:]
}

// extractAgentIDFromJID parses "agent-<uuid>@agents.local" to extract the UUID.
func extractAgentIDFromJID(jid string) string {
	if idx := strings.Index(jid, "/"); idx >= 0 {
		jid = jid[:idx]
	}

	atIdx := strings.Index(jid, "@")
	if atIdx < 0 {
		return ""
	}

	local := jid[:atIdx]
	if strings.HasPrefix(local, "agent-") {
		return local[6:]
	}

	return local
}
