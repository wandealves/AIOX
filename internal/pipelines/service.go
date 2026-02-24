package pipelines

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"text/template"
	"time"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/agents"
	inats "github.com/aiox-platform/aiox/internal/nats"
)

const defaultStepTimeout = 120 * time.Second

// Service handles pipeline business logic and execution.
type Service struct {
	repo      *Repository
	agentSvc  *agents.Service
	publisher *inats.Publisher

	mu       sync.Mutex
	waiters  map[string]chan *stepResponse // requestID → result channel
}

type stepResponse struct {
	Text       string
	Tokens     int
	DurationMs int
	Error      string
}

// NewService creates a new pipeline service.
func NewService(repo *Repository, agentSvc *agents.Service, publisher *inats.Publisher) *Service {
	return &Service{
		repo:      repo,
		agentSvc:  agentSvc,
		publisher: publisher,
		waiters:   make(map[string]chan *stepResponse),
	}
}

// HandlePipelineResult is called by the dispatcher when a pipeline step completes.
// This is registered as a callback during wiring in main.go.
func (s *Service) HandlePipelineResult(requestID, text string, tokens, durationMs int, errMsg string) {
	s.mu.Lock()
	ch, ok := s.waiters[requestID]
	if ok {
		delete(s.waiters, requestID)
	}
	s.mu.Unlock()

	if ok {
		ch <- &stepResponse{
			Text:       text,
			Tokens:     tokens,
			DurationMs: durationMs,
			Error:      errMsg,
		}
	}
}

// Create creates a new pipeline.
func (s *Service) Create(ctx context.Context, ownerUserID uuid.UUID, req CreatePipelineRequest) (*Pipeline, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("pipeline name is required")
	}
	if len(req.Steps) == 0 {
		return nil, fmt.Errorf("pipeline must have at least one step")
	}

	// Validate each step references an accessible agent
	for i, step := range req.Steps {
		agent, err := s.agentSvc.GetByID(ctx, step.AgentID)
		if err != nil || agent == nil {
			return nil, fmt.Errorf("step %d: agent %s not found", i, step.AgentID)
		}
		if agent.OwnerUserID != ownerUserID {
			return nil, fmt.Errorf("step %d: agent %s not owned by user", i, step.AgentID)
		}
	}

	now := time.Now()
	p := &Pipeline{
		ID:          uuid.New(),
		OwnerUserID: ownerUserID,
		OrgID:       req.OrgID,
		Name:        req.Name,
		Description: req.Description,
		Steps:       req.Steps,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreatePipeline(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetByID returns a pipeline by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Pipeline, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByOwner returns all pipelines for a user.
func (s *Service) ListByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]Pipeline, error) {
	return s.repo.ListByOwner(ctx, ownerUserID)
}

