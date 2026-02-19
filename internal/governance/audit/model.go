package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog matches the audit_logs table schema.
type AuditLog struct {
	ID           uuid.UUID       `json:"id"`
	OwnerUserID  uuid.UUID       `json:"owner_user_id"`
	EventType    string          `json:"event_type"`
	Severity     string          `json:"severity"`
	ResourceType string          `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID      `json:"resource_id,omitempty"`
	Details      json.RawMessage `json:"details,omitempty"`
	IPAddress    string          `json:"ip_address,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// ListParams holds pagination and filtering parameters for audit log queries.
type ListParams struct {
	EventType string
	Severity  string
	From      *time.Time
	To        *time.Time
	Page      int
	PageSize  int
}

// DefaultListParams returns sensible defaults.
func DefaultListParams() ListParams {
	return ListParams{
		Page:     1,
		PageSize: 20,
	}
}
