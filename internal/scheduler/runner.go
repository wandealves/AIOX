package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/agents"
	inats "github.com/aiox-platform/aiox/internal/nats"
)

const pollInterval = 60 * time.Second

// Runner is a background goroutine that polls for due scheduled tasks and executes them.
type Runner struct {
	repo      *Repository
	agentSvc  *agents.Service
	publisher *inats.Publisher
}

// NewRunner creates a new scheduler runner.
func NewRunner(repo *Repository, agentSvc *agents.Service, publisher *inats.Publisher) *Runner {
	return &Runner{
		repo:      repo,
		agentSvc:  agentSvc,
		publisher: publisher,
	}
}

// Start begins the polling loop. Blocks until ctx is cancelled.
func (r *Runner) Start(ctx context.Context) error {
	slog.Info("scheduler runner started", "poll_interval", pollInterval)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			r.poll(ctx)
		}
	}
}

func (r *Runner) poll(ctx context.Context) {
	now := time.Now()
	tasks, err := r.repo.GetDueTasks(ctx, now)
	if err != nil {
		slog.Error("scheduler: fetching due tasks", "error", err)
		return
	}

	for _, task := range tasks {
		r.executeTask(ctx, task, now)
	}
}

func (r *Runner) executeTask(ctx context.Context, task ScheduledTask, now time.Time) {
	slog.Info("scheduler: executing task", "task_id", task.ID, "name", task.Name)

	// Compute input from template (or use as-is)
	input := task.InputTemplate
	if input == "" {
		input = "Scheduled execution of " + task.Name
	}

	if task.AgentID != nil {
		r.executeAgentTask(ctx, task, input, now)
	} else if task.PipelineID != nil {
		r.executePipelineTask(ctx, task, input, now)
	}

	// Compute next run and update
	nextRun, err := ComputeNextRun(task.CronExpression, now)
	if err != nil {
		slog.Error("scheduler: computing next run", "error", err, "task_id", task.ID)
		return
	}
	if err := r.repo.MarkRun(ctx, task.ID, now, nextRun); err != nil {
		slog.Error("scheduler: marking run", "error", err, "task_id", task.ID)
	}
}

func (r *Runner) executeAgentTask(ctx context.Context, task ScheduledTask, input string, now time.Time) {
	agent, err := r.agentSvc.GetByID(ctx, *task.AgentID)
	if err != nil || agent == nil {
		slog.Error("scheduler: agent not found", "agent_id", task.AgentID, "error", err)
		return
	}

	// Publish an InboundMessage to NATS, which will flow through orchestrator → dispatcher → worker
	msg := inats.InboundMessage{
		ID:         uuid.New().String(),
		FromJID:    "scheduler:" + task.ID.String(),
		ToJID:      agent.JID,
		Body:       input,
		StanzaType: "message",
		ReceivedAt: now,
		Channel:    "ws", // scheduled results are routed to WS (or can be captured via API)
	}

	if err := r.publisher.PublishInboundMessage(ctx, msg); err != nil {
		slog.Error("scheduler: publishing inbound message", "error", err, "task_id", task.ID)
	}
}

func (r *Runner) executePipelineTask(ctx context.Context, task ScheduledTask, input string, now time.Time) {
	// For pipeline execution, publish an inbound message that the pipeline service can handle.
	// The pipeline execution is triggered by publishing to a special subject or via direct service call.
	// For simplicity, we publish an InboundMessage with a pipeline marker.
	slog.Info("scheduler: pipeline execution requested",
		"task_id", task.ID,
		"pipeline_id", task.PipelineID,
		"input", input,
	)
	// Pipeline scheduled execution would require the pipeline service to be injected.
	// For now, log the intent. Full pipeline-schedule integration requires the pipeline service.
}
