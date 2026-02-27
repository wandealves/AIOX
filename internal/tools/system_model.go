package tools

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SystemTool represents a global tool available to all agents.
type SystemTool struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Category     string          `json:"category"`
	Parameters   json.RawMessage `json:"parameters"`
	OutputSchema json.RawMessage `json:"output_schema"`
	ToolType     string          `json:"tool_type"` // "builtin" | "http"
	EndpointURL  string          `json:"endpoint_url"`
	HTTPMethod   string          `json:"http_method"`
	Headers      json.RawMessage `json:"headers"`
	AuthType     string          `json:"auth_type"`
	AuthConfig   json.RawMessage `json:"auth_config"`
	IsActive     bool            `json:"is_active"`
	TimeoutSec   int             `json:"timeout_sec"`
	Version      string          `json:"version"`
	PluginID     *uuid.UUID      `json:"plugin_id,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CreateSystemToolRequest is the request body for creating a system tool.
type CreateSystemToolRequest struct {
	Name         string          `json:"name" validate:"required,min=1,max=100"`
	Description  string          `json:"description" validate:"max=500"`
	Category     string          `json:"category" validate:"max=50"`
	Parameters   json.RawMessage `json:"parameters"`
	OutputSchema json.RawMessage `json:"output_schema"`
	ToolType     string          `json:"tool_type" validate:"oneof=builtin http"`
	EndpointURL  string          `json:"endpoint_url"`
	HTTPMethod   string          `json:"http_method" validate:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	Headers      json.RawMessage `json:"headers"`
	AuthType     string          `json:"auth_type" validate:"omitempty,oneof='' bearer api_key"`
	AuthConfig   json.RawMessage `json:"auth_config"`
	TimeoutSec   int             `json:"timeout_sec" validate:"omitempty,min=1,max=300"`
	Version      string          `json:"version" validate:"max=20"`
}

// UpdateSystemToolRequest is the request body for updating a system tool.
type UpdateSystemToolRequest struct {
	Name         *string          `json:"name" validate:"omitempty,min=1,max=100"`
	Description  *string          `json:"description" validate:"omitempty,max=500"`
	Category     *string          `json:"category" validate:"omitempty,max=50"`
	Parameters   *json.RawMessage `json:"parameters"`
	OutputSchema *json.RawMessage `json:"output_schema"`
	ToolType     *string          `json:"tool_type" validate:"omitempty,oneof=builtin http"`
	EndpointURL  *string          `json:"endpoint_url"`
	HTTPMethod   *string          `json:"http_method" validate:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	Headers      *json.RawMessage `json:"headers"`
	AuthType     *string          `json:"auth_type" validate:"omitempty,oneof='' bearer api_key"`
	AuthConfig   *json.RawMessage `json:"auth_config"`
	IsActive     *bool            `json:"is_active"`
	TimeoutSec   *int             `json:"timeout_sec" validate:"omitempty,min=1,max=300"`
	Version      *string          `json:"version" validate:"omitempty,max=20"`
}
