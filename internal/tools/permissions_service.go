package tools

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PermissionsService provides business logic for tool permissions.
type PermissionsService struct {
	repo *PermissionsRepository
}

// NewPermissionsService creates a new permissions service.
func NewPermissionsService(repo *PermissionsRepository) *PermissionsService {
	return &PermissionsService{repo: repo}
}

// Grant creates or updates a permission for a tool.
func (s *PermissionsService) Grant(ctx context.Context, req CreatePermissionRequest, grantedBy uuid.UUID) (*ToolPermission, error) {
	perm := &ToolPermission{
		ID:          uuid.New(),
		ToolID:      req.ToolID,
		ToolSource:  req.ToolSource,
		GranteeType: req.GranteeType,
		GranteeID:   req.GranteeID,
		Permission:  req.Permission,
		GrantedBy:   grantedBy,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

// HasPermission checks if an entity has permission to use a tool.
func (s *PermissionsService) HasPermission(ctx context.Context, toolID uuid.UUID, toolSource, granteeType, granteeID, permission string) (bool, error) {
	return s.repo.HasPermission(ctx, toolID, toolSource, granteeType, granteeID, permission)
}

// ListForTool returns all permissions for a tool.
func (s *PermissionsService) ListForTool(ctx context.Context, toolID uuid.UUID, toolSource string) ([]*ToolPermission, error) {
	return s.repo.ListForTool(ctx, toolID, toolSource)
}

// ListForGrantee returns all permissions for a grantee.
func (s *PermissionsService) ListForGrantee(ctx context.Context, granteeType, granteeID string) ([]*ToolPermission, error) {
	return s.repo.ListForGrantee(ctx, granteeType, granteeID)
}

// Revoke removes a permission.
func (s *PermissionsService) Revoke(ctx context.Context, permID uuid.UUID) error {
	return s.repo.Delete(ctx, permID)
}

// RevokeAllForTool removes all permissions for a tool.
func (s *PermissionsService) RevokeAllForTool(ctx context.Context, toolID uuid.UUID, toolSource string) error {
	return s.repo.DeleteForTool(ctx, toolID, toolSource)
}

// CanExecute is a convenience method that checks "execute" permission.
// System tools are accessible by default; agent tools require explicit permission.
func (s *PermissionsService) CanExecute(ctx context.Context, toolID uuid.UUID, toolSource string, userID uuid.UUID, roles []string) (bool, error) {
	// System tools are accessible by default
	if toolSource == "system" {
		return true, nil
	}

	// Check user-level permission
	has, err := s.repo.HasPermission(ctx, toolID, toolSource, "agent", userID.String(), "execute")
	if err != nil {
		return false, err
	}
	if has {
		return true, nil
	}

	// Check role-level permissions
	for _, role := range roles {
		has, err := s.repo.HasPermission(ctx, toolID, toolSource, "role", role, "execute")
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}

	return false, nil
}
