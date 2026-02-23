//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListConversations_Empty(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("chat-empty-%s@test.com", uuid.New().String()[:8])
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	// Create an agent
	agentBody := map[string]any{
		"name":          "Chat Agent",
		"description":   "Test agent for chat",
		"system_prompt": "You are a helpful assistant",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	result := ParseResponse(t, resp)
	agentData := result["data"].(map[string]any)
	agentID := agentData["id"].(string)

	// List conversations — should be empty
	resp = DoRequest(t, env, "GET", "/api/v1/agents/"+agentID+"/conversations", nil, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	result = ParseResponse(t, resp)
	data := result["data"]
	if data == nil {
		// nil means empty list, which is expected
		return
	}
	items := data.([]any)
	assert.Empty(t, items)
}

func TestListConversations_Pagination(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("chat-page-%s@test.com", uuid.New().String()[:8])
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	// Create an agent
	agentBody := map[string]any{
		"name":          "Chat Agent Pagination",
		"description":   "Test agent for pagination",
		"system_prompt": "You are a helpful assistant",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	result := ParseResponse(t, resp)
	agentData := result["data"].(map[string]any)
	agentID := agentData["id"].(string)

	// List with pagination params
	resp = DoRequest(t, env, "GET", "/api/v1/agents/"+agentID+"/conversations?page=1&page_size=5", nil, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	result = ParseResponse(t, resp)
	assert.Equal(t, float64(1), result["page"])
	assert.Equal(t, float64(5), result["page_size"])
	assert.Equal(t, float64(0), result["total_count"])
}

func TestListConversations_OwnershipIsolation(t *testing.T) {
	env := SetupTestEnv(t)

	// User 1 creates an agent
	email1 := fmt.Sprintf("chat-own1-%s@test.com", uuid.New().String()[:8])
	RegisterUser(t, env, email1, "password123")
	token1 := LoginUser(t, env, email1, "password123")

	agentBody := map[string]any{
		"name":          "Owned Agent",
		"description":   "Test ownership",
		"system_prompt": "System prompt",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token1)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	result := ParseResponse(t, resp)
	agentData := result["data"].(map[string]any)
	agentID := agentData["id"].(string)

	// User 2 tries to list conversations for User 1's agent
	email2 := fmt.Sprintf("chat-own2-%s@test.com", uuid.New().String()[:8])
	RegisterUser(t, env, email2, "password123")
	token2 := LoginUser(t, env, email2, "password123")

	resp = DoRequest(t, env, "GET", "/api/v1/agents/"+agentID+"/conversations", nil, token2)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
