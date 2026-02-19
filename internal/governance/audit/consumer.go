package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"

	inats "github.com/aiox-platform/aiox/internal/nats"
)

// Consumer listens on the audit event NATS subject and persists entries to the database.
type Consumer struct {
	repo        *Repository
	consumerMgr *inats.ConsumerManager
}

// NewConsumer creates a new audit event Consumer.
func NewConsumer(repo *Repository, consumerMgr *inats.ConsumerManager) *Consumer {
	return &Consumer{
		repo:        repo,
		consumerMgr: consumerMgr,
	}
}

// Start begins the consume loop. Blocks until ctx is cancelled.
func (c *Consumer) Start(ctx context.Context) error {
	consumer, err := c.consumerMgr.EnsureConsumer(ctx, inats.StreamEvents, "audit-persister", inats.SubjectAuditEvent)
	if err != nil {
		return err
	}

	slog.Info("audit consumer started", "consumer", "audit-persister")

	for {
		msgs, err := consumer.Fetch(10, jetstream.FetchMaxWait(inats.FetchTimeout))
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Debug("audit consumer: fetching events", "error", err)
			continue
		}

		for msg := range msgs.Messages() {
			c.handleEvent(ctx, msg)
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}

func (c *Consumer) handleEvent(ctx context.Context, msg jetstream.Msg) {
	var event inats.AuditEvent
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		slog.Error("audit consumer: unmarshaling event", "error", err)
		_ = msg.Nak()
		return
	}

	// Convert NATS AuditEvent to database AuditLog
	log := &AuditLog{
		ID:           uuid.New(),
		OwnerUserID:  event.OwnerUserID,
		EventType:    event.EventType,
		Severity:     event.Severity,
		ResourceType: event.ResourceType,
		CreatedAt:    event.Timestamp,
	}

	// Parse ResourceID — it may be a non-UUID string; use nil on failure
	if event.ResourceID != "" {
		if parsed, err := uuid.Parse(event.ResourceID); err == nil {
			log.ResourceID = &parsed
		}
	}

	// Store Details as JSONB {"message": "..."}
	detailsMap := map[string]string{"message": event.Details}
	if data, err := json.Marshal(detailsMap); err == nil {
		log.Details = data
	}

	if err := c.repo.Insert(ctx, log); err != nil {
		slog.Error("audit consumer: persisting audit log", "error", err, "event_type", event.EventType)
		_ = msg.Nak()
		return
	}

	_ = msg.Ack()

	slog.Debug("audit consumer: persisted event",
		"event_type", event.EventType,
		"owner", event.OwnerUserID,
		"resource_id", event.ResourceID,
	)
}
