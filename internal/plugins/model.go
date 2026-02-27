package plugins

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Plugin represents a registered plugin that can provide tools.
type Plugin struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	ManifestURL  string     `json:"manifest_url"`
	IsActive     bool       `json:"is_active"`
	AutoSync     bool       `json:"auto_sync"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	CreatedBy    uuid.UUID  `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Manifest represents a plugin manifest for tool discovery.
type Manifest struct {
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Tools   []ManifestTool `json:"tools"`
}

// ManifestTool represents a tool definition within a plugin manifest.
type ManifestTool struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Parameters   json.RawMessage `json:"parameters"`
	EndpointURL  string          `json:"endpoint_url"`
	HTTPMethod   string          `json:"http_method"`
	AuthType     string          `json:"auth_type"`
	TimeoutSec   int             `json:"timeout_sec"`
	Category     string          `json:"category"`
	OutputSchema json.RawMessage `json:"output_schema"`
}

// RegisterPluginRequest is the request body for registering a plugin.
type RegisterPluginRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	ManifestURL string `json:"manifest_url"`
	AutoSync    bool   `json:"auto_sync"`
	// Inline tool definitions (alternative to manifest URL)
	Tools []ManifestTool `json:"tools"`
}
