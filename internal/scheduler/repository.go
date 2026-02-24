package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles scheduled task persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new scheduler repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new scheduled task.
func (r *Repository) Create(ctx context.Context, st *ScheduledTask) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO scheduled_tasks (id, owner_user_id, agent_id, pipeline_id, name, cron_expression, input_template, timezone, is_active, next_run_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id`,
		st.ID, st.OwnerUserID, st.AgentID, st.PipelineID, st.Name, st.CronExpression,
		st.InputTemplate, st.Timezone, st.IsActive, st.NextRunAt, st.CreatedAt, st.UpdatedAt,
	).Scan(&st.ID)
}

// GetByID returns a scheduled task by ID (soft-delete aware).
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*ScheduledTask, error) {
	var st ScheduledTask
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_user_id, agent_id, pipeline_id, name, cron_expression, input_template,
		        timezone, is_active, last_run_at, next_run_at, created_at, updated_at, deleted_at
		 FROM scheduled_tasks WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&st.ID, &st.OwnerUserID, &st.AgentID, &st.PipelineID, &st.Name, &st.CronExpression,
		&st.InputTemplate, &st.Timezone, &st.IsActive, &st.LastRunAt, &st.NextRunAt,
		&st.CreatedAt, &st.UpdatedAt, &st.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// ListByOwner returns all scheduled tasks for a user.
func (r *Repository) ListByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]ScheduledTask, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_user_id, agent_id, pipeline_id, name, cron_expression, input_template,
		        timezone, is_active, last_run_at, next_run_at, created_at, updated_at
		 FROM scheduled_tasks WHERE owner_user_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at DESC`, ownerUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ScheduledTask
	for rows.Next() {
		var st ScheduledTask
		if err := rows.Scan(&st.ID, &st.OwnerUserID, &st.AgentID, &st.PipelineID, &st.Name, &st.CronExpression,
			&st.InputTemplate, &st.Timezone, &st.IsActive, &st.LastRunAt, &st.NextRunAt,
			&st.CreatedAt, &st.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, st)
	}
	return result, rows.Err()
}

// Update updates a scheduled task's mutable fields.
func (r *Repository) Update(ctx context.Context, st *ScheduledTask) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE scheduled_tasks
		 SET name = $1, cron_expression = $2, input_template = $3, timezone = $4,
		     is_active = $5, next_run_at = $6, updated_at = $7
		 WHERE id = $8 AND deleted_at IS NULL`,
		st.Name, st.CronExpression, st.InputTemplate, st.Timezone,
		st.IsActive, st.NextRunAt, time.Now(), st.ID,
	)
	return err
}

// Delete soft-deletes a scheduled task.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE scheduled_tasks SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	return err
}

// GetDueTasks returns active scheduled tasks whose next_run_at is <= now.
func (r *Repository) GetDueTasks(ctx context.Context, now time.Time) ([]ScheduledTask, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_user_id, agent_id, pipeline_id, name, cron_expression, input_template,
		        timezone, is_active, last_run_at, next_run_at, created_at, updated_at
		 FROM scheduled_tasks
		 WHERE is_active = true AND deleted_at IS NULL AND next_run_at IS NOT NULL AND next_run_at <= $1
		 ORDER BY next_run_at ASC
		 LIMIT 100`, now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ScheduledTask
	for rows.Next() {
		var st ScheduledTask
		if err := rows.Scan(&st.ID, &st.OwnerUserID, &st.AgentID, &st.PipelineID, &st.Name, &st.CronExpression,
			&st.InputTemplate, &st.Timezone, &st.IsActive, &st.LastRunAt, &st.NextRunAt,
			&st.CreatedAt, &st.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, st)
	}
	return result, rows.Err()
}

// MarkRun updates last_run_at and next_run_at after executing a task.
func (r *Repository) MarkRun(ctx context.Context, id uuid.UUID, lastRun, nextRun time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE scheduled_tasks SET last_run_at = $1, next_run_at = $2, updated_at = now()
		 WHERE id = $3`,
		lastRun, nextRun, id,
	)
	return err
}
