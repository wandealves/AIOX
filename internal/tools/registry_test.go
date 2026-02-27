package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvedTool_Sources(t *testing.T) {
	rt := &ResolvedTool{
		Name:   "test-tool",
		Source: "system",
	}
	assert.Equal(t, "system", rt.Source)
	assert.Equal(t, "test-tool", rt.Name)
}

func TestResolvedTool_AgentSource(t *testing.T) {
	rt := &ResolvedTool{
		Name:     "agent-tool",
		Source:   "agent",
		Version:  "2.0.0",
		Category: "custom",
		ToolType: "http",
	}
	assert.Equal(t, "agent", rt.Source)
	assert.Equal(t, "2.0.0", rt.Version)
	assert.Equal(t, "custom", rt.Category)
	assert.Equal(t, "http", rt.ToolType)
}

func TestMCPToolConversion(t *testing.T) {
	mcp := MCPTool{
		Name:        "test-tool",
		Description: "A test tool",
	}

	req := MCPToCreateRequest(mcp)
	assert.Equal(t, "test-tool", req.Name)
	assert.Equal(t, "A test tool", req.Description)
	assert.Equal(t, "POST", req.HTTPMethod)
	assert.Equal(t, 30, req.TimeoutSec)
}

func TestMCPToolBidirectional(t *testing.T) {
	// Create a tool definition and convert to MCP
	td := &ToolDefinition{
		Name:        "my-tool",
		Description: "My description",
	}

	mcpTool := ToolToMCP(td)
	assert.Equal(t, "my-tool", mcpTool.Name)
	assert.Equal(t, "My description", mcpTool.Description)

	// Convert back
	req := MCPToCreateRequest(mcpTool)
	assert.Equal(t, td.Name, req.Name)
	assert.Equal(t, td.Description, req.Description)
}

func TestMCPToSystemToolRequest(t *testing.T) {
	mcp := MCPTool{
		Name:        "sys-tool",
		Description: "A system tool",
	}

	req := MCPToSystemToolRequest(mcp)
	assert.Equal(t, "sys-tool", req.Name)
	assert.Equal(t, "A system tool", req.Description)
	assert.Equal(t, "http", req.ToolType)
	assert.Equal(t, "imported", req.Category)
}
