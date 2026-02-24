package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", id)

		// Add trace ID to response headers if a span is active
		span := trace.SpanFromContext(r.Context())
		if span.SpanContext().IsValid() {
			w.Header().Set("X-Trace-ID", span.SpanContext().TraceID().String())
		}

		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
