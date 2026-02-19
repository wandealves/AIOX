package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/aiox-platform/aiox/internal/metrics"
)

// Metrics records HTTP request count and latency as Prometheus metrics.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		// Use the route pattern for a low-cardinality path label.
		path := "unknown"
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			if pat := rctx.RoutePattern(); pat != "" {
				path = pat
			}
		}

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(ww.status)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}
