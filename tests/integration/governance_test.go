//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiox-platform/aiox/internal/governance/audit"
)

func TestGovernance_Quota_API(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("govquota-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	// GET quota — should return defaults with zero usage
	resp := DoRequest(t, env, "GET", "/api/v1/governance/quota", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result := ParseResponse(t, resp)
	data := result["data"].(map[string]any)

	assert.Equal(t, float64(0), data["tokens_used_today"])
	assert.Equal(t, float64(100000), data["tokens_limit_day"])
	assert.Equal(t, float64(0), data["requests_today"])
	assert.Equal(t, float64(1000), data["requests_limit_day"])
	assert.Equal(t, float64(0), data["tokens_used_minute"])
	assert.Equal(t, float64(10000), data["tokens_limit_minute"])
}

func TestGovernance_Quota_Unauthenticated(t *testing.T) {
	env := SetupTestEnv(t)

	resp := DoRequest(t, env, "GET", "/api/v1/governance/quota", nil, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGovernance_AuditLogs_Empty(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("govaudit-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	// GET audit logs — should be empty for new user
	resp := DoRequest(t, env, "GET", "/api/v1/governance/audit", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result := ParseResponse(t, resp)
	assert.Equal(t, float64(0), result["total_count"])
}

func TestGovernance_AuditLogs_WithPersistedEntries(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("govauditpersist-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	// Get user ID from login
	loginResp := DoRequest(t, env, "POST", "/api/v1/auth/login", map[string]string{"email": email, "password": "password123"}, "")
	loginResult := ParseResponse(t, loginResp)
	loginData := loginResult["data"].(map[string]any)
	token = loginData["access_token"].(string)

	// We need to find the user_id — let's get it from a token claim by creating an agent
	agentBody := map[string]any{
		"name":          "Audit Test Agent",
		"system_prompt": "Test agent.",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	agentResult := ParseResponse(t, resp)
	agentData := agentResult["data"].(map[string]any)
	agentID := agentData["id"].(string)
	ownerUserID := agentData["owner_user_id"].(string)

	// Insert audit log directly via repository
	auditRepo := audit.NewRepository(env.Pool)
	parsedOwnerID, _ := uuid.Parse(ownerUserID)
	parsedAgentID, _ := uuid.Parse(agentID)

	log1 := &audit.AuditLog{
		OwnerUserID:  parsedOwnerID,
		EventType:    "message_routed",
		Severity:     "info",
		ResourceType: "agent",
		ResourceID:   &parsedAgentID,
	}
	require.NoError(t, auditRepo.Insert(context.Background(), log1))

	log2 := &audit.AuditLog{
		OwnerUserID:  parsedOwnerID,
		EventType:    "task_completed",
		Severity:     "info",
		ResourceType: "agent",
		ResourceID:   &parsedAgentID,
	}
	require.NoError(t, auditRepo.Insert(context.Background(), log2))

	log3 := &audit.AuditLog{
		OwnerUserID:  parsedOwnerID,
		EventType:    "task_failed",
		Severity:     "warn",
		ResourceType: "agent",
		ResourceID:   &parsedAgentID,
	}
	require.NoError(t, auditRepo.Insert(context.Background(), log3))

	// List all audit logs
	resp = DoRequest(t, env, "GET", "/api/v1/governance/audit", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result := ParseResponse(t, resp)
	assert.Equal(t, float64(3), result["total_count"])

	// Filter by event_type
	resp = DoRequest(t, env, "GET", "/api/v1/governance/audit?event_type=task_failed", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result = ParseResponse(t, resp)
	assert.Equal(t, float64(1), result["total_count"])

	// Filter by severity
	resp = DoRequest(t, env, "GET", "/api/v1/governance/audit?severity=warn", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result = ParseResponse(t, resp)
	assert.Equal(t, float64(1), result["total_count"])

	// Agent-specific audit logs
	resp = DoRequest(t, env, "GET", fmt.Sprintf("/api/v1/agents/%s/audit", agentID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result = ParseResponse(t, resp)
	assert.Equal(t, float64(3), result["total_count"])
}

func TestGovernance_AuditLogs_OwnershipIsolation(t *testing.T) {
	env := SetupTestEnv(t)

	// User 1
	email1 := fmt.Sprintf("goviso1-%d@test.com", uniqueID())
	RegisterUser(t, env, email1, "password123")
	token1 := LoginUser(t, env, email1, "password123")

	// User 2
	email2 := fmt.Sprintf("goviso2-%d@test.com", uniqueID())
	RegisterUser(t, env, email2, "password123")
	token2 := LoginUser(t, env, email2, "password123")

	// Create agent for user 1
	agentBody := map[string]any{
		"name":          "User1 Agent",
		"system_prompt": "Test.",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token1)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	agentResult := ParseResponse(t, resp)
	agentData := agentResult["data"].(map[string]any)
	agentID := agentData["id"].(string)
	ownerUserID := agentData["owner_user_id"].(string)

	// Insert audit log for user 1
	auditRepo := audit.NewRepository(env.Pool)
	parsedOwnerID, _ := uuid.Parse(ownerUserID)
	parsedAgentID, _ := uuid.Parse(agentID)

	log := &audit.AuditLog{
		OwnerUserID:  parsedOwnerID,
		EventType:    "message_routed",
		Severity:     "info",
		ResourceType: "agent",
		ResourceID:   &parsedAgentID,
	}
	require.NoError(t, auditRepo.Insert(context.Background(), log))

	// User 1 should see their audit logs
	resp = DoRequest(t, env, "GET", "/api/v1/governance/audit", nil, token1)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result := ParseResponse(t, resp)
	assert.GreaterOrEqual(t, result["total_count"].(float64), float64(1))

	// User 2 should NOT see user 1's audit logs
	resp = DoRequest(t, env, "GET", "/api/v1/governance/audit", nil, token2)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result = ParseResponse(t, resp)
	assert.Equal(t, float64(0), result["total_count"])

	// User 2 should NOT be able to access user 1's agent audit logs
	resp = DoRequest(t, env, "GET", fmt.Sprintf("/api/v1/agents/%s/audit", agentID), nil, token2)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGovernance_BlockedAgent(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("govblocked-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	// Create an agent with blocked governance
	agentBody := map[string]any{
		"name":          "Blocked Agent",
		"system_prompt": "I am blocked.",
		"governance": map[string]any{
			"blocked": true,
		},
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	agentResult := ParseResponse(t, resp)
	agentData := agentResult["data"].(map[string]any)

	// Verify governance was stored
	governance := agentData["governance"].(map[string]any)
	assert.Equal(t, true, governance["blocked"])

	// Can still update the agent
	updateBody := map[string]any{
		"governance": map[string]any{
			"blocked": false,
		},
	}
	agentID := agentData["id"].(string)
	resp = DoRequest(t, env, "PUT", fmt.Sprintf("/api/v1/agents/%s", agentID), updateBody, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	updateResult := ParseResponse(t, resp)
	updateData := updateResult["data"].(map[string]any)
	updatedGov := updateData["governance"].(map[string]any)
	assert.Equal(t, false, updatedGov["blocked"])
}
