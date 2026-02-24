package tracing

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestNATSHeaderCarrier_SetGet(t *testing.T) {
	h := make(nats.Header)
	carrier := NATSHeaderCarrier(h)

	carrier.Set("traceparent", "00-abc-def-01")
	if got := carrier.Get("traceparent"); got != "00-abc-def-01" {
		t.Errorf("expected '00-abc-def-01', got '%s'", got)
	}
}

func TestNATSHeaderCarrier_Keys(t *testing.T) {
	h := make(nats.Header)
	carrier := NATSHeaderCarrier(h)
	carrier.Set("key1", "val1")
	carrier.Set("key2", "val2")

	keys := carrier.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestInjectContext_SetsHeaders(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	defer tp.Shutdown(context.Background())
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	ctx, span := tp.Tracer("test").Start(context.Background(), "test-span")
	defer span.End()

	msg := &nats.Msg{Subject: "test"}
	InjectContext(ctx, msg)

	if msg.Header == nil {
		t.Fatal("expected headers to be set")
	}
	// W3C TraceContext propagator uses "traceparent" (lowercase)
	found := false
	for k := range msg.Header {
		if k == "Traceparent" || k == "traceparent" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected traceparent header to be set, got headers: %v", msg.Header)
	}
}

func TestInit_DisabledReturnsNil(t *testing.T) {
	tp, err := Init(context.Background(), TracingConfig{Enabled: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tp != nil {
		t.Error("expected nil tracer provider when disabled")
	}

	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test")
	if span.SpanContext().IsValid() {
		t.Error("expected invalid span context from no-op tracer")
	}
}

func TestInjectExtract_Roundtrip(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	defer tp.Shutdown(context.Background())
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	ctx, span := tp.Tracer("test").Start(context.Background(), "roundtrip")
	defer span.End()
	originalTraceID := span.SpanContext().TraceID()

	msg := &nats.Msg{Subject: "test"}
	InjectContext(ctx, msg)

	mock := &mockJSMsg{headers: nats.Header(msg.Header)}

	extractedCtx := ExtractContext(context.Background(), mock)
	extractedSpan := trace.SpanFromContext(extractedCtx)

	if extractedSpan.SpanContext().TraceID() != originalTraceID {
		t.Errorf("trace ID mismatch: got %s, want %s",
			extractedSpan.SpanContext().TraceID(), originalTraceID)
	}
}

// mockJSMsg implements the subset of jetstream.Msg needed for ExtractContext.
type mockJSMsg struct {
	headers nats.Header
}

func (m *mockJSMsg) Headers() nats.Header                  { return m.headers }
func (m *mockJSMsg) Data() []byte                           { return nil }
func (m *mockJSMsg) Subject() string                        { return "" }
func (m *mockJSMsg) Reply() string                          { return "" }
func (m *mockJSMsg) Ack() error                             { return nil }
func (m *mockJSMsg) Nak() error                             { return nil }
func (m *mockJSMsg) NakWithDelay(d time.Duration) error      { return nil }
func (m *mockJSMsg) InProgress() error                      { return nil }
func (m *mockJSMsg) Term() error                            { return nil }
func (m *mockJSMsg) TermWithReason(reason string) error     { return nil }
func (m *mockJSMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *mockJSMsg) DoubleAck(ctx context.Context) error    { return nil }
