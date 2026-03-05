package utcp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UTCPManual represents a connected UTCP manual for an agent.
type UTCPManual struct {
	ID           uuid.UUID       `json:"id"`
	AgentID      uuid.UUID       `json:"agent_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	ManualURL    string          `json:"manual_url"`
	UTCPVersion  string          `json:"utcp_version"`
	AuthTemplate json.RawMessage `json:"auth_template"`
	AuthValues   string          `json:"auth_values"` // encrypted
	IsActive     bool            `json:"is_active"`
	LastSyncedAt *time.Time      `json:"last_synced_at,omitempty"`
	ToolsCount   int             `json:"tools_count"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// ConnectManualRequest is the request body for connecting a UTCP manual.
type ConnectManualRequest struct {
	Name       string            `json:"name"`
	ManualURL  string            `json:"manual_url"`
	AuthValues map[string]string `json:"auth_values,omitempty"`
}

// DiscoverRequest is the request body for discovering tools from a UTCP manual URL.
type DiscoverRequest struct {
	ManualURL string `json:"manual_url"`
}

// UTCPManualJSON represents the parsed remote UTCP manual document.
type UTCPManualJSON struct {
	ManualVersion string         `json:"manual_version"`
	UTCPVersion   string         `json:"utcp_version"`
	Auth          *UTCPAuthSpec  `json:"auth,omitempty"`
	Tools         []UTCPToolSpec `json:"tools"`

	// Extended fields
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	BaseURL     string `json:"base_url"`
}

// UTCPAuthSpec describes authentication requirements in a UTCP manual.
type UTCPAuthSpec struct {
	// Common fields
	AuthType    string `json:"auth_type"`    // "api_key", "basic", "bearer"
	Location    string `json:"location"`     // "header", "query", "body"
	VarName     string `json:"var_name"`     // header/query parameter name
	Prefix      string `json:"prefix"`       // e.g. "Bearer"
	Description string `json:"description"`

	// api_key specific (official format)
	APIKey string `json:"api_key"` // e.g. "${API_KEY}"

	// Simplified format
	Type string            `json:"type"` // fallback if auth_type is empty
	Vars map[string]string `json:"vars"` // variable name -> description
}

// GetAuthType returns the effective auth type.
func (a *UTCPAuthSpec) GetAuthType() string {
	if a.AuthType != "" {
		return a.AuthType
	}
	return a.Type
}

// UTCPToolProvider describes how to invoke a tool (the real UTCP field).
type UTCPToolProvider struct {
	ProviderType string        `json:"provider_type"` // "http"
	Name         string        `json:"name"`
	HTTPMethod   string        `json:"http_method"`
	URL          string        `json:"url"`
	ContentType  string        `json:"content_type"`
	Auth         *UTCPAuthSpec `json:"auth,omitempty"` // per-tool auth override
}

// UTCPToolCallTemplate is the alternate format from utcp.io docs.
type UTCPToolCallTemplate struct {
	CallTemplateType string `json:"call_template_type"`
	URL              string `json:"url"`
	HTTPMethod       string `json:"http_method"`
}

// UTCPToolSpec describes a single tool in a UTCP manual.
type UTCPToolSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`

	// Real UTCP format
	Inputs       json.RawMessage       `json:"inputs,omitempty"`
	ToolProvider *UTCPToolProvider      `json:"tool_provider,omitempty"`

	// Alternate format (utcp.io docs)
	ToolCallTemplate *UTCPToolCallTemplate `json:"tool_call_template,omitempty"`

	// Simplified format
	Path       string          `json:"path,omitempty"`
	Method     string          `json:"method,omitempty"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
	Headers    json.RawMessage `json:"headers,omitempty"`
	TimeoutSec int             `json:"timeout_sec,omitempty"`
}

// GetEndpointURL returns the tool's endpoint URL, checking all format variants.
func (t *UTCPToolSpec) GetEndpointURL(baseURL string) string {
	if t.ToolProvider != nil && t.ToolProvider.URL != "" {
		return t.ToolProvider.URL
	}
	if t.ToolCallTemplate != nil && t.ToolCallTemplate.URL != "" {
		return t.ToolCallTemplate.URL
	}
	if t.Path != "" && baseURL != "" {
		return baseURL + "/" + trimSlashes(t.Path)
	}
	if t.Path != "" {
		return t.Path
	}
	return ""
}

// GetHTTPMethod returns the tool's HTTP method, checking all format variants.
func (t *UTCPToolSpec) GetHTTPMethod() string {
	if t.ToolProvider != nil && t.ToolProvider.HTTPMethod != "" {
		return t.ToolProvider.HTTPMethod
	}
	if t.ToolCallTemplate != nil && t.ToolCallTemplate.HTTPMethod != "" {
		return t.ToolCallTemplate.HTTPMethod
	}
	if t.Method != "" {
		return t.Method
	}
	return "POST"
}

// GetParameters returns the tool's parameters from either format.
func (t *UTCPToolSpec) GetParameters() json.RawMessage {
	if len(t.Inputs) > 0 {
		return t.Inputs
	}
	if len(t.Parameters) > 0 {
		return t.Parameters
	}
	return json.RawMessage(`{}`)
}

// GetToolAuth returns per-tool auth override if present.
func (t *UTCPToolSpec) GetToolAuth() *UTCPAuthSpec {
	if t.ToolProvider != nil && t.ToolProvider.Auth != nil {
		return t.ToolProvider.Auth
	}
	return nil
}

func trimSlashes(s string) string {
	for len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// DiscoverResult is the response for the discover endpoint.
type DiscoverResult struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	AuthVars    []string       `json:"auth_vars,omitempty"`
	Tools       []UTCPToolSpec `json:"tools"`
}
