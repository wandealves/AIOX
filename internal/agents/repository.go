package agents

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, row *AgentRow) error
	GetByID(ctx context.Context, id uuid.UUID) (*AgentRow, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*AgentRow, error)
	CountByOwner(ctx context.Context, ownerID uuid.UUID) (int64, error)
	Update(ctx context.Context, row *AgentRow) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) Create(ctx context.Context, row *AgentRow) error {
	query := `
		INSERT INTO agents (id, owner_user_id, jid, profile, llm_config, capabilities, memory_config, governance, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		row.ID, row.OwnerUserID, row.JID,
		row.Profile, row.LLMConfig, row.Capabilities,
		row.MemoryConfig, row.Governance, row.Visibility,
		row.CreatedAt, row.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting agent: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*AgentRow, error) {
	query := `
		SELECT id, owner_user_id, jid, profile, llm_config, capabilities, memory_config, governance, visibility, created_at, updated_at, deleted_at
		FROM agents
		WHERE id = $1 AND deleted_at IS NULL`

	row := &AgentRow{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&row.ID, &row.OwnerUserID, &row.JID,
		&row.Profile, &row.LLMConfig, &row.Capabilities,
		&row.MemoryConfig, &row.Governance, &row.Visibility,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying agent by id: %w", err)
	}
	return row, nil
}

func (r *postgresRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*AgentRow, error) {
	query := `
		SELECT id, owner_user_id, jid, profile, llm_config, capabilities, memory_config, governance, visibility, created_at, updated_at, deleted_at
		FROM agents
		WHERE owner_user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing agents: %w", err)
	}
	defer rows.Close()

	var agents []*AgentRow
	for rows.Next() {
		row := &AgentRow{}
		err := rows.Scan(
			&row.ID, &row.OwnerUserID, &row.JID,
			&row.Profile, &row.LLMConfig, &row.Capabilities,
			&row.MemoryConfig, &row.Governance, &row.Visibility,
			&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning agent row: %w", err)
		}
		agents = append(agents, row)
	}
	return agents, rows.Err()
}

func (r *postgresRepository) CountByOwner(ctx context.Context, ownerID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM agents WHERE owner_user_id = $1 AND deleted_at IS NULL`

	var count int64
	err := r.pool.QueryRow(ctx, query, ownerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting agents: %w", err)
	}
	return count, nil
}

func (r *postgresRepository) Update(ctx context.Context, row *AgentRow) error {
	query := `
		UPDATE agents
		SET profile = $2, llm_config = $3, capabilities = $4, memory_config = $5, governance = $6, visibility = $7, updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query,
		row.ID, row.Profile, row.LLMConfig, row.Capabilities,
		row.MemoryConfig, row.Governance, row.Visibility, row.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating agent: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("agent not found or already deleted")
	}
	return nil
}

func (r *postgresRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE agents SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft deleting agent: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("agent not found or already deleted")
	}
	return nil
}
