//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemory_CRUD(t *testing.T) {
	env := SetupTestEnv(t)

	// Register and login
	email := fmt.Sprintf("memtest-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	// Create an agent with memory enabled
	agentBody := map[string]any{
		"name":          "Memory Test Agent",
		"system_prompt": "You are a helpful agent.",
		"memory_config": map[string]any{
			"enabled":            true,
			"short_term_enabled": true,
			"long_term_enabled":  true,
		},
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	agentResult := ParseResponse(t, resp)
	agentData := agentResult["data"].(map[string]any)
	agentID := agentData["id"].(string)

	// List memories (should be empty)
	resp = DoRequest(t, env, "GET", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	listResult := ParseResponse(t, resp)
	assert.Equal(t, float64(0), listResult["total_count"])

	// Create a memory
	memBody := map[string]any{
		"content":     "User prefers dark mode",
		"memory_type": "preference",
	}
	resp = DoRequest(t, env, "POST", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), memBody, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	memResult := ParseResponse(t, resp)
	memData := memResult["data"].(map[string]any)
	memoryID := memData["id"].(string)
	assert.Equal(t, "User prefers dark mode", memData["content"])
	assert.Equal(t, "preference", memData["memory_type"])

	// List memories (should have 1)
	resp = DoRequest(t, env, "GET", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	listResult = ParseResponse(t, resp)
	assert.Equal(t, float64(1), listResult["total_count"])

	// Delete single memory
	resp = DoRequest(t, env, "DELETE", fmt.Sprintf("/api/v1/agents/%s/memories/%s", agentID, memoryID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deleted
	resp = DoRequest(t, env, "GET", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	listResult = ParseResponse(t, resp)
	assert.Equal(t, float64(0), listResult["total_count"])
}

func TestMemory_DeleteAll(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("memdelall-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	agentBody := map[string]any{
		"name":          "Delete All Agent",
		"system_prompt": "You are helpful.",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	agentData := ParseResponse(t, resp)["data"].(map[string]any)
	agentID := agentData["id"].(string)

	// Create multiple memories
	for i := 0; i < 3; i++ {
		memBody := map[string]any{
			"content":     fmt.Sprintf("Memory %d", i),
			"memory_type": "fact",
		}
		resp = DoRequest(t, env, "POST", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), memBody, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Verify 3 memories
	resp = DoRequest(t, env, "GET", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, float64(3), ParseResponse(t, resp)["total_count"])

	// Delete all
	resp = DoRequest(t, env, "DELETE", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify empty
	resp = DoRequest(t, env, "GET", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, float64(0), ParseResponse(t, resp)["total_count"])
}

func TestMemory_OwnershipIsolation(t *testing.T) {
	env := SetupTestEnv(t)

	// User 1
	email1 := fmt.Sprintf("memowner1-%d@test.com", uniqueID())
	RegisterUser(t, env, email1, "password123")
	token1 := LoginUser(t, env, email1, "password123")

	// User 2
	email2 := fmt.Sprintf("memowner2-%d@test.com", uniqueID())
	RegisterUser(t, env, email2, "password123")
	token2 := LoginUser(t, env, email2, "password123")

	// User 1 creates agent
	agentBody := map[string]any{
		"name":          "User1 Agent",
		"system_prompt": "Hello",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token1)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	agentData := ParseResponse(t, resp)["data"].(map[string]any)
	agentID := agentData["id"].(string)

	// User 1 creates memory
	memBody := map[string]any{
		"content":     "Secret memory",
		"memory_type": "fact",
	}
	resp = DoRequest(t, env, "POST", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), memBody, token1)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// User 2 cannot access User 1's agent memories
	resp = DoRequest(t, env, "GET", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), nil, token2)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	ParseResponse(t, resp) // drain body
}

func TestMemory_SearchWithEmbedding(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("memsearch-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	agentBody := map[string]any{
		"name":          "Search Agent",
		"system_prompt": "Search test",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	agentData := ParseResponse(t, resp)["data"].(map[string]any)
	agentID := agentData["id"].(string)

	// Create memories with embeddings (384-dim)
	emb1 := make([]float64, 384)
	emb1[0] = 1.0
	emb2 := make([]float64, 384)
	emb2[0] = 0.9
	emb2[1] = 0.1
	emb3 := make([]float64, 384)
	emb3[383] = 1.0

	memBodies := []map[string]any{
		{"content": "close match", "memory_type": "fact", "embedding": emb1},
		{"content": "nearby match", "memory_type": "fact", "embedding": emb2},
		{"content": "far away", "memory_type": "fact", "embedding": emb3},
	}

	for _, mb := range memBodies {
		resp = DoRequest(t, env, "POST", fmt.Sprintf("/api/v1/agents/%s/memories", agentID), mb, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Search with embedding close to emb1
	searchBody := map[string]any{
		"embedding": emb1,
		"limit":     2,
		"threshold": 0.5,
	}
	resp = DoRequest(t, env, "POST", fmt.Sprintf("/api/v1/agents/%s/memories/search", agentID), searchBody, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	searchResult := ParseResponse(t, resp)
	results := searchResult["data"].([]any)
	assert.LessOrEqual(t, len(results), 2)
	if len(results) > 0 {
		first := results[0].(map[string]any)
		mem := first["memory"].(map[string]any)
		assert.Equal(t, "close match", mem["content"])
	}
}

var _uniqueCounter int64

func uniqueID() int64 {
	_uniqueCounter++
	return _uniqueCounter
}
