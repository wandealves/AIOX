package plugins

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles DB operations for plugins.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new plugin repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new plugin.
func (r *Repository) Create(ctx context.Context, p *Plugin) error {
	query := `
		INSERT INTO plugins (id, name, description, manifest_url, is_active, auto_sync, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Name, p.Description, p.ManifestURL,
		p.IsActive, p.AutoSync, p.CreatedBy, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting plugin: %w", err)
	}
	return nil
}

// GetByID retrieves a plugin by ID.
func (r *Repository) GetByID(ctx context.Context, pluginID uuid.UUID) (*Plugin, error) {
	query := `
		SELECT id, name, description, manifest_url, is_active, auto_sync,
		       last_synced_at, created_by, created_at, updated_at
		FROM plugins WHERE id = $1`

	var p Plugin
	err := r.pool.QueryRow(ctx, query, pluginID).Scan(
		&p.ID, &p.Name, &p.Description, &p.ManifestURL,
		&p.IsActive, &p.AutoSync, &p.LastSyncedAt,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting plugin by ID: %w", err)
	}
	return &p, nil
}

// GetByName retrieves a plugin by name.
func (r *Repository) GetByName(ctx context.Context, name string) (*Plugin, error) {
	query := `
		SELECT id, name, description, manifest_url, is_active, auto_sync,
		       last_synced_at, created_by, created_at, updated_at
		FROM plugins WHERE name = $1`

	var p Plugin
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&p.ID, &p.Name, &p.Description, &p.ManifestURL,
		&p.IsActive, &p.AutoSync, &p.LastSyncedAt,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting plugin by name: %w", err)
	}
	return &p, nil
}

// List retrieves all plugins.
func (r *Repository) List(ctx context.Context) ([]*Plugin, error) {
	query := `
		SELECT id, name, description, manifest_url, is_active, auto_sync,
		       last_synced_at, created_by, created_at, updated_at
		FROM plugins ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*Plugin
	for rows.Next() {
		var p Plugin
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.ManifestURL,
			&p.IsActive, &p.AutoSync, &p.LastSyncedAt,
			&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning plugin: %w", err)
		}
		plugins = append(plugins, &p)
	}
	return plugins, nil
}

// UpdateLastSyncedAt updates the last synced timestamp.
func (r *Repository) UpdateLastSyncedAt(ctx context.Context, pluginID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE plugins SET last_synced_at = NOW(), updated_at = NOW() WHERE id = $1`,
		pluginID,
	)
	if err != nil {
		return fmt.Errorf("updating last_synced_at: %w", err)
	}
	return nil
}

// Delete removes a plugin by ID. CASCADE deletes linked system_tools.
func (r *Repository) Delete(ctx context.Context, pluginID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM plugins WHERE id = $1`, pluginID)
	if err != nil {
		return fmt.Errorf("deleting plugin: %w", err)
	}
	return nil
}
