package pipelines

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PipelineStep defines a single step in a pipeline.
type PipelineStep struct {
	AgentID    uuid.UUID `json:"agent_id"`
	Transform  string    `json:"transform,omitempty"`  // Go template to transform previous output → next input
	TimeoutSec int       `json:"timeout_sec,omitempty"` // per-step timeout (0 = use default)
}

// Pipeline defines a sequence of agent steps.
type Pipeline struct {
	ID          uuid.UUID      `json:"id"`
	OwnerUserID uuid.UUID      `json:"owner_user_id"`
	OrgID       *uuid.UUID     `json:"org_id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []PipelineStep `json:"steps"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty"`
}

// StepResult records the outcome of a single pipeline step execution.
type StepResult struct {
	StepIndex  int    `json:"step_index"`
	AgentID    string `json:"agent_id"`
	Input      string `json:"input"`
	Output     string `json:"output"`
	Tokens     int    `json:"tokens"`
	DurationMs int    `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// PipelineExecution records a single run of a pipeline.
type PipelineExecution struct {
	ID              uuid.UUID    `json:"id"`
	PipelineID      uuid.UUID    `json:"pipeline_id"`
	OwnerUserID     uuid.UUID    `json:"owner_user_id"`
	Input           string       `json:"input"`
	Output          string       `json:"output"`
	CurrentStep     int          `json:"current_step"`
	TotalSteps      int          `json:"total_steps"`
	Status          string       `json:"status"` // pending, running, completed, error
	StepResults     []StepResult `json:"step_results"`
	TotalTokens     int          `json:"total_tokens"`
	TotalDurationMs int          `json:"total_duration_ms"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// CreatePipelineRequest is the payload for creating a pipeline.
type CreatePipelineRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []PipelineStep `json:"steps"`
	OrgID       *uuid.UUID     `json:"org_id,omitempty"`
}

// UpdatePipelineRequest is the payload for updating a pipeline.
type UpdatePipelineRequest struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Steps       *[]PipelineStep `json:"steps,omitempty"`
	IsActive    *bool           `json:"is_active,omitempty"`
}

// ExecutePipelineRequest is the payload for triggering a pipeline execution.
type ExecutePipelineRequest struct {
	Input string `json:"input"`
}

// PipelineResultCallback is called by the dispatcher when a pipeline step completes.
type PipelineResultCallback func(requestID, text string, tokens, durationMs int, errMsg string)

// StepsToJSON marshals pipeline steps to JSON for storage.
func StepsToJSON(steps []PipelineStep) json.RawMessage {
	data, err := json.Marshal(steps)
	if err != nil {
		return json.RawMessage("[]")
	}
	return data
}

// StepResultsToJSON marshals step results to JSON for storage.
func StepResultsToJSON(results []StepResult) json.RawMessage {
	data, err := json.Marshal(results)
	if err != nil {
		return json.RawMessage("[]")
	}
	return data
}
