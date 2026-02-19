package xmpp

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
	"gosrc.io/xmpp"

	inats "github.com/aiox-platform/aiox/internal/nats"
)

// OutboundRelay consumes outbound messages from NATS and sends them via XMPP.
type OutboundRelay struct {
	handler    *Handler
	sender     xmpp.Sender
	consumerMgr *inats.ConsumerManager
}

// NewOutboundRelay creates a new OutboundRelay.
func NewOutboundRelay(handler *Handler, sender xmpp.Sender, consumerMgr *inats.ConsumerManager) *OutboundRelay {
	return &OutboundRelay{
		handler:    handler,
		sender:     sender,
		consumerMgr: consumerMgr,
	}
}

// Start begins consuming outbound messages and sending them via XMPP.
func (r *OutboundRelay) Start(ctx context.Context) error {
	consumer, err := r.consumerMgr.EnsureConsumer(ctx, inats.StreamMessages, "outbound-relay", inats.SubjectOutboundMessage)
	if err != nil {
		return err
	}

	slog.Info("outbound relay started", "consumer", "outbound-relay")

	for {
		msgs, err := consumer.Fetch(10, jetstream.FetchMaxWait(inats.FetchTimeout))
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Debug("fetching outbound messages", "error", err)
			continue
		}

		for msg := range msgs.Messages() {
			var outbound inats.OutboundMessage
			if err := json.Unmarshal(msg.Data(), &outbound); err != nil {
				slog.Error("unmarshaling outbound message", "error", err)
				_ = msg.Nak()
				continue
			}

			if err := r.handler.SendOutboundMessage(r.sender, outbound); err != nil {
				slog.Error("sending outbound XMPP message", "error", err, "to", outbound.ToJID)
				_ = msg.Nak()
				continue
			}

			slog.Debug("sent outbound XMPP message", "to", outbound.ToJID, "from", outbound.FromJID)
			_ = msg.Ack()
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}
