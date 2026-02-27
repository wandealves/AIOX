package tools

import (
	"time"

	"github.com/google/uuid"
)

// ToolPermission represents a permission grant for a tool.
type ToolPermission struct {
	ID          uuid.UUID `json:"id"`
	ToolID      uuid.UUID `json:"tool_id"`
	ToolSource  string    `json:"tool_source"`  // "system" | "agent"
	GranteeType string    `json:"grantee_type"` // "agent" | "role" | "org"
	GranteeID   string    `json:"grantee_id"`
	Permission  string    `json:"permission"` // "execute" | "configure" | "admin"
	GrantedBy   uuid.UUID `json:"granted_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreatePermissionRequest is the request body for creating a tool permission.
type CreatePermissionRequest struct {
	ToolID      uuid.UUID `json:"tool_id" validate:"required"`
	ToolSource  string    `json:"tool_source" validate:"required,oneof=system agent"`
	GranteeType string    `json:"grantee_type" validate:"required,oneof=agent role org"`
	GranteeID   string    `json:"grantee_id" validate:"required"`
	Permission  string    `json:"permission" validate:"required,oneof=execute configure admin"`
}
