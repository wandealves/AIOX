package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"request_id", r.Context().Value(requestIDKey),
		}

		span := trace.SpanFromContext(r.Context())
		if span.SpanContext().IsValid() {
			attrs = append(attrs,
				"trace_id", span.SpanContext().TraceID().String(),
				"span_id", span.SpanContext().SpanID().String(),
			)
		}

		slog.Info("request", attrs...)
	})
}
