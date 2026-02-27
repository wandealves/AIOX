package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SystemRepository handles DB operations for system tool definitions.
type SystemRepository struct {
	pool *pgxpool.Pool
}

// NewSystemRepository creates a new system tool repository.
func NewSystemRepository(pool *pgxpool.Pool) *SystemRepository {
	return &SystemRepository{pool: pool}
}

const systemToolColumns = `id, name, description, category, parameters, output_schema,
		tool_type, endpoint_url, http_method, headers, auth_type, auth_config,
		is_active, timeout_sec, version, plugin_id, created_at, updated_at`

func scanSystemTool(row interface{ Scan(dest ...any) error }) (*SystemTool, error) {
	var t SystemTool
	err := row.Scan(
		&t.ID, &t.Name, &t.Description, &t.Category, &t.Parameters, &t.OutputSchema,
		&t.ToolType, &t.EndpointURL, &t.HTTPMethod, &t.Headers, &t.AuthType, &t.AuthConfig,
		&t.IsActive, &t.TimeoutSec, &t.Version, &t.PluginID, &t.CreatedAt, &t.UpdatedAt,
	)
	return &t, err
}

// Create inserts a new system tool definition.
func (r *SystemRepository) Create(ctx context.Context, tool *SystemTool) error {
	query := `
		INSERT INTO system_tools (id, name, description, category, parameters, output_schema,
			tool_type, endpoint_url, http_method, headers, auth_type, auth_config,
			is_active, timeout_sec, version, plugin_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	_, err := r.pool.Exec(ctx, query,
		tool.ID, tool.Name, tool.Description, tool.Category, tool.Parameters, tool.OutputSchema,
		tool.ToolType, tool.EndpointURL, tool.HTTPMethod, tool.Headers, tool.AuthType, tool.AuthConfig,
		tool.IsActive, tool.TimeoutSec, tool.Version, tool.PluginID, tool.CreatedAt, tool.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting system tool: %w", err)
	}
	return nil
}

// GetByID retrieves a system tool by ID.
func (r *SystemRepository) GetByID(ctx context.Context, toolID uuid.UUID) (*SystemTool, error) {
	query := `SELECT ` + systemToolColumns + ` FROM system_tools WHERE id = $1`

	t, err := scanSystemTool(r.pool.QueryRow(ctx, query, toolID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting system tool by ID: %w", err)
	}
	return t, nil
}

// GetByName retrieves a system tool by name.
func (r *SystemRepository) GetByName(ctx context.Context, name string) (*SystemTool, error) {
	query := `SELECT ` + systemToolColumns + ` FROM system_tools WHERE name = $1`

	t, err := scanSystemTool(r.pool.QueryRow(ctx, query, name))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting system tool by name: %w", err)
	}
	return t, nil
}

// List retrieves all system tools.
func (r *SystemRepository) List(ctx context.Context) ([]*SystemTool, error) {
	query := `SELECT ` + systemToolColumns + ` FROM system_tools ORDER BY category, name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing system tools: %w", err)
	}
	defer rows.Close()

	var tools []*SystemTool
	for rows.Next() {
		t, err := scanSystemTool(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning system tool: %w", err)
		}
		tools = append(tools, t)
	}
	return tools, nil
}

// ListActive retrieves only active system tools.
func (r *SystemRepository) ListActive(ctx context.Context) ([]*SystemTool, error) {
	query := `SELECT ` + systemToolColumns + ` FROM system_tools WHERE is_active = true ORDER BY category, name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing active system tools: %w", err)
	}
	defer rows.Close()

	var tools []*SystemTool
	for rows.Next() {
		t, err := scanSystemTool(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning system tool: %w", err)
		}
		tools = append(tools, t)
	}
	return tools, nil
}

// ListByPlugin retrieves all system tools belonging to a plugin.
func (r *SystemRepository) ListByPlugin(ctx context.Context, pluginID uuid.UUID) ([]*SystemTool, error) {
	query := `SELECT ` + systemToolColumns + ` FROM system_tools WHERE plugin_id = $1 ORDER BY name`

	rows, err := r.pool.Query(ctx, query, pluginID)
	if err != nil {
		return nil, fmt.Errorf("listing tools by plugin: %w", err)
	}
	defer rows.Close()

	var tools []*SystemTool
	for rows.Next() {
		t, err := scanSystemTool(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning system tool: %w", err)
		}
		tools = append(tools, t)
	}
	return tools, nil
}

// Update updates an existing system tool definition.
func (r *SystemRepository) Update(ctx context.Context, tool *SystemTool) error {
	query := `
		UPDATE system_tools SET
			name = $2, description = $3, category = $4, parameters = $5,
			output_schema = $6, tool_type = $7, endpoint_url = $8,
			http_method = $9, headers = $10, auth_type = $11, auth_config = $12,
			is_active = $13, timeout_sec = $14, version = $15, updated_at = $16
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query,
		tool.ID, tool.Name, tool.Description, tool.Category, tool.Parameters,
		tool.OutputSchema, tool.ToolType, tool.EndpointURL,
		tool.HTTPMethod, tool.Headers, tool.AuthType, tool.AuthConfig,
		tool.IsActive, tool.TimeoutSec, tool.Version, tool.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating system tool: %w", err)
	}
	return nil
}

// Delete removes a system tool by ID.
func (r *SystemRepository) Delete(ctx context.Context, toolID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM system_tools WHERE id = $1`, toolID)
	if err != nil {
		return fmt.Errorf("deleting system tool: %w", err)
	}
	return nil
}

// DeleteByPlugin removes all system tools for a given plugin.
func (r *SystemRepository) DeleteByPlugin(ctx context.Context, pluginID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM system_tools WHERE plugin_id = $1`, pluginID)
	if err != nil {
		return fmt.Errorf("deleting tools by plugin: %w", err)
	}
	return nil
}
