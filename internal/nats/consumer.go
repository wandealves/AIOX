package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// ConsumerManager handles durable consumer creation and retrieval.
type ConsumerManager struct {
	js jetstream.JetStream
}

// NewConsumerManager creates a new ConsumerManager.
func NewConsumerManager(js jetstream.JetStream) *ConsumerManager {
	return &ConsumerManager{js: js}
}

// EnsureConsumer creates or updates a durable consumer on the given stream.
func (cm *ConsumerManager) EnsureConsumer(ctx context.Context, stream, name, filterSubject string) (jetstream.Consumer, error) {
	cfg := jetstream.ConsumerConfig{
		Durable:       name,
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
	}

	consumer, err := cm.js.CreateOrUpdateConsumer(ctx, stream, cfg)
	if err != nil {
		return nil, fmt.Errorf("ensuring consumer %s on %s: %w", name, stream, err)
	}
	return consumer, nil
}
