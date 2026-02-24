package tracing

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
)

// NATSHeaderCarrier adapts nats.Header for OpenTelemetry propagation.
type NATSHeaderCarrier nats.Header

// Get returns the value for a key.
func (c NATSHeaderCarrier) Get(key string) string {
	return nats.Header(c).Get(key)
}

// Set sets a key-value pair.
func (c NATSHeaderCarrier) Set(key, value string) {
	nats.Header(c).Set(key, value)
}

// Keys returns all keys.
func (c NATSHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// InjectContext injects the span context from ctx into a NATS message's headers.
func InjectContext(ctx context.Context, msg *nats.Msg) {
	if msg.Header == nil {
		msg.Header = make(nats.Header)
	}
	otel.GetTextMapPropagator().Inject(ctx, NATSHeaderCarrier(msg.Header))
}

// ExtractContext extracts a span context from a JetStream message's headers.
func ExtractContext(ctx context.Context, msg jetstream.Msg) context.Context {
	headers := msg.Headers()
	if headers == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, NATSHeaderCarrier(headers))
}
