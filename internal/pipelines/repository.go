package pipelines

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles pipeline persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new pipeline repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreatePipeline inserts a new pipeline.
func (r *Repository) CreatePipeline(ctx context.Context, p *Pipeline) error {
	stepsJSON := StepsToJSON(p.Steps)
	return r.pool.QueryRow(ctx,
		`INSERT INTO pipelines (id, owner_user_id, org_id, name, description, steps, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		p.ID, p.OwnerUserID, p.OrgID, p.Name, p.Description, stepsJSON, p.IsActive, p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID)
}

// GetByID returns a pipeline by ID (soft-delete aware).
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Pipeline, error) {
	var p Pipeline
	var stepsJSON json.RawMessage
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_user_id, org_id, name, description, steps, is_active, created_at, updated_at, deleted_at
		 FROM pipelines WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&p.ID, &p.OwnerUserID, &p.OrgID, &p.Name, &p.Description, &stepsJSON, &p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(stepsJSON, &p.Steps); err != nil {
		p.Steps = []PipelineStep{}
	}
	return &p, nil
}

// ListByOwner returns all pipelines for a user.
func (r *Repository) ListByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]Pipeline, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_user_id, org_id, name, description, steps, is_active, created_at, updated_at
		 FROM pipelines WHERE owner_user_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at DESC`, ownerUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Pipeline
	for rows.Next() {
		var p Pipeline
		var stepsJSON json.RawMessage
		if err := rows.Scan(&p.ID, &p.OwnerUserID, &p.OrgID, &p.Name, &p.Description, &stepsJSON, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(stepsJSON, &p.Steps); err != nil {
			p.Steps = []PipelineStep{}
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// Update updates a pipeline's mutable fields.
func (r *Repository) Update(ctx context.Context, p *Pipeline) error {
	stepsJSON := StepsToJSON(p.Steps)
	_, err := r.pool.Exec(ctx,
		`UPDATE pipelines SET name = $1, description = $2, steps = $3, is_active = $4, updated_at = $5
		 WHERE id = $6 AND deleted_at IS NULL`,
		p.Name, p.Description, stepsJSON, p.IsActive, time.Now(), p.ID,
	)
	return err
}

// Delete soft-deletes a pipeline.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE pipelines SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	return err
}

// CreateExecution inserts a new pipeline execution record.
func (r *Repository) CreateExecution(ctx context.Context, e *PipelineExecution) error {
	stepResultsJSON := StepResultsToJSON(e.StepResults)
	return r.pool.QueryRow(ctx,
		`INSERT INTO pipeline_executions (id, pipeline_id, owner_user_id, input, output, current_step, total_steps, status, step_results, total_tokens, total_duration_ms, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 RETURNING id`,
		e.ID, e.PipelineID, e.OwnerUserID, e.Input, e.Output, e.CurrentStep, e.TotalSteps,
		e.Status, stepResultsJSON, e.TotalTokens, e.TotalDurationMs, e.CreatedAt, e.UpdatedAt,
	).Scan(&e.ID)
}

// UpdateExecution updates an execution's progress.
func (r *Repository) UpdateExecution(ctx context.Context, e *PipelineExecution) error {
	stepResultsJSON := StepResultsToJSON(e.StepResults)
	_, err := r.pool.Exec(ctx,
		`UPDATE pipeline_executions
		 SET output = $1, current_step = $2, status = $3, step_results = $4,
		     total_tokens = $5, total_duration_ms = $6, updated_at = $7
		 WHERE id = $8`,
		e.Output, e.CurrentStep, e.Status, stepResultsJSON,
		e.TotalTokens, e.TotalDurationMs, time.Now(), e.ID,
	)
	return err
}

// GetExecution returns a pipeline execution by ID.
func (r *Repository) GetExecution(ctx context.Context, id uuid.UUID) (*PipelineExecution, error) {
	var e PipelineExecution
	var stepResultsJSON json.RawMessage
	err := r.pool.QueryRow(ctx,
		`SELECT id, pipeline_id, owner_user_id, input, output, current_step, total_steps, status,
		        step_results, total_tokens, total_duration_ms, created_at, updated_at
		 FROM pipeline_executions WHERE id = $1`, id,
	).Scan(&e.ID, &e.PipelineID, &e.OwnerUserID, &e.Input, &e.Output, &e.CurrentStep, &e.TotalSteps,
		&e.Status, &stepResultsJSON, &e.TotalTokens, &e.TotalDurationMs, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(stepResultsJSON, &e.StepResults); err != nil {
		e.StepResults = []StepResult{}
	}
	return &e, nil
}

// ListExecutionsByPipeline returns executions for a pipeline.
func (r *Repository) ListExecutionsByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]PipelineExecution, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, pipeline_id, owner_user_id, input, output, current_step, total_steps, status,
		        step_results, total_tokens, total_duration_ms, created_at, updated_at
		 FROM pipeline_executions WHERE pipeline_id = $1
		 ORDER BY created_at DESC LIMIT 50`, pipelineID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PipelineExecution
	for rows.Next() {
		var e PipelineExecution
		var stepResultsJSON json.RawMessage
		if err := rows.Scan(&e.ID, &e.PipelineID, &e.OwnerUserID, &e.Input, &e.Output, &e.CurrentStep, &e.TotalSteps,
			&e.Status, &stepResultsJSON, &e.TotalTokens, &e.TotalDurationMs, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(stepResultsJSON, &e.StepResults); err != nil {
			e.StepResults = []StepResult{}
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
