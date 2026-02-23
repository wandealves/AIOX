package orgs

import (
	"time"

	"github.com/google/uuid"
)

// Org roles ordered by permission level.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// roleLevel returns numeric level for role hierarchy comparisons.
func roleLevel(role string) int {
	switch role {
	case RoleOwner:
		return 4
	case RoleAdmin:
		return 3
	case RoleEditor:
		return 2
	case RoleViewer:
		return 1
	default:
		return 0
	}
}

// HasPermission checks if the given role meets or exceeds the required level.
func HasPermission(memberRole, requiredRole string) bool {
	return roleLevel(memberRole) >= roleLevel(requiredRole)
}

// Organization represents an organization.
type Organization struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	Settings    []byte     `json:"settings"`
	QuotaConfig []byte     `json:"quota_config"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Member represents an organization member.
type Member struct {
	ID       uuid.UUID `json:"id"`
	OrgID    uuid.UUID `json:"org_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"`
	Email    string    `json:"email,omitempty"`
	JoinedAt time.Time `json:"joined_at"`
}

// Invite represents a pending organization invitation.
type Invite struct {
	ID         uuid.UUID  `json:"id"`
	OrgID      uuid.UUID  `json:"org_id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Token      string     `json:"token"`
	InvitedBy  uuid.UUID  `json:"invited_by"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Request types

type CreateOrgRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Slug        string `json:"slug" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateOrgRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin editor viewer"`
}
