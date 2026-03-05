package utcp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles DB operations for UTCP manuals.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new UTCP repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const manualColumns = `id, agent_id, name, COALESCE(description,''), manual_url,
	COALESCE(utcp_version,''), COALESCE(auth_template,'{}'::jsonb),
	COALESCE(auth_values,''), COALESCE(is_active,true), last_synced_at,
	COALESCE(tools_count,0), created_at, COALESCE(updated_at, created_at)`

func scanManual(row interface{ Scan(dest ...any) error }) (*UTCPManual, error) {
	var m UTCPManual
	err := row.Scan(
		&m.ID, &m.AgentID, &m.Name, &m.Description, &m.ManualURL,
		&m.UTCPVersion, &m.AuthTemplate,
		&m.AuthValues, &m.IsActive, &m.LastSyncedAt,
		&m.ToolsCount, &m.CreatedAt, &m.UpdatedAt,
	)
	return &m, err
}

// Create inserts a new UTCP manual.
func (r *Repository) Create(ctx context.Context, m *UTCPManual) error {
	query := `
		INSERT INTO utcp_manuals (id, agent_id, name, description, manual_url, utcp_version, auth_template, auth_values, is_active, last_synced_at, tools_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.pool.Exec(ctx, query,
		m.ID, m.AgentID, m.Name, m.Description, m.ManualURL,
		m.UTCPVersion, m.AuthTemplate,
		m.AuthValues, m.IsActive, m.LastSyncedAt,
		m.ToolsCount, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting UTCP manual: %w", err)
	}
	return nil
}

// GetByID retrieves a UTCP manual by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*UTCPManual, error) {
	query := `SELECT ` + manualColumns + ` FROM utcp_manuals WHERE id = $1`
	m, err := scanManual(r.pool.QueryRow(ctx, query, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting UTCP manual: %w", err)
	}
	return m, nil
}

// ListByAgent retrieves all UTCP manuals for an agent.
func (r *Repository) ListByAgent(ctx context.Context, agentID uuid.UUID) ([]*UTCPManual, error) {
	query := `SELECT ` + manualColumns + ` FROM utcp_manuals WHERE agent_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, fmt.Errorf("listing UTCP manuals: %w", err)
	}
	defer rows.Close()

	var manuals []*UTCPManual
	for rows.Next() {
		m, err := scanManual(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning UTCP manual: %w", err)
		}
		manuals = append(manuals, m)
	}
	return manuals, nil
}

// GetByAgentAndName retrieves a UTCP manual by agent ID and name.
func (r *Repository) GetByAgentAndName(ctx context.Context, agentID uuid.UUID, name string) (*UTCPManual, error) {
	query := `SELECT ` + manualColumns + ` FROM utcp_manuals WHERE agent_id = $1 AND name = $2`
	m, err := scanManual(r.pool.QueryRow(ctx, query, agentID, name))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting UTCP manual by name: %w", err)
	}
	return m, nil
}

// Delete removes a UTCP manual by ID (cascades to agent_tools via FK).
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM utcp_manuals WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting UTCP manual: %w", err)
	}
	return nil
}

// UpdateSyncInfo updates the last_synced_at and tools_count fields.
func (r *Repository) UpdateSyncInfo(ctx context.Context, id uuid.UUID, toolsCount int) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE utcp_manuals SET last_synced_at = $1, tools_count = $2, updated_at = $3 WHERE id = $4`,
		now, toolsCount, now, id,
	)
	if err != nil {
		return fmt.Errorf("updating UTCP manual sync info: %w", err)
	}
	return nil
}

// ManualBelongsToAgent checks if a UTCP manual belongs to a specific agent.
func (r *Repository) ManualBelongsToAgent(ctx context.Context, manualID, agentID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM utcp_manuals WHERE id = $1 AND agent_id = $2)`,
		manualID, agentID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking UTCP manual ownership: %w", err)
	}
	return exists, nil
}

// DeleteToolsByManual deletes all agent_tools that belong to a UTCP manual.
func (r *Repository) DeleteToolsByManual(ctx context.Context, manualID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM agent_tools WHERE utcp_manual_id = $1`, manualID)
	if err != nil {
		return fmt.Errorf("deleting tools by manual: %w", err)
	}
	return nil
}
