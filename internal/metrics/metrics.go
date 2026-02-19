package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aiox_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aiox_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	TasksDispatchedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "aiox_tasks_dispatched_total",
			Help: "Total number of tasks dispatched to workers.",
		},
	)

	TasksCompletedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aiox_tasks_completed_total",
			Help: "Total number of tasks completed by workers.",
		},
		[]string{"status"},
	)

	WorkerPoolConnected = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "aiox_worker_pool_connected",
			Help: "Number of connected gRPC workers.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		TasksDispatchedTotal,
		TasksCompletedTotal,
		WorkerPoolConnected,
	)
}
