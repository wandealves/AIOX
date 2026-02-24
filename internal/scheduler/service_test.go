package scheduler

import (
	"testing"
	"time"
)

func TestComputeNextRun_Valid(t *testing.T) {
	// Every minute
	now := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	next, err := ComputeNextRun("* * * * *", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2026, 1, 15, 10, 31, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestComputeNextRun_Hourly(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	next, err := ComputeNextRun("0 * * * *", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestComputeNextRun_Daily(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	next, err := ComputeNextRun("0 9 * * *", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Next 9:00 is tomorrow
	expected := time.Date(2026, 1, 16, 9, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestComputeNextRun_InvalidExpression(t *testing.T) {
	_, err := ComputeNextRun("not a cron", time.Now())
	if err == nil {
		t.Error("expected error for invalid cron expression")
	}
}

func TestComputeNextRun_Weekly(t *testing.T) {
	// Every Monday at 8:00
	now := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC) // Monday
	next, err := ComputeNextRun("0 8 * * 1", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Next Monday 8:00 is Feb 2
	expected := time.Date(2026, 3, 2, 8, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}
