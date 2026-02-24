package worker

import (
	"testing"
	"time"
)

func TestCircuitBreaker_StartsClosedAndAllows(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	if cb.State() != StateClosed {
		t.Errorf("expected closed, got %d", cb.State())
	}
	if !cb.Allow() {
		t.Error("expected allow in closed state")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateClosed {
		t.Error("expected closed after 2 failures")
	}

	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Error("expected open after 3 failures")
	}
	if cb.Allow() {
		t.Error("expected deny in open state")
	}
}

func TestCircuitBreaker_ResetOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()

	if cb.State() != StateClosed {
		t.Error("expected closed after success reset")
	}
	// Need 3 more failures to trip
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateClosed {
		t.Error("expected closed after 2 failures post-reset")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatal("expected open")
	}

	time.Sleep(60 * time.Millisecond)
	if !cb.Allow() {
		t.Error("expected allow in half-open state")
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("expected half-open, got %d", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenSuccessCloses(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)

	cb.Allow() // transitions to half-open
	cb.RecordSuccess()

	if cb.State() != StateClosed {
		t.Errorf("expected closed after half-open success, got %d", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)

	cb.Allow() // transitions to half-open
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("expected open after half-open failure, got %d", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenAllowsOnlyOneRequest(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)

	if !cb.Allow() {
		t.Error("first allow should succeed in half-open")
	}
	if cb.Allow() {
		t.Error("second allow should be denied in half-open")
	}
}

func TestCircuitBreaker_DefaultValues(t *testing.T) {
	cb := NewCircuitBreaker(0, 0)
	if cb.threshold != 5 {
		t.Errorf("expected default threshold 5, got %d", cb.threshold)
	}
	if cb.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cb.timeout)
	}
}
