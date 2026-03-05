package tools

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ToolDefinition represents an external tool that an agent can invoke.
type ToolDefinition struct {
	ID           uuid.UUID       `json:"id"`
	AgentID      uuid.UUID       `json:"agent_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Parameters   json.RawMessage `json:"parameters"`
	EndpointURL  string          `json:"endpoint_url"`
	HTTPMethod   string          `json:"http_method"`
	Headers      json.RawMessage `json:"headers"`
	AuthType     string          `json:"auth_type"`
	AuthConfig   json.RawMessage `json:"auth_config"`
	IsActive     bool            `json:"is_active"`
	TimeoutSec   int             `json:"timeout_sec"`
	Version      string          `json:"version"`
	Category     string          `json:"category"`
	OutputSchema json.RawMessage `json:"output_schema"`
	ToolType     string          `json:"tool_type"`
	UTCPManualID *uuid.UUID      `json:"utcp_manual_id,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CreateToolRequest is the request body for creating a tool.
type CreateToolRequest struct {
	Name        string          `json:"name" validate:"required,min=1,max=100"`
	Description string          `json:"description" validate:"max=500"`
	Parameters  json.RawMessage `json:"parameters"`
	EndpointURL string          `json:"endpoint_url" validate:"required,url"`
	HTTPMethod  string          `json:"http_method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	Headers     json.RawMessage `json:"headers"`
	AuthType    string          `json:"auth_type" validate:"oneof='' bearer api_key"`
	AuthConfig  json.RawMessage `json:"auth_config"`
	TimeoutSec  int             `json:"timeout_sec" validate:"min=1,max=300"`
}

// UpdateToolRequest is the request body for updating a tool.
type UpdateToolRequest struct {
	Name        *string          `json:"name" validate:"omitempty,min=1,max=100"`
	Description *string          `json:"description" validate:"omitempty,max=500"`
	Parameters  *json.RawMessage `json:"parameters"`
	EndpointURL *string          `json:"endpoint_url" validate:"omitempty,url"`
	HTTPMethod  *string          `json:"http_method" validate:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	Headers     *json.RawMessage `json:"headers"`
	AuthType    *string          `json:"auth_type" validate:"omitempty,oneof='' bearer api_key"`
	AuthConfig  *json.RawMessage `json:"auth_config"`
	IsActive    *bool            `json:"is_active"`
	TimeoutSec  *int             `json:"timeout_sec" validate:"omitempty,min=1,max=300"`
}
