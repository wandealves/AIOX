package pipelines

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestStepsToJSON(t *testing.T) {
	steps := []PipelineStep{
		{AgentID: uuid.New(), Transform: "Summarize: {{.Input}}", TimeoutSec: 60},
		{AgentID: uuid.New(), TimeoutSec: 0},
	}

	data := StepsToJSON(steps)
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}

	var parsed []PipelineStep
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(parsed))
	}
	if parsed[0].Transform != "Summarize: {{.Input}}" {
		t.Errorf("expected transform, got %q", parsed[0].Transform)
	}
	if parsed[0].AgentID != steps[0].AgentID {
		t.Errorf("agent ID mismatch")
	}
}

func TestStepResultsToJSON(t *testing.T) {
	results := []StepResult{
		{StepIndex: 0, AgentID: uuid.New().String(), Input: "hello", Output: "world", Tokens: 100, DurationMs: 500},
		{StepIndex: 1, AgentID: uuid.New().String(), Input: "world", Output: "done", Tokens: 50, DurationMs: 300, Error: ""},
	}

	data := StepResultsToJSON(results)
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}

	var parsed []StepResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("expected 2 results, got %d", len(parsed))
	}
	if parsed[0].Output != "world" {
		t.Errorf("expected output 'world', got %q", parsed[0].Output)
	}
	if parsed[0].Tokens != 100 {
		t.Errorf("expected tokens 100, got %d", parsed[0].Tokens)
	}
}

func TestStepsToJSON_NilSlice(t *testing.T) {
	data := StepsToJSON(nil)
	if string(data) != "null" {
		t.Errorf("expected null for nil slice, got %q", string(data))
	}
}

func TestStepResultsToJSON_Empty(t *testing.T) {
	data := StepResultsToJSON([]StepResult{})
	if string(data) != "[]" {
		t.Errorf("expected [], got %q", string(data))
	}
}
