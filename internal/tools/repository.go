package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles DB operations for tool definitions.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new tool repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const toolColumns = `id, agent_id, tool_name, COALESCE(description,''), COALESCE(parameters,'{}'::jsonb),
		       COALESCE(endpoint_url,''), COALESCE(http_method,'GET'), COALESCE(headers,'{}'::jsonb),
		       COALESCE(auth_type,''), COALESCE(auth_config,'{}'::jsonb), COALESCE(is_active,true),
		       COALESCE(timeout_sec,30), COALESCE(version,'1.0.0'), COALESCE(category,'custom'),
		       COALESCE(output_schema,'{}'::jsonb), COALESCE(tool_type,'http'), utcp_manual_id,
		       created_at, COALESCE(updated_at, created_at)`

func scanTool(row interface{ Scan(dest ...any) error }) (*ToolDefinition, error) {
	var t ToolDefinition
	err := row.Scan(
		&t.ID, &t.AgentID, &t.Name, &t.Description, &t.Parameters,
		&t.EndpointURL, &t.HTTPMethod, &t.Headers,
		&t.AuthType, &t.AuthConfig, &t.IsActive, &t.TimeoutSec,
		&t.Version, &t.Category, &t.OutputSchema,
		&t.ToolType, &t.UTCPManualID,
		&t.CreatedAt, &t.UpdatedAt,
	)
	return &t, err
}

// Create inserts a new tool definition.
func (r *Repository) Create(ctx context.Context, tool *ToolDefinition) error {
	query := `
		INSERT INTO agent_tools (id, agent_id, tool_name, description, parameters, endpoint_url, http_method, headers, auth_type, auth_config, is_active, timeout_sec, version, category, output_schema, policy, tool_type, utcp_manual_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, '{}'::jsonb, $16, $17, $18, $19)`

	_, err := r.pool.Exec(ctx, query,
		tool.ID, tool.AgentID, tool.Name, tool.Description,
		tool.Parameters, tool.EndpointURL, tool.HTTPMethod, tool.Headers,
		tool.AuthType, tool.AuthConfig, tool.IsActive, tool.TimeoutSec,
		tool.Version, tool.Category, tool.OutputSchema,
		tool.ToolType, tool.UTCPManualID,
		tool.CreatedAt, tool.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting tool: %w", err)
	}
	return nil
}

// GetByID retrieves a tool definition by ID.
func (r *Repository) GetByID(ctx context.Context, toolID uuid.UUID) (*ToolDefinition, error) {
	query := `SELECT ` + toolColumns + ` FROM agent_tools WHERE id = $1`

	t, err := scanTool(r.pool.QueryRow(ctx, query, toolID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting tool by ID: %w", err)
	}
	return t, nil
}

// ListByAgent retrieves all tools for a given agent.
func (r *Repository) ListByAgent(ctx context.Context, agentID uuid.UUID) ([]*ToolDefinition, error) {
	query := `SELECT ` + toolColumns + ` FROM agent_tools WHERE agent_id = $1 ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, fmt.Errorf("listing tools by agent: %w", err)
	}
	defer rows.Close()

	var tools []*ToolDefinition
	for rows.Next() {
		t, err := scanTool(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning tool: %w", err)
		}
		tools = append(tools, t)
	}
	return tools, nil
}

// ListActiveByAgent retrieves only active tools for a given agent.
func (r *Repository) ListActiveByAgent(ctx context.Context, agentID uuid.UUID) ([]*ToolDefinition, error) {
	query := `SELECT ` + toolColumns + ` FROM agent_tools WHERE agent_id = $1 AND is_active = true ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, fmt.Errorf("listing active tools by agent: %w", err)
	}
	defer rows.Close()

	var tools []*ToolDefinition
	for rows.Next() {
		t, err := scanTool(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning tool: %w", err)
		}
		tools = append(tools, t)
	}
	return tools, nil
}

// Update updates an existing tool definition.
func (r *Repository) Update(ctx context.Context, tool *ToolDefinition) error {
	query := `
		UPDATE agent_tools SET
			tool_name = $2, description = $3, parameters = $4,
			endpoint_url = $5, http_method = $6, headers = $7,
			auth_type = $8, auth_config = $9, is_active = $10,
			timeout_sec = $11, version = $12, category = $13,
			output_schema = $14, updated_at = $15
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query,
		tool.ID, tool.Name, tool.Description, tool.Parameters,
		tool.EndpointURL, tool.HTTPMethod, tool.Headers,
		tool.AuthType, tool.AuthConfig, tool.IsActive, tool.TimeoutSec,
		tool.Version, tool.Category, tool.OutputSchema,
		tool.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating tool: %w", err)
	}
	return nil
}

// Delete removes a tool definition by ID.
func (r *Repository) Delete(ctx context.Context, toolID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM agent_tools WHERE id = $1`, toolID)
	if err != nil {
		return fmt.Errorf("deleting tool: %w", err)
	}
	return nil
}

// ToolBelongsToAgent checks if a tool belongs to a specific agent.
func (r *Repository) ToolBelongsToAgent(ctx context.Context, toolID, agentID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM agent_tools WHERE id = $1 AND agent_id = $2)`, toolID, agentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking tool ownership: %w", err)
	}
	return exists, nil
}

// ToolsToJSON converts a list of tools to a JSON array suitable for executions.tools_called.
func ToolsToJSON(calls []ToolCallRecord) json.RawMessage {
	if len(calls) == 0 {
		return json.RawMessage(`[]`)
	}
	data, err := json.Marshal(calls)
	if err != nil {
		return json.RawMessage(`[]`)
	}
	return data
}

// ToolCallRecord is a simplified representation of a tool call for DB storage.
type ToolCallRecord struct {
	ToolName  string `json:"tool_name"`
	Arguments string `json:"arguments"`
	Result    string `json:"result"`
	Duration  int    `json:"duration_ms"`
	Error     string `json:"error,omitempty"`
}
