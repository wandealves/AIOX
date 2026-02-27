package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PermissionsRepository handles DB operations for tool permissions.
type PermissionsRepository struct {
	pool *pgxpool.Pool
}

// NewPermissionsRepository creates a new permissions repository.
func NewPermissionsRepository(pool *pgxpool.Pool) *PermissionsRepository {
	return &PermissionsRepository{pool: pool}
}

// Create inserts a new tool permission.
func (r *PermissionsRepository) Create(ctx context.Context, perm *ToolPermission) error {
	query := `
		INSERT INTO tool_permissions (id, tool_id, tool_source, grantee_type, grantee_id, permission, granted_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tool_id, tool_source, grantee_type, grantee_id) DO UPDATE SET permission = $6`

	_, err := r.pool.Exec(ctx, query,
		perm.ID, perm.ToolID, perm.ToolSource, perm.GranteeType,
		perm.GranteeID, perm.Permission, perm.GrantedBy, perm.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting tool permission: %w", err)
	}
	return nil
}

// HasPermission checks if a grantee has a specific permission on a tool.
func (r *PermissionsRepository) HasPermission(ctx context.Context, toolID uuid.UUID, toolSource, granteeType, granteeID, permission string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM tool_permissions
			WHERE tool_id = $1 AND tool_source = $2
			AND grantee_type = $3 AND grantee_id = $4
			AND permission = $5
		)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, toolID, toolSource, granteeType, granteeID, permission).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking tool permission: %w", err)
	}
	return exists, nil
}

// ListForTool retrieves all permissions for a specific tool.
func (r *PermissionsRepository) ListForTool(ctx context.Context, toolID uuid.UUID, toolSource string) ([]*ToolPermission, error) {
	query := `
		SELECT id, tool_id, tool_source, grantee_type, grantee_id, permission, granted_by, created_at
		FROM tool_permissions WHERE tool_id = $1 AND tool_source = $2 ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, toolID, toolSource)
	if err != nil {
		return nil, fmt.Errorf("listing tool permissions: %w", err)
	}
	defer rows.Close()

	var perms []*ToolPermission
	for rows.Next() {
		var p ToolPermission
		if err := rows.Scan(
			&p.ID, &p.ToolID, &p.ToolSource, &p.GranteeType,
			&p.GranteeID, &p.Permission, &p.GrantedBy, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning tool permission: %w", err)
		}
		perms = append(perms, &p)
	}
	return perms, nil
}

// ListForGrantee retrieves all permissions for a specific grantee.
func (r *PermissionsRepository) ListForGrantee(ctx context.Context, granteeType, granteeID string) ([]*ToolPermission, error) {
	query := `
		SELECT id, tool_id, tool_source, grantee_type, grantee_id, permission, granted_by, created_at
		FROM tool_permissions WHERE grantee_type = $1 AND grantee_id = $2 ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, granteeType, granteeID)
	if err != nil {
		return nil, fmt.Errorf("listing grantee permissions: %w", err)
	}
	defer rows.Close()

	var perms []*ToolPermission
	for rows.Next() {
		var p ToolPermission
		if err := rows.Scan(
			&p.ID, &p.ToolID, &p.ToolSource, &p.GranteeType,
			&p.GranteeID, &p.Permission, &p.GrantedBy, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning tool permission: %w", err)
		}
		perms = append(perms, &p)
	}
	return perms, nil
}

// Delete removes a permission by ID.
func (r *PermissionsRepository) Delete(ctx context.Context, permID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tool_permissions WHERE id = $1`, permID)
	if err != nil {
		return fmt.Errorf("deleting tool permission: %w", err)
	}
	return nil
}

// DeleteForTool removes all permissions for a tool.
func (r *PermissionsRepository) DeleteForTool(ctx context.Context, toolID uuid.UUID, toolSource string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tool_permissions WHERE tool_id = $1 AND tool_source = $2`, toolID, toolSource)
	if err != nil {
		return fmt.Errorf("deleting tool permissions: %w", err)
	}
	return nil
}

// GetByID retrieves a permission by ID.
func (r *PermissionsRepository) GetByID(ctx context.Context, permID uuid.UUID) (*ToolPermission, error) {
	query := `
		SELECT id, tool_id, tool_source, grantee_type, grantee_id, permission, granted_by, created_at
		FROM tool_permissions WHERE id = $1`

	var p ToolPermission
	err := r.pool.QueryRow(ctx, query, permID).Scan(
		&p.ID, &p.ToolID, &p.ToolSource, &p.GranteeType,
		&p.GranteeID, &p.Permission, &p.GrantedBy, &p.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting permission by ID: %w", err)
	}
	return &p, nil
}
