package orchestrator

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"

	inats "github.com/aiox-platform/aiox/internal/nats"
)

// Orchestrator consumes inbound messages, validates ownership, routes them,
// and publishes tasks and outbound responses.
type Orchestrator struct {
	publisher   *inats.Publisher
	consumerMgr *inats.ConsumerManager
	validator   *Validator
	router      *Router
}

// NewOrchestrator creates a new Orchestrator.
func NewOrchestrator(
	publisher *inats.Publisher,
	consumerMgr *inats.ConsumerManager,
	validator *Validator,
	router *Router,
) *Orchestrator {
	return &Orchestrator{
		publisher:   publisher,
		consumerMgr: consumerMgr,
		validator:   validator,
		router:      router,
	}
}

// Start begins the orchestrator event loop.
func (o *Orchestrator) Start(ctx context.Context) error {
	consumer, err := o.consumerMgr.EnsureConsumer(ctx, inats.StreamMessages, "orchestrator", inats.SubjectInboundMessage)
	if err != nil {
		return err
	}

	slog.Info("orchestrator started", "consumer", "orchestrator")

	for {
		msgs, err := consumer.Fetch(10, jetstream.FetchMaxWait(inats.FetchTimeout))
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Debug("fetching inbound messages", "error", err)
			continue
		}

		for msg := range msgs.Messages() {
			o.processMessage(ctx, msg)
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}

func (o *Orchestrator) processMessage(ctx context.Context, msg jetstream.Msg) {
	var inbound inats.InboundMessage
	if err := json.Unmarshal(msg.Data(), &inbound); err != nil {
		slog.Error("unmarshaling inbound message", "error", err)
		_ = msg.Nak()
		return
	}

	slog.Debug("orchestrator processing message",
		"id", inbound.ID,
		"from", inbound.FromJID,
		"to", inbound.ToJID,
	)

	// Route: resolve target agent from JID
	route, err := o.router.Route(ctx, inbound.ToJID)
	if err != nil {
		slog.Warn("routing failed", "error", err, "to_jid", inbound.ToJID)
		o.sendErrorResponse(ctx, inbound, "Agent not found")
		_ = msg.Ack()
		return
	}

	// Validate ownership and governance
	if err := o.validator.Validate(route); err != nil {
		slog.Warn("validation failed", "error", err, "agent_id", route.AgentID)
		o.sendErrorResponse(ctx, inbound, "Message not authorized")
		_ = msg.Ack()
		return
	}

	// Publish task for future Python worker processing (Phase 3)
	task := inats.TaskMessage{
		RequestID:   inbound.ID,
		AgentID:     route.AgentID,
		OwnerUserID: route.OwnerUserID,
		Message:     inbound.Body,
		FromJID:     inbound.FromJID,
	}
	if err := o.publisher.PublishTask(ctx, route.AgentID.String(), task); err != nil {
		slog.Error("publishing task", "error", err)
	}

	// Phase 2 placeholder response — Phase 3 replaces this with AI worker output
	outbound := inats.OutboundMessage{
		ID:        uuid.New().String(),
		ToJID:     inbound.FromJID,
		FromJID:   route.AgentJID,
		Body:      "[" + route.AgentName + "] Message received. AI processing will be available in Phase 3.",
		InReplyTo: inbound.ID,
	}
	if err := o.publisher.PublishOutboundMessage(ctx, outbound); err != nil {
		slog.Error("publishing outbound message", "error", err)
	}

	// Publish audit event
	audit := inats.AuditEvent{
		OwnerUserID:  route.OwnerUserID,
		EventType:    "message_routed",
		Severity:     "info",
		ResourceType: "agent",
		ResourceID:   route.AgentID.String(),
		Details:      "Message routed from " + inbound.FromJID,
		Timestamp:    time.Now().UTC(),
	}
	if err := o.publisher.PublishAuditEvent(ctx, audit); err != nil {
		slog.Error("publishing audit event", "error", err)
	}

	_ = msg.Ack()
}

func (o *Orchestrator) sendErrorResponse(ctx context.Context, inbound inats.InboundMessage, errMsg string) {
	outbound := inats.OutboundMessage{
		ID:        uuid.New().String(),
		ToJID:     inbound.FromJID,
		FromJID:   inbound.ToJID,
		Body:      "Error: " + errMsg,
		InReplyTo: inbound.ID,
	}
	if err := o.publisher.PublishOutboundMessage(ctx, outbound); err != nil {
		slog.Error("publishing error response", "error", err)
	}
}
