//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentCRUD(t *testing.T) {
	env := SetupTestEnv(t)

	RegisterUser(t, env, "agent-crud@example.com", "password123")
	token := LoginUser(t, env, "agent-crud@example.com", "password123")

	var agentID string

	t.Run("create agent", func(t *testing.T) {
		body := map[string]any{
			"name":         "Test Agent",
			"description":  "A test agent",
			"system_prompt": "You are a helpful assistant.",
			"personality_traits": []string{"helpful", "concise"},
			"llm_config": map[string]any{
				"provider":    "openai",
				"model":       "gpt-4-turbo",
				"temperature": 0.7,
			},
			"visibility": "private",
		}

		resp := DoRequest(t, env, "POST", "/api/v1/agents", body, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		result := ParseResponse(t, resp)
		data := result["data"].(map[string]any)

		assert.NotEmpty(t, data["id"])
		assert.NotEmpty(t, data["jid"])
		assert.Equal(t, "private", data["visibility"])

		profile := data["profile"].(map[string]any)
		assert.Equal(t, "Test Agent", profile["name"])
		assert.Equal(t, "A test agent", profile["description"])
		assert.Equal(t, "You are a helpful assistant.", profile["system_prompt"])

		agentID = data["id"].(string)
	})

	t.Run("list agents", func(t *testing.T) {
		resp := DoRequest(t, env, "GET", "/api/v1/agents", nil, token)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseResponse(t, resp)
		data := result["data"].([]any)
		assert.GreaterOrEqual(t, len(data), 1)
		assert.NotZero(t, result["total_count"])
	})

	t.Run("get agent", func(t *testing.T) {
		resp := DoRequest(t, env, "GET", "/api/v1/agents/"+agentID, nil, token)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseResponse(t, resp)
		data := result["data"].(map[string]any)
		assert.Equal(t, agentID, data["id"])

		profile := data["profile"].(map[string]any)
		// system_prompt should be decrypted in response
		assert.Equal(t, "You are a helpful assistant.", profile["system_prompt"])
	})

	t.Run("update agent", func(t *testing.T) {
		body := map[string]any{
			"name":          "Updated Agent",
			"description":   "An updated test agent",
			"system_prompt": "You are an updated assistant.",
		}

		resp := DoRequest(t, env, "PUT", "/api/v1/agents/"+agentID, body, token)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseResponse(t, resp)
		data := result["data"].(map[string]any)
		profile := data["profile"].(map[string]any)
		assert.Equal(t, "Updated Agent", profile["name"])
		assert.Equal(t, "An updated test agent", profile["description"])
		assert.Equal(t, "You are an updated assistant.", profile["system_prompt"])
	})

	t.Run("delete agent (soft)", func(t *testing.T) {
		resp := DoRequest(t, env, "DELETE", "/api/v1/agents/"+agentID, nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Should not appear in list
		listResp := DoRequest(t, env, "GET", "/api/v1/agents", nil, token)
		listResult := ParseResponse(t, listResp)
		data := listResult["data"].([]any)
		for _, a := range data {
			agent := a.(map[string]any)
			assert.NotEqual(t, agentID, agent["id"])
		}
	})

	t.Run("get deleted agent returns 404", func(t *testing.T) {
		resp := DoRequest(t, env, "GET", "/api/v1/agents/"+agentID, nil, token)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestAgentValidation(t *testing.T) {
	env := SetupTestEnv(t)

	RegisterUser(t, env, "agent-val@example.com", "password123")
	token := LoginUser(t, env, "agent-val@example.com", "password123")

	t.Run("missing name", func(t *testing.T) {
		body := map[string]any{
			"system_prompt": "You are a helper.",
		}
		resp := DoRequest(t, env, "POST", "/api/v1/agents", body, token)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing system_prompt", func(t *testing.T) {
		body := map[string]any{
			"name": "No Prompt Agent",
		}
		resp := DoRequest(t, env, "POST", "/api/v1/agents", body, token)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid agent ID", func(t *testing.T) {
		resp := DoRequest(t, env, "GET", "/api/v1/agents/not-a-uuid", nil, token)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAgentJIDGeneration(t *testing.T) {
	env := SetupTestEnv(t)

	RegisterUser(t, env, "agent-jid@example.com", "password123")
	token := LoginUser(t, env, "agent-jid@example.com", "password123")

	body := map[string]any{
		"name":          "JID Agent",
		"system_prompt": "Test prompt",
	}

	resp := DoRequest(t, env, "POST", "/api/v1/agents", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	result := ParseResponse(t, resp)
	data := result["data"].(map[string]any)
	jid := data["jid"].(string)

	// JID should match format: agent-<uuid>@agents.<domain>
	assert.Contains(t, jid, "agent-")
	assert.Contains(t, jid, "@agents.test.aiox.local")
}

func TestAgentSystemPromptEncryption(t *testing.T) {
	env := SetupTestEnv(t)

	RegisterUser(t, env, "agent-enc@example.com", "password123")
	token := LoginUser(t, env, "agent-enc@example.com", "password123")

	body := map[string]any{
		"name":          "Encrypted Agent",
		"system_prompt": "Super secret prompt that should be encrypted",
	}

	resp := DoRequest(t, env, "POST", "/api/v1/agents", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	result := ParseResponse(t, resp)
	data := result["data"].(map[string]any)
	agentID := data["id"].(string)

	// Verify in DB that it's actually encrypted
	var profileBytes []byte
	err := env.Pool.QueryRow(
		t.Context(),
		"SELECT profile FROM agents WHERE id = $1",
		agentID,
	).Scan(&profileBytes)
	require.NoError(t, err)

	var dbProfile map[string]any
	require.NoError(t, json.Unmarshal(profileBytes, &dbProfile))

	// The stored system_prompt should NOT be the plaintext
	assert.NotEqual(t, "Super secret prompt that should be encrypted", dbProfile["system_prompt"])
	assert.Equal(t, true, dbProfile["encrypted"])

	// But the API response should return it decrypted
	getResp := DoRequest(t, env, "GET", "/api/v1/agents/"+agentID, nil, token)
	getResult := ParseResponse(t, getResp)
	getData := getResult["data"].(map[string]any)
	profile := getData["profile"].(map[string]any)
	assert.Equal(t, "Super secret prompt that should be encrypted", profile["system_prompt"])
}
