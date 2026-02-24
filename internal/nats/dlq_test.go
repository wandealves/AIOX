package nats

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDLQMessage_Serialization(t *testing.T) {
	msg := DLQMessage{
		OriginalSubject: "aiox.tasks.abc-123",
		OriginalData:    json.RawMessage(`{"request_id":"req-1"}`),
		FailureReason:   "max delivery attempts exceeded",
		Attempts:        5,
		FailedAt:        time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded DLQMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.OriginalSubject != msg.OriginalSubject {
		t.Errorf("subject mismatch: got %s, want %s", decoded.OriginalSubject, msg.OriginalSubject)
	}
	if decoded.FailureReason != msg.FailureReason {
		t.Errorf("reason mismatch: got %s, want %s", decoded.FailureReason, msg.FailureReason)
	}
	if decoded.Attempts != msg.Attempts {
		t.Errorf("attempts mismatch: got %d, want %d", decoded.Attempts, msg.Attempts)
	}
}

func TestDLQSubjectForStream(t *testing.T) {
	tests := []struct {
		category string
		want     string
	}{
		{"tasks", "aiox.dlq.tasks"},
		{"events", "aiox.dlq.events"},
	}

	for _, tt := range tests {
		got := DLQSubjectForStream(tt.category)
		if got != tt.want {
			t.Errorf("DLQSubjectForStream(%s) = %s, want %s", tt.category, got, tt.want)
		}
	}
}
