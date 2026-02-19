package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Execution represents a recorded task execution.
type Execution struct {
	ID              uuid.UUID
	OwnerUserID     uuid.UUID
	AgentID         uuid.UUID
	Input           string
	Output          string
	TokensUsed      int
	WorkerID        string
	DurationMs      int
	GoLatencyMs     int
	PythonLatencyMs int
	Status          string
	ErrorMessage    string
	CreatedAt       time.Time
}

// Repository handles DB operations for workers and executions.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new worker repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// RecordExecution inserts an execution record into the database.
func (r *Repository) RecordExecution(ctx context.Context, exec *Execution) error {
	query := `
		INSERT INTO executions (id, owner_user_id, agent_id, input, output, tokens_used, worker_id, duration_ms, go_latency_ms, python_latency_ms, status, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.pool.Exec(ctx, query,
		exec.ID, exec.OwnerUserID, exec.AgentID,
		exec.Input, exec.Output, exec.TokensUsed,
		exec.WorkerID, exec.DurationMs, exec.GoLatencyMs, exec.PythonLatencyMs,
		exec.Status, exec.ErrorMessage, exec.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting execution: %w", err)
	}
	return nil
}

// UpsertWorker inserts or updates a worker record on registration.
func (r *Repository) UpsertWorker(ctx context.Context, workerID, host string, port int, capabilities []byte) error {
	query := `
		INSERT INTO ai_workers (id, worker_id, host, port, status, capabilities, last_heartbeat, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, 'healthy', $4, NOW(), NOW(), NOW())
		ON CONFLICT (worker_id) DO UPDATE
		SET host = $2, port = $3, status = 'healthy', capabilities = $4, last_heartbeat = NOW(), updated_at = NOW()`

	_, err := r.pool.Exec(ctx, query, workerID, host, port, capabilities)
	if err != nil {
		return fmt.Errorf("upserting worker: %w", err)
	}
	return nil
}

// UpdateWorkerHeartbeat updates heartbeat metrics for a worker.
func (r *Repository) UpdateWorkerHeartbeat(ctx context.Context, workerID string, activeRequests, avgLatencyMs, memoryUsageMb int) error {
	query := `
		UPDATE ai_workers
		SET last_heartbeat = NOW(), active_requests = $2, avg_latency_ms = $3, memory_usage_mb = $4, updated_at = NOW()
		WHERE worker_id = $1`

	_, err := r.pool.Exec(ctx, query, workerID, activeRequests, avgLatencyMs, memoryUsageMb)
	if err != nil {
		return fmt.Errorf("updating worker heartbeat: %w", err)
	}
	return nil
}

// MarkWorkerOffline sets a worker's status to "offline".
func (r *Repository) MarkWorkerOffline(ctx context.Context, workerID string) error {
	query := `UPDATE ai_workers SET status = 'offline', updated_at = NOW() WHERE worker_id = $1`

	_, err := r.pool.Exec(ctx, query, workerID)
	if err != nil {
		return fmt.Errorf("marking worker offline: %w", err)
	}
	return nil
}
