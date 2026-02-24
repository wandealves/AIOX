package worker

import (
	"sync"
	"time"

	"github.com/aiox-platform/aiox/internal/metrics"
)

// CircuitState represents the current state of the circuit breaker.
type CircuitState int

const (
	StateClosed   CircuitState = 0
	StateOpen     CircuitState = 1
	StateHalfOpen CircuitState = 2
)

// CircuitBreaker implements the circuit breaker pattern for gRPC worker connections.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failures     int
	threshold    int
	timeout      time.Duration
	lastFailure  time.Time
	halfOpenUsed bool
}

// NewCircuitBreaker creates a new CircuitBreaker.
// threshold is the number of consecutive failures before opening.
// timeout is how long to wait before transitioning from Open to HalfOpen.
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	cb := &CircuitBreaker{
		state:     StateClosed,
		threshold: threshold,
		timeout:   timeout,
	}
	metrics.CircuitBreakerState.Set(0)
	return cb
}

// Allow returns true if a request is allowed through the circuit breaker.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = StateHalfOpen
			cb.halfOpenUsed = false
			metrics.CircuitBreakerState.Set(2)
		} else {
			return false
		}
		fallthrough
	case StateHalfOpen:
		if cb.halfOpenUsed {
			return false
		}
		cb.halfOpenUsed = true
		return true
	}
	return false
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	if cb.state != StateClosed {
		cb.state = StateClosed
		metrics.CircuitBreakerState.Set(0)
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.threshold {
			cb.state = StateOpen
			metrics.CircuitBreakerState.Set(1)
			metrics.CircuitBreakerTripsTotal.Inc()
		}
	case StateHalfOpen:
		cb.state = StateOpen
		metrics.CircuitBreakerState.Set(1)
		metrics.CircuitBreakerTripsTotal.Inc()
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
