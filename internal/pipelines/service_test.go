package pipelines

import (
	"testing"
)

func TestApplyTransform(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple template",
			tmpl:     "Summarize this: {{.Input}}",
			input:    "Hello world",
			expected: "Summarize this: Hello world",
		},
		{
			name:     "no template vars",
			tmpl:     "Just a static prompt",
			input:    "anything",
			expected: "Just a static prompt",
		},
		{
			name:    "invalid template",
			tmpl:    "{{.Invalid",
			input:   "test",
			wantErr: true,
		},
		{
			name:     "multi-line input",
			tmpl:     "Translate to Spanish:\n{{.Input}}",
			input:    "Good morning\nHow are you?",
			expected: "Translate to Spanish:\nGood morning\nHow are you?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := applyTransform(tt.tmpl, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHandlePipelineResult_NoWaiter(t *testing.T) {
	svc := &Service{
		waiters: make(map[string]chan *stepResponse),
	}
	// Should not panic when no waiter exists
	svc.HandlePipelineResult("unknown-id", "text", 100, 500, "")
}

func TestHandlePipelineResult_WithWaiter(t *testing.T) {
	svc := &Service{
		waiters: make(map[string]chan *stepResponse),
	}

	ch := make(chan *stepResponse, 1)
	svc.waiters["req-123"] = ch

	svc.HandlePipelineResult("req-123", "result text", 42, 100, "")

	result := <-ch
	if result.Text != "result text" {
		t.Errorf("expected 'result text', got %q", result.Text)
	}
	if result.Tokens != 42 {
		t.Errorf("expected 42 tokens, got %d", result.Tokens)
	}
	if result.DurationMs != 100 {
		t.Errorf("expected 100 ms, got %d", result.DurationMs)
	}

	// Verify waiter was cleaned up
	if _, exists := svc.waiters["req-123"]; exists {
		t.Error("expected waiter to be cleaned up")
	}
}
