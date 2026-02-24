package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	natspkg "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/aiox-platform/aiox/internal/metrics"
	"github.com/aiox-platform/aiox/internal/tracing"
)

// DLQMessage represents a message that has been sent to the dead-letter queue.
type DLQMessage struct {
	OriginalSubject string          `json:"original_subject"`
	OriginalData    json.RawMessage `json:"original_data"`
	FailureReason   string          `json:"failure_reason"`
	Attempts        int             `json:"attempts"`
	FailedAt        time.Time       `json:"failed_at"`
}

// DLQPublisher publishes failed messages to the dead-letter queue.
type DLQPublisher struct {
	js jetstream.JetStream
}

// NewDLQPublisher creates a new DLQPublisher.
func NewDLQPublisher(js jetstream.JetStream) *DLQPublisher {
	return &DLQPublisher{js: js}
}

// PublishToDLQ routes a failed message to the appropriate DLQ subject.
func (d *DLQPublisher) PublishToDLQ(ctx context.Context, dlqSubject string, originalData []byte, reason string, attempts int) {
	dlqMsg := DLQMessage{
		OriginalSubject: dlqSubject,
		OriginalData:    originalData,
		FailureReason:   reason,
		Attempts:        attempts,
		FailedAt:        time.Now().UTC(),
	}

	payload, err := json.Marshal(dlqMsg)
	if err != nil {
		slog.Error("dlq: marshaling message", "error", err)
		return
	}

	msg := &natspkg.Msg{
		Subject: dlqSubject,
		Data:    payload,
	}
	tracing.InjectContext(ctx, msg)

	if _, err := d.js.PublishMsg(ctx, msg); err != nil {
		slog.Error("dlq: publishing message", "error", err, "subject", dlqSubject)
		return
	}

	// Extract source stream name from subject for metrics
	source := "unknown"
	if len(dlqSubject) > len("aiox.dlq.") {
		source = dlqSubject[len("aiox.dlq."):]
	}
	metrics.DLQMessagesTotal.WithLabelValues(source).Inc()

	slog.Warn("dlq: message routed to dead-letter queue",
		"subject", dlqSubject,
		"reason", reason,
		"attempts", attempts,
	)
}

// ShouldDLQ checks if a message has exceeded its max delivery attempts.
func ShouldDLQ(msg jetstream.Msg, maxDeliver int) (bool, int) {
	meta, err := msg.Metadata()
	if err != nil {
		return false, 0
	}
	numDelivered := int(meta.NumDelivered)
	return numDelivered >= maxDeliver, numDelivered
}

// DLQSubjectForStream returns the DLQ subject for a given source category.
func DLQSubjectForStream(category string) string {
	return fmt.Sprintf("aiox.dlq.%s", category)
}
