package tools

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateToolInput_NoSchema(t *testing.T) {
	err := ValidateToolInput(json.RawMessage(`{}`), json.RawMessage(`{"key": "value"}`))
	assert.NoError(t, err)
}

func TestValidateToolInput_EmptySchema(t *testing.T) {
	err := ValidateToolInput(nil, json.RawMessage(`{"key": "value"}`))
	assert.NoError(t, err)
}

func TestValidateToolInput_RequiredFieldPresent(t *testing.T) {
	schema := json.RawMessage(`{
		"type": "object",
		"required": ["name"],
		"properties": {
			"name": {"type": "string"}
		}
	}`)
	input := json.RawMessage(`{"name": "test"}`)
	err := ValidateToolInput(schema, input)
	assert.NoError(t, err)
}

func TestValidateToolInput_RequiredFieldMissing(t *testing.T) {
	schema := json.RawMessage(`{
		"type": "object",
		"required": ["name"],
		"properties": {
			"name": {"type": "string"}
		}
	}`)
	input := json.RawMessage(`{"other": "value"}`)
	err := ValidateToolInput(schema, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: name")
}

func TestValidateToolInput_TypeMismatch(t *testing.T) {
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"count": {"type": "integer"}
		}
	}`)
	input := json.RawMessage(`{"count": "not a number"}`)
	err := ValidateToolInput(schema, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count")
}

func TestValidateToolInput_TypeMatch(t *testing.T) {
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"count": {"type": "number"},
			"name": {"type": "string"},
			"active": {"type": "boolean"}
		}
	}`)
	input := json.RawMessage(`{"count": 42, "name": "test", "active": true}`)
	err := ValidateToolInput(schema, input)
	assert.NoError(t, err)
}

func TestValidateToolOutput_NoSchema(t *testing.T) {
	err := ValidateToolOutput(json.RawMessage(`{}`), json.RawMessage(`{"result": "ok"}`))
	assert.NoError(t, err)
}

func TestValidateToolOutput_EmptyOutput(t *testing.T) {
	schema := json.RawMessage(`{"type": "object"}`)
	err := ValidateToolOutput(schema, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "output is empty")
}

func TestValidateToolOutput_TypeMatch(t *testing.T) {
	schema := json.RawMessage(`{"type": "object"}`)
	output := json.RawMessage(`{"key": "value"}`)
	err := ValidateToolOutput(schema, output)
	assert.NoError(t, err)
}

func TestValidateToolOutput_TypeMismatch(t *testing.T) {
	schema := json.RawMessage(`{"type": "array"}`)
	output := json.RawMessage(`{"key": "value"}`)
	err := ValidateToolOutput(schema, output)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type mismatch")
}

func TestValidateToolName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "my-tool_v1", false},
		{"empty name", "", true},
		{"with spaces", "my tool", true},
		{"with special chars", "my@tool!", true},
		{"too long", string(make([]byte, 101)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid semver", "1.0.0", false},
		{"empty is ok", "", false},
		{"missing patch", "1.0", true},
		{"just major", "1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeToolResponse_Success(t *testing.T) {
	result := NormalizeToolResponse(`{"data": "hello"}`, 100, "")
	var resp map[string]interface{}
	err := json.Unmarshal([]byte(result), &resp)
	require.NoError(t, err)
	assert.Equal(t, "success", resp["status"])
	assert.NotNil(t, resp["data"])
	assert.Nil(t, resp["error"])
	assert.Equal(t, float64(100), resp["execution_time_ms"])
}

func TestNormalizeToolResponse_Error(t *testing.T) {
	result := NormalizeToolResponse("", 50, "something failed")
	var resp map[string]interface{}
	err := json.Unmarshal([]byte(result), &resp)
	require.NoError(t, err)
	assert.Equal(t, "error", resp["status"])
	assert.Equal(t, "something failed", resp["error"])
}
