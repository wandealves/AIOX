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

	WSConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "aiox_ws_connections_active",
			Help: "Number of active WebSocket connections.",
		},
	)

	WSMessagesReceivedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "aiox_ws_messages_received_total",
			Help: "Total number of messages received from WebSocket clients.",
		},
	)

	WSMessagesSentTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "aiox_ws_messages_sent_total",
			Help: "Total number of messages sent to WebSocket clients.",
		},
	)

	DLQMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aiox_dlq_messages_total",
			Help: "Total number of messages sent to the dead-letter queue.",
		},
		[]string{"source"},
	)

	CircuitBreakerState = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "aiox_circuit_breaker_state",
			Help: "Circuit breaker state: 0=closed, 1=open, 2=half-open.",
		},
	)

	CircuitBreakerTripsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "aiox_circuit_breaker_trips_total",
			Help: "Total number of times the circuit breaker tripped to open.",
		},
	)

	// Tool execution metrics (Phase 12)
	ToolExecutionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aiox_tool_executions_total",
			Help: "Total tool executions.",
		},
		[]string{"tool_name", "tool_type", "status"},
	)

	ToolExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aiox_tool_execution_duration_seconds",
			Help:    "Tool execution duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tool_name", "tool_type"},
	)

	ToolErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aiox_tool_errors_total",
			Help: "Total tool execution errors.",
		},
		[]string{"tool_name", "error_type"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		TasksDispatchedTotal,
		TasksCompletedTotal,
		WorkerPoolConnected,
		WSConnectionsActive,
		WSMessagesReceivedTotal,
		WSMessagesSentTotal,
		DLQMessagesTotal,
		CircuitBreakerState,
		CircuitBreakerTripsTotal,
		ToolExecutionsTotal,
		ToolExecutionDuration,
		ToolErrorsTotal,
	)
}
