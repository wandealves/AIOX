package xmpp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"

	inats "github.com/aiox-platform/aiox/internal/nats"
)

// Handler processes incoming XMPP stanzas and bridges them to NATS.
type Handler struct {
	publisher *inats.Publisher
}

// NewHandler creates a new XMPP stanza handler.
func NewHandler(publisher *inats.Publisher) *Handler {
	return &Handler{publisher: publisher}
}

// HandleMessage processes incoming <message> stanzas and publishes them to NATS.
func (h *Handler) HandleMessage(s xmpp.Sender, p stanza.Packet) {
	msg, ok := p.(stanza.Message)
	if !ok {
		return
	}

	if msg.Body == "" {
		return
	}

	slog.Debug("XMPP message received",
		"from", msg.From,
		"to", msg.To,
		"type", string(msg.Type),
	)

	inbound := inats.InboundMessage{
		ID:         uuid.New().String(),
		FromJID:    msg.From,
		ToJID:      msg.To,
		Body:       msg.Body,
		StanzaType: string(msg.Type),
		ReceivedAt: time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.publisher.PublishInboundMessage(ctx, inbound); err != nil {
		slog.Error("publishing inbound message", "error", err, "from", msg.From)
		h.sendError(s, msg.From, msg.To, "Internal error processing your message")
		return
	}
}

// HandlePresence processes incoming <presence> stanzas, auto-approving subscribe requests.
func (h *Handler) HandlePresence(s xmpp.Sender, p stanza.Packet) {
	pres, ok := p.(stanza.Presence)
	if !ok {
		return
	}

	slog.Debug("XMPP presence received",
		"from", pres.From,
		"to", pres.To,
		"type", string(pres.Type),
	)

	if pres.Type == "subscribe" {
		reply := stanza.Presence{
			Attrs: stanza.Attrs{
				From: pres.To,
				To:   pres.From,
				Type: "subscribed",
			},
		}
		if err := s.Send(reply); err != nil {
			slog.Error("sending presence subscribed reply", "error", err)
		}
	}
}

// HandleIQ processes incoming <iq> stanzas.
func (h *Handler) HandleIQ(_ xmpp.Sender, p stanza.Packet) {
	iq, ok := p.(*stanza.IQ)
	if !ok {
		return
	}
	slog.Debug("XMPP IQ received", "from", iq.From, "to", iq.To, "type", string(iq.Type))
}

// SendOutboundMessage sends a <message> stanza via XMPP.
func (h *Handler) SendOutboundMessage(s xmpp.Sender, outbound inats.OutboundMessage) error {
	msg := stanza.Message{
		Attrs: stanza.Attrs{
			From: outbound.FromJID,
			To:   outbound.ToJID,
			Type: "chat",
			Id:   outbound.ID,
		},
		Body: outbound.Body,
	}
	return s.Send(msg)
}

func (h *Handler) sendError(s xmpp.Sender, to, from, body string) {
	msg := stanza.Message{
		Attrs: stanza.Attrs{
			From: from,
			To:   to,
			Type: "chat",
		},
		Body: body,
	}
	if err := s.Send(msg); err != nil {
		slog.Error("sending error message", "error", err)
	}
}

// ExtractAgentID parses an agent UUID from a JID like "agent-<uuid>@agents.domain".
func ExtractAgentID(jid string) (uuid.UUID, error) {
	// Strip resource part (e.g., /resource)
	bare := jid
	if idx := strings.Index(jid, "/"); idx >= 0 {
		bare = jid[:idx]
	}

	// Get local part before @
	local := bare
	if idx := strings.Index(bare, "@"); idx >= 0 {
		local = bare[:idx]
	}

	// Remove "agent-" prefix
	if !strings.HasPrefix(local, "agent-") {
		return uuid.Nil, fmt.Errorf("JID %q does not match agent-<uuid> format", jid)
	}

	idStr := strings.TrimPrefix(local, "agent-")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid agent UUID in JID %q: %w", jid, err)
	}
	return id, nil
}
