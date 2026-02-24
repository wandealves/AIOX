package scheduler

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"context"
)

// Service handles scheduled task business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new scheduler service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new scheduled task.
func (s *Service) Create(ctx context.Context, ownerUserID uuid.UUID, req CreateScheduledTaskRequest) (*ScheduledTask, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.CronExpression == "" {
		return nil, fmt.Errorf("cron_expression is required")
	}
	if req.AgentID == nil && req.PipelineID == nil {
		return nil, fmt.Errorf("either agent_id or pipeline_id is required")
	}

	// Validate cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(req.CronExpression)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	tz := req.Timezone
	if tz == "" {
		tz = "UTC"
	}
	// Validate timezone
	if _, err := time.LoadLocation(tz); err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}

	now := time.Now()
	nextRun := schedule.Next(now)

	st := &ScheduledTask{
		ID:             uuid.New(),
		OwnerUserID:    ownerUserID,
		AgentID:        req.AgentID,
		PipelineID:     req.PipelineID,
		Name:           req.Name,
		CronExpression: req.CronExpression,
		InputTemplate:  req.InputTemplate,
		Timezone:       tz,
		IsActive:       true,
		NextRunAt:      &nextRun,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

// GetByID returns a scheduled task by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*ScheduledTask, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByOwner returns all scheduled tasks for a user.
func (s *Service) ListByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]ScheduledTask, error) {
	return s.repo.ListByOwner(ctx, ownerUserID)
}

// Update updates a scheduled task.
func (s *Service) Update(ctx context.Context, id, ownerUserID uuid.UUID, req UpdateScheduledTaskRequest) (*ScheduledTask, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if st.OwnerUserID != ownerUserID {
		return nil, fmt.Errorf("scheduled task not owned by user")
	}

	if req.Name != nil {
		st.Name = *req.Name
	}
	if req.InputTemplate != nil {
		st.InputTemplate = *req.InputTemplate
	}
	if req.Timezone != nil {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			return nil, fmt.Errorf("invalid timezone: %w", err)
		}
		st.Timezone = *req.Timezone
	}
	if req.IsActive != nil {
		st.IsActive = *req.IsActive
	}
	if req.CronExpression != nil {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, err := parser.Parse(*req.CronExpression)
		if err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
		st.CronExpression = *req.CronExpression
		nextRun := schedule.Next(time.Now())
		st.NextRunAt = &nextRun
	}

	if err := s.repo.Update(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

// Delete soft-deletes a scheduled task.
func (s *Service) Delete(ctx context.Context, id, ownerUserID uuid.UUID) error {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if st.OwnerUserID != ownerUserID {
		return fmt.Errorf("scheduled task not owned by user")
	}
	return s.repo.Delete(ctx, id)
}

// ComputeNextRun calculates the next run time for a cron expression.
func ComputeNextRun(cronExpr string, after time.Time) (time.Time, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(after), nil
}
