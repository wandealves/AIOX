package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// Publisher provides typed methods for publishing events to NATS JetStream.
type Publisher struct {
	js jetstream.JetStream
}

// NewPublisher creates a new Publisher.
func NewPublisher(js jetstream.JetStream) *Publisher {
	return &Publisher{js: js}
}

// PublishInboundMessage publishes an inbound XMPP message for orchestrator processing.
func (p *Publisher) PublishInboundMessage(ctx context.Context, msg InboundMessage) error {
	return p.publish(ctx, SubjectInboundMessage, msg)
}

// PublishOutboundMessage publishes an outbound message for XMPP delivery.
func (p *Publisher) PublishOutboundMessage(ctx context.Context, msg OutboundMessage) error {
	return p.publish(ctx, SubjectOutboundMessage, msg)
}

// PublishTask publishes a task for a specific agent (future Python worker processing).
func (p *Publisher) PublishTask(ctx context.Context, agentID string, msg TaskMessage) error {
	subject := fmt.Sprintf("%s.%s", SubjectTaskPrefix, agentID)
	return p.publish(ctx, subject, msg)
}

// PublishAgentEvent publishes an agent lifecycle event.
func (p *Publisher) PublishAgentEvent(ctx context.Context, event AgentEvent) error {
	return p.publish(ctx, SubjectAgentEvent, event)
}

// PublishAuditEvent publishes an audit event.
func (p *Publisher) PublishAuditEvent(ctx context.Context, event AuditEvent) error {
	return p.publish(ctx, SubjectAuditEvent, event)
}

func (p *Publisher) publish(ctx context.Context, subject string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling event for %s: %w", subject, err)
	}
	_, err = p.js.Publish(ctx, subject, payload)
	if err != nil {
		return fmt.Errorf("publishing to %s: %w", subject, err)
	}
	return nil
}
