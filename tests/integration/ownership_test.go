//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOwnershipIsolation(t *testing.T) {
	env := SetupTestEnv(t)

	// Create two users
	RegisterUser(t, env, "owner-a@example.com", "password123")
	RegisterUser(t, env, "owner-b@example.com", "password123")

	tokenA := LoginUser(t, env, "owner-a@example.com", "password123")
	tokenB := LoginUser(t, env, "owner-b@example.com", "password123")

	// User A creates an agent
	body := map[string]any{
		"name":          "User A Agent",
		"system_prompt": "I belong to user A",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", body, tokenA)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	result := ParseResponse(t, resp)
	data := result["data"].(map[string]any)
	agentAID := data["id"].(string)

	t.Run("owner can access own agent", func(t *testing.T) {
		resp := DoRequest(t, env, "GET", "/api/v1/agents/"+agentAID, nil, tokenA)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("other user cannot GET agent", func(t *testing.T) {
		resp := DoRequest(t, env, "GET", "/api/v1/agents/"+agentAID, nil, tokenB)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("other user cannot UPDATE agent", func(t *testing.T) {
		updateBody := map[string]any{"name": "Hacked Name"}
		resp := DoRequest(t, env, "PUT", "/api/v1/agents/"+agentAID, updateBody, tokenB)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("other user cannot DELETE agent", func(t *testing.T) {
		resp := DoRequest(t, env, "DELETE", "/api/v1/agents/"+agentAID, nil, tokenB)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("listing only returns own agents", func(t *testing.T) {
		// User B creates their own agent
		bodyB := map[string]any{
			"name":          "User B Agent",
			"system_prompt": "I belong to user B",
		}
		DoRequest(t, env, "POST", "/api/v1/agents", bodyB, tokenB)

		// User A's list should not contain User B's agents
		listResp := DoRequest(t, env, "GET", "/api/v1/agents", nil, tokenA)
		require.Equal(t, http.StatusOK, listResp.StatusCode)

		listResult := ParseResponse(t, listResp)
		agents := listResult["data"].([]any)
		for _, a := range agents {
			agent := a.(map[string]any)
			profile := agent["profile"].(map[string]any)
			assert.NotEqual(t, "User B Agent", profile["name"],
				"User A should not see User B's agents")
		}
	})

	t.Run("unauthenticated access denied", func(t *testing.T) {
		resp := DoRequest(t, env, "GET", "/api/v1/agents/"+agentAID, nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid token access denied", func(t *testing.T) {
		resp := DoRequest(t, env, "GET", "/api/v1/agents/"+agentAID, nil, "invalid-jwt-token")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
