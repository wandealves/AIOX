package plugins

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManifest_Parse(t *testing.T) {
	data := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"tools": [
			{
				"name": "tool-a",
				"description": "Tool A",
				"endpoint_url": "https://example.com/a",
				"http_method": "POST",
				"timeout_sec": 10
			},
			{
				"name": "tool-b",
				"description": "Tool B",
				"endpoint_url": "https://example.com/b"
			}
		]
	}`

	var manifest Manifest
	err := json.Unmarshal([]byte(data), &manifest)
	assert.NoError(t, err)
	assert.Equal(t, "test-plugin", manifest.Name)
	assert.Equal(t, "1.0.0", manifest.Version)
	assert.Len(t, manifest.Tools, 2)
	assert.Equal(t, "tool-a", manifest.Tools[0].Name)
	assert.Equal(t, "POST", manifest.Tools[0].HTTPMethod)
	assert.Equal(t, 10, manifest.Tools[0].TimeoutSec)
}

func TestManifest_EmptyTools(t *testing.T) {
	data := `{"name": "empty", "version": "0.1.0", "tools": []}`
	var manifest Manifest
	err := json.Unmarshal([]byte(data), &manifest)
	assert.NoError(t, err)
	assert.Len(t, manifest.Tools, 0)
}

func TestManifestTool_Defaults(t *testing.T) {
	data := `{"name": "minimal", "description": "A minimal tool"}`
	var mt ManifestTool
	err := json.Unmarshal([]byte(data), &mt)
	assert.NoError(t, err)
	assert.Equal(t, "minimal", mt.Name)
	assert.Equal(t, "A minimal tool", mt.Description)
	assert.Equal(t, "", mt.HTTPMethod)
	assert.Equal(t, 0, mt.TimeoutSec)
	assert.Equal(t, "", mt.Category)
}

func TestRegisterPluginRequest_InlineTools(t *testing.T) {
	req := RegisterPluginRequest{
		Name:        "my-plugin",
		Description: "Plugin with inline tools",
		Tools: []ManifestTool{
			{
				Name:        "inline-tool",
				Description: "Inline tool",
				EndpointURL: "https://example.com/inline",
				HTTPMethod:  "GET",
			},
		},
	}
	assert.Equal(t, "my-plugin", req.Name)
	assert.Len(t, req.Tools, 1)
	assert.Equal(t, "inline-tool", req.Tools[0].Name)
	assert.Empty(t, req.ManifestURL)
}

func TestRegisterPluginRequest_ManifestURL(t *testing.T) {
	req := RegisterPluginRequest{
		Name:        "remote-plugin",
		ManifestURL: "https://example.com/manifest.json",
		AutoSync:    true,
	}
	assert.Equal(t, "remote-plugin", req.Name)
	assert.Equal(t, "https://example.com/manifest.json", req.ManifestURL)
	assert.True(t, req.AutoSync)
	assert.Len(t, req.Tools, 0)
}

func TestPlugin_Struct(t *testing.T) {
	plugin := Plugin{
		Name:        "test",
		Description: "Test plugin",
		ManifestURL: "https://example.com/manifest.json",
		IsActive:    true,
		AutoSync:    false,
	}
	assert.Equal(t, "test", plugin.Name)
	assert.True(t, plugin.IsActive)
	assert.False(t, plugin.AutoSync)
	assert.Nil(t, plugin.LastSyncedAt)
}

func TestManifestTool_WithParameters(t *testing.T) {
	data := `{
		"name": "parameterized",
		"description": "Has params",
		"parameters": {"type": "object", "properties": {"query": {"type": "string"}}},
		"output_schema": {"type": "object"},
		"category": "search"
	}`
	var mt ManifestTool
	err := json.Unmarshal([]byte(data), &mt)
	assert.NoError(t, err)
	assert.Equal(t, "parameterized", mt.Name)
	assert.Equal(t, "search", mt.Category)
	assert.NotNil(t, mt.Parameters)
	assert.NotNil(t, mt.OutputSchema)
}