// Update updates a pipeline.
func (s *Service) Update(ctx context.Context, id uuid.UUID, ownerUserID uuid.UUID, req UpdatePipelineRequest) (*Pipeline, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.OwnerUserID != ownerUserID {
		return nil, fmt.Errorf("pipeline not owned by user")
	}

	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Steps != nil {
		// Validate new steps
		for i, step := range *req.Steps {
			agent, err := s.agentSvc.GetByID(ctx, step.AgentID)
			if err != nil || agent == nil {
				return nil, fmt.Errorf("step %d: agent %s not found", i, step.AgentID)
			}
			if agent.OwnerUserID != ownerUserID {
				return nil, fmt.Errorf("step %d: agent %s not owned by user", i, step.AgentID)
			}
		}
		p.Steps = *req.Steps
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete soft-deletes a pipeline.
func (s *Service) Delete(ctx context.Context, id uuid.UUID, ownerUserID uuid.UUID) error {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if p.OwnerUserID != ownerUserID {
		return fmt.Errorf("pipeline not owned by user")
	}
	return s.repo.Delete(ctx, id)
}

// Execute runs a pipeline synchronously, step by step.
func (s *Service) Execute(ctx context.Context, ownerUserID uuid.UUID, pipelineID uuid.UUID, input string) (*PipelineExecution, error) {
	p, err := s.repo.GetByID(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	if p.OwnerUserID != ownerUserID {
		return nil, fmt.Errorf("pipeline not owned by user")
	}
	if !p.IsActive {
		return nil, fmt.Errorf("pipeline is not active")
	}
	if len(p.Steps) == 0 {
		return nil, fmt.Errorf("pipeline has no steps")
	}

	now := time.Now()
	exec := &PipelineExecution{
		ID:          uuid.New(),
		PipelineID:  pipelineID,
		OwnerUserID: ownerUserID,
		Input:       input,
		TotalSteps:  len(p.Steps),
		Status:      "running",
		StepResults: []StepResult{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateExecution(ctx, exec); err != nil {
		return nil, fmt.Errorf("creating execution: %w", err)
	}

	currentInput := input
	totalTokens := 0
	startTime := time.Now()

	for i, step := range p.Steps {
		exec.CurrentStep = i + 1

		// Apply transform template if configured
		stepInput := currentInput
		if step.Transform != "" {
			transformed, err := applyTransform(step.Transform, currentInput)
			if err != nil {
				slog.Warn("pipeline: transform failed, using raw input", "error", err, "step", i)
			} else {
				stepInput = transformed
			}
		}

		// Execute this step
		stepTimeout := defaultStepTimeout
		if step.TimeoutSec > 0 {
			stepTimeout = time.Duration(step.TimeoutSec) * time.Second
		}

		result, err := s.executeStep(ctx, ownerUserID, exec.ID.String(), step, stepInput, i, stepTimeout)
		if err != nil {
			stepResult := StepResult{
				StepIndex: i,
				AgentID:   step.AgentID.String(),
				Input:     stepInput,
				Error:     err.Error(),
			}
			exec.StepResults = append(exec.StepResults, stepResult)
			exec.Status = "error"
			exec.Output = "Pipeline failed at step " + fmt.Sprintf("%d: %s", i+1, err.Error())
			exec.TotalTokens = totalTokens
			exec.TotalDurationMs = int(time.Since(startTime).Milliseconds())
			_ = s.repo.UpdateExecution(ctx, exec)
			return exec, nil
		}

		stepResult := StepResult{
			StepIndex:  i,
			AgentID:    step.AgentID.String(),
			Input:      stepInput,
			Output:     result.Text,
			Tokens:     result.Tokens,
			DurationMs: result.DurationMs,
			Error:      result.Error,
		}
		exec.StepResults = append(exec.StepResults, stepResult)
		totalTokens += result.Tokens

		if result.Error != "" {
			exec.Status = "error"
			exec.Output = "Step " + fmt.Sprintf("%d error: %s", i+1, result.Error)
			exec.TotalTokens = totalTokens
			exec.TotalDurationMs = int(time.Since(startTime).Milliseconds())
			_ = s.repo.UpdateExecution(ctx, exec)
			return exec, nil
		}

		currentInput = result.Text
	}

	exec.Status = "completed"
	exec.Output = currentInput
	exec.TotalTokens = totalTokens
	exec.TotalDurationMs = int(time.Since(startTime).Milliseconds())
	_ = s.repo.UpdateExecution(ctx, exec)

	return exec, nil
}

func (s *Service) executeStep(
	ctx context.Context,
	ownerUserID uuid.UUID,
	executionID string,
	step PipelineStep,
	input string,
	stepIndex int,
	timeout time.Duration,
) (*stepResponse, error) {
	// Fetch agent to get JID info
	agent, err := s.agentSvc.GetByID(ctx, step.AgentID)
	if err != nil || agent == nil {
		return nil, fmt.Errorf("agent %s not found", step.AgentID)
	}

	requestID := uuid.New().String()

	// Register waiter before publishing
	ch := make(chan *stepResponse, 1)
	s.mu.Lock()
	s.waiters[requestID] = ch
	s.mu.Unlock()

	// Cleanup on exit
	defer func() {
		s.mu.Lock()
		delete(s.waiters, requestID)
		s.mu.Unlock()
	}()

	// Publish task to NATS (same subject pattern as orchestrator)
	task := inats.TaskMessage{
		RequestID:           requestID,
		AgentID:             step.AgentID,
		OwnerUserID:         ownerUserID,
		Message:             input,
		FromJID:             "pipeline:" + executionID,
		AgentJID:            agent.JID,
		AgentName:           agent.Profile.Name,
		Channel:             "pipeline",
		PipelineExecutionID: executionID,
		PipelineStepIndex:   stepIndex,
	}

	if err := s.publisher.PublishTask(ctx, step.AgentID.String(), task); err != nil {
		return nil, fmt.Errorf("publishing task: %w", err)
	}

	// Wait for result
	select {
	case result := <-ch:
		return result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("step timed out after %s", timeout)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetExecution returns a pipeline execution by ID.
func (s *Service) GetExecution(ctx context.Context, id uuid.UUID) (*PipelineExecution, error) {
	return s.repo.GetExecution(ctx, id)
}

// ListExecutions returns executions for a pipeline.
func (s *Service) ListExecutions(ctx context.Context, pipelineID uuid.UUID) ([]PipelineExecution, error) {
	return s.repo.ListExecutionsByPipeline(ctx, pipelineID)
}

// applyTransform executes a Go text/template to transform the input.
func applyTransform(tmpl string, input string) (string, error) {
	t, err := template.New("transform").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]string{"Input": input}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
