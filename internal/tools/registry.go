package tools

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// ResolvedTool represents a tool resolved from either system or agent scope.
type ResolvedTool struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Parameters   json.RawMessage `json:"parameters"`
	OutputSchema json.RawMessage `json:"output_schema"`
	EndpointURL  string          `json:"endpoint_url"`
	HTTPMethod   string          `json:"http_method"`
	Headers      json.RawMessage `json:"headers"`
	AuthType     string          `json:"auth_type"`
	AuthConfig   json.RawMessage `json:"auth_config"`
	IsActive     bool            `json:"is_active"`
	TimeoutSec   int             `json:"timeout_sec"`
	Version      string          `json:"version"`
	Category     string          `json:"category"`
	ToolType     string          `json:"tool_type"`
	Source       string          `json:"source"` // "system" | "agent"
}

// Registry merges system tools and agent tools into a unified view.
type Registry struct {
	systemSvc *SystemService
	agentSvc  *Service
}

// NewRegistry creates a new tool registry.
func NewRegistry(systemSvc *SystemService, agentSvc *Service) *Registry {
	return &Registry{
		systemSvc: systemSvc,
		agentSvc:  agentSvc,
	}
}

// GetToolsForAgent returns all tools available to an agent, merging system and agent tools.
// Agent tools with the same name override system tools.
func (r *Registry) GetToolsForAgent(ctx context.Context, agentID uuid.UUID) ([]*ResolvedTool, error) {
	toolMap := make(map[string]*ResolvedTool)

	// First, add all active system tools
	if r.systemSvc != nil {
		systemTools, err := r.systemSvc.ListActive(ctx)
		if err != nil {
			return nil, err
		}
		for _, st := range systemTools {
			toolMap[st.Name] = &ResolvedTool{
				ID:           st.ID,
				Name:         st.Name,
				Description:  st.Description,
				Parameters:   st.Parameters,
				OutputSchema: st.OutputSchema,
				EndpointURL:  st.EndpointURL,
				HTTPMethod:   st.HTTPMethod,
				Headers:      st.Headers,
				AuthType:     st.AuthType,
				AuthConfig:   st.AuthConfig,
				IsActive:     st.IsActive,
				TimeoutSec:   st.TimeoutSec,
				Version:      st.Version,
				Category:     st.Category,
				ToolType:     st.ToolType,
				Source:       "system",
			}
		}
	}

	// Then, add agent tools — they override system tools with the same name
	if r.agentSvc != nil {
		agentTools, err := r.agentSvc.ListActiveByAgent(ctx, agentID)
		if err != nil {
			return nil, err
		}
		for _, at := range agentTools {
			toolType := "http"
			toolMap[at.Name] = &ResolvedTool{
				ID:           at.ID,
				Name:         at.Name,
				Description:  at.Description,
				Parameters:   at.Parameters,
				OutputSchema: at.OutputSchema,
				EndpointURL:  at.EndpointURL,
				HTTPMethod:   at.HTTPMethod,
				Headers:      at.Headers,
				AuthType:     at.AuthType,
				AuthConfig:   at.AuthConfig,
				IsActive:     at.IsActive,
				TimeoutSec:   at.TimeoutSec,
				Version:      at.Version,
				Category:     at.Category,
				ToolType:     toolType,
				Source:       "agent",
			}
		}
	}

	result := make([]*ResolvedTool, 0, len(toolMap))
	for _, t := range toolMap {
		result = append(result, t)
	}
	return result, nil
}

// GetSystemTools returns all active system tools.
func (r *Registry) GetSystemTools(ctx context.Context) ([]*ResolvedTool, error) {
	if r.systemSvc == nil {
		return nil, nil
	}

	systemTools, err := r.systemSvc.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*ResolvedTool, 0, len(systemTools))
	for _, st := range systemTools {
		result = append(result, &ResolvedTool{
			ID:           st.ID,
			Name:         st.Name,
			Description:  st.Description,
			Parameters:   st.Parameters,
			OutputSchema: st.OutputSchema,
			EndpointURL:  st.EndpointURL,
			HTTPMethod:   st.HTTPMethod,
			Headers:      st.Headers,
			AuthType:     st.AuthType,
			AuthConfig:   st.AuthConfig,
			IsActive:     st.IsActive,
			TimeoutSec:   st.TimeoutSec,
			Version:      st.Version,
			Category:     st.Category,
			ToolType:     st.ToolType,
			Source:       "system",
		})
	}
	return result, nil
}
