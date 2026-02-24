package scheduler

import (
	"time"

	"github.com/google/uuid"
)

// ScheduledTask defines a recurring task that runs on a cron schedule.
type ScheduledTask struct {
	ID             uuid.UUID  `json:"id"`
	OwnerUserID    uuid.UUID  `json:"owner_user_id"`
	AgentID        *uuid.UUID `json:"agent_id,omitempty"`
	PipelineID     *uuid.UUID `json:"pipeline_id,omitempty"`
	Name           string     `json:"name"`
	CronExpression string     `json:"cron_expression"`
	InputTemplate  string     `json:"input_template"`
	Timezone       string     `json:"timezone"`
	IsActive       bool       `json:"is_active"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// CreateScheduledTaskRequest is the payload for creating a scheduled task.
type CreateScheduledTaskRequest struct {
	AgentID        *uuid.UUID `json:"agent_id,omitempty"`
	PipelineID     *uuid.UUID `json:"pipeline_id,omitempty"`
	Name           string     `json:"name"`
	CronExpression string     `json:"cron_expression"`
	InputTemplate  string     `json:"input_template"`
	Timezone       string     `json:"timezone"`
}

// UpdateScheduledTaskRequest is the payload for updating a scheduled task.
type UpdateScheduledTaskRequest struct {
	Name           *string `json:"name,omitempty"`
	CronExpression *string `json:"cron_expression,omitempty"`
	InputTemplate  *string `json:"input_template,omitempty"`
	Timezone       *string `json:"timezone,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
}
