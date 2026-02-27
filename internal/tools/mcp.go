package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
)

// MCPTool represents a tool in MCP (Model Context Protocol) format.
type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// MCPManifest represents an MCP tool manifest for export/import.
type MCPManifest struct {
	Schema  string    `json:"schema"`
	Name    string    `json:"name"`
	Version string    `json:"version"`
	Tools   []MCPTool `json:"tools"`
}

// MCPImportRequest is the request body for importing MCP tools.
type MCPImportRequest struct {
	Tools []MCPTool `json:"tools"`
}

// MCPHandler handles MCP tool export/import HTTP requests.
type MCPHandler struct {
	svc *Service
}

// NewMCPHandler creates a new MCP handler.
func NewMCPHandler(svc *Service) *MCPHandler {
	return &MCPHandler{svc: svc}
}

// ExportMCPTools handles GET /agents/{agentID}/tools/mcp
// Exports agent tools in MCP format.
func (h *MCPHandler) ExportMCPTools(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	tools, err := h.svc.ListActiveByAgent(r.Context(), agentID)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusInternalServerError, "failed to list tools")
		return
	}

	mcpTools := make([]MCPTool, 0, len(tools))
	for _, t := range tools {
		mcpTools = append(mcpTools, ToolToMCP(t))
	}

	manifest := MCPManifest{
		Schema:  "mcp/1.0",
		Name:    fmt.Sprintf("agent-%s-tools", agentIDStr),
		Version: "1.0.0",
		Tools:   mcpTools,
	}

	api.JSON(w, http.StatusOK, manifest)
}

// ImportMCPTools handles POST /agents/{agentID}/tools/import
// Imports tools from an MCP manifest into the agent.
func (h *MCPHandler) ImportMCPTools(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	var req MCPImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.JSONErrorMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	imported := make([]*ToolDefinition, 0, len(req.Tools))
	for _, mcpTool := range req.Tools {
		createReq := MCPToCreateRequest(mcpTool)
		tool, err := h.svc.Create(r.Context(), agentID, createReq)
		if err != nil {
			api.JSONErrorMessage(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to import tool '%s': %s", mcpTool.Name, err.Error()))
			return
		}
		imported = append(imported, tool)
	}

	api.JSON(w, http.StatusCreated, map[string]interface{}{
		"imported": len(imported),
		"tools":    imported,
	})
}

// ToolToMCP converts a ToolDefinition to MCP format.
func ToolToMCP(t *ToolDefinition) MCPTool {
	inputSchema := t.Parameters
	if len(inputSchema) == 0 {
		inputSchema = json.RawMessage(`{"type": "object", "properties": {}}`)
	}

	return MCPTool{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: inputSchema,
	}
}

// MCPToToolDefinition converts an MCP tool to a ToolDefinition.
func MCPToToolDefinition(mcp MCPTool, agentID uuid.UUID) *ToolDefinition {
	params := mcp.InputSchema
	if len(params) == 0 {
		params = json.RawMessage(`{}`)
	}

	return &ToolDefinition{
		AgentID:      agentID,
		Name:         mcp.Name,
		Description:  mcp.Description,
		Parameters:   params,
		HTTPMethod:   "POST",
		Headers:      json.RawMessage(`{}`),
		AuthConfig:   json.RawMessage(`{}`),
		OutputSchema: json.RawMessage(`{}`),
		IsActive:     true,
		TimeoutSec:   30,
		Version:      "1.0.0",
		Category:     "imported",
	}
}

// MCPToCreateRequest converts an MCP tool to a CreateToolRequest.
func MCPToCreateRequest(mcp MCPTool) CreateToolRequest {
	params := mcp.InputSchema
	if len(params) == 0 {
		params = json.RawMessage(`{}`)
	}

	return CreateToolRequest{
		Name:        mcp.Name,
		Description: mcp.Description,
		Parameters:  params,
		EndpointURL: "", // MCP tools need endpoint configured separately
		HTTPMethod:  "POST",
		TimeoutSec:  30,
	}
}

// SystemToolToMCP converts a SystemTool to MCP format.
func SystemToolToMCP(t *SystemTool) MCPTool {
	inputSchema := t.Parameters
	if len(inputSchema) == 0 {
		inputSchema = json.RawMessage(`{"type": "object", "properties": {}}`)
	}

	return MCPTool{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: inputSchema,
	}
}

// MCPToSystemToolRequest converts an MCP tool to a CreateSystemToolRequest.
func MCPToSystemToolRequest(mcp MCPTool) CreateSystemToolRequest {
	params := mcp.InputSchema
	if len(params) == 0 {
		params = json.RawMessage(`{}`)
	}

	return CreateSystemToolRequest{
		Name:        mcp.Name,
		Description: mcp.Description,
		Parameters:  params,
		ToolType:    "http",
		HTTPMethod:  "POST",
		TimeoutSec:  30,
		Category:    "imported",
	}
}

// ExportRegistryToMCP exports all registry tools as MCP format.
func ExportRegistryToMCP(ctx context.Context, registry *Registry, agentID uuid.UUID) (*MCPManifest, error) {
	tools, err := registry.GetToolsForAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	mcpTools := make([]MCPTool, 0, len(tools))
	for _, t := range tools {
		inputSchema := t.Parameters
		if len(inputSchema) == 0 {
			inputSchema = json.RawMessage(`{"type": "object", "properties": {}}`)
		}

		mcpTools = append(mcpTools, MCPTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: inputSchema,
		})
	}

	return &MCPManifest{
		Schema:  "mcp/1.0",
		Name:    fmt.Sprintf("agent-%s-tools", agentID.String()),
		Version: "1.0.0",
		Tools:   mcpTools,
	}, nil
}
