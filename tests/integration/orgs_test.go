//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrgs_CreateAndList(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("org-owner-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	// Create an org
	resp := DoRequest(t, env, "POST", "/api/v1/orgs", map[string]string{
		"name": "Test Org",
		"slug": fmt.Sprintf("test-org-%d", uniqueID()),
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	result := ParseResponse(t, resp)
	data := result["data"].(map[string]any)
	orgID := data["id"].(string)
	assert.Equal(t, "Test Org", data["name"])

	// List orgs
	resp = DoRequest(t, env, "GET", "/api/v1/orgs", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result = ParseResponse(t, resp)
	orgList := result["data"].([]any)
	assert.GreaterOrEqual(t, len(orgList), 1)

	// Get org by ID
	resp = DoRequest(t, env, "GET", "/api/v1/orgs/"+orgID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result = ParseResponse(t, resp)
	data = result["data"].(map[string]any)
	assert.Equal(t, "Test Org", data["name"])
}

func TestOrgs_UpdateAndDelete(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("org-admin-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	// Create org
	slug := fmt.Sprintf("update-org-%d", uniqueID())
	resp := DoRequest(t, env, "POST", "/api/v1/orgs", map[string]string{
		"name": "Original",
		"slug": slug,
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	result := ParseResponse(t, resp)
	orgID := result["data"].(map[string]any)["id"].(string)

	// Update
	resp = DoRequest(t, env, "PUT", "/api/v1/orgs/"+orgID, map[string]string{
		"name": "Updated Name",
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result = ParseResponse(t, resp)
	assert.Equal(t, "Updated Name", result["data"].(map[string]any)["name"])

	// Delete
	resp = DoRequest(t, env, "DELETE", "/api/v1/orgs/"+orgID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deleted
	resp = DoRequest(t, env, "GET", "/api/v1/orgs/"+orgID, nil, token)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestOrgs_MemberManagement(t *testing.T) {
	env := SetupTestEnv(t)

	// Owner
	ownerEmail := fmt.Sprintf("owner-%d@test.com", uniqueID())
	RegisterUser(t, env, ownerEmail, "password123")
	ownerToken := LoginUser(t, env, ownerEmail, "password123")

	// Create org
	resp := DoRequest(t, env, "POST", "/api/v1/orgs", map[string]string{
		"name": "Member Test Org",
		"slug": fmt.Sprintf("member-org-%d", uniqueID()),
	}, ownerToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	orgID := ParseResponse(t, resp)["data"].(map[string]any)["id"].(string)

	// List members (should have owner)
	resp = DoRequest(t, env, "GET", "/api/v1/orgs/"+orgID+"/members", nil, ownerToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	result := ParseResponse(t, resp)
	members := result["data"].([]any)
	assert.Equal(t, 1, len(members))
	assert.Equal(t, "owner", members[0].(map[string]any)["role"])
}

func TestOrgs_InviteFlow(t *testing.T) {
	env := SetupTestEnv(t)

	// Owner
	ownerEmail := fmt.Sprintf("inv-owner-%d@test.com", uniqueID())
	RegisterUser(t, env, ownerEmail, "password123")
	ownerToken := LoginUser(t, env, ownerEmail, "password123")

	// Invitee
	inviteeEmail := fmt.Sprintf("inv-user-%d@test.com", uniqueID())
	RegisterUser(t, env, inviteeEmail, "password123")
	inviteeToken := LoginUser(t, env, inviteeEmail, "password123")

	// Create org
	resp := DoRequest(t, env, "POST", "/api/v1/orgs", map[string]string{
		"name": "Invite Test Org",
		"slug": fmt.Sprintf("invite-org-%d", uniqueID()),
	}, ownerToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	orgID := ParseResponse(t, resp)["data"].(map[string]any)["id"].(string)

	// Create invite
	resp = DoRequest(t, env, "POST", "/api/v1/orgs/"+orgID+"/invites", map[string]string{
		"email": inviteeEmail,
		"role":  "editor",
	}, ownerToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	inviteData := ParseResponse(t, resp)["data"].(map[string]any)
	token := inviteData["token"].(string)
	assert.NotEmpty(t, token)

	// List invites
	resp = DoRequest(t, env, "GET", "/api/v1/orgs/"+orgID+"/invites", nil, ownerToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	invites := ParseResponse(t, resp)["data"].([]any)
	assert.Equal(t, 1, len(invites))

	// Accept invite
	resp = DoRequest(t, env, "POST", "/api/v1/invites/"+token+"/accept", nil, inviteeToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify invitee is now a member
	resp = DoRequest(t, env, "GET", "/api/v1/orgs/"+orgID+"/members", nil, inviteeToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	members := ParseResponse(t, resp)["data"].([]any)
	assert.Equal(t, 2, len(members))
}

func TestOrgs_RoleBasedAccess(t *testing.T) {
	env := SetupTestEnv(t)

	// Owner
	ownerEmail := fmt.Sprintf("rbac-owner-%d@test.com", uniqueID())
	RegisterUser(t, env, ownerEmail, "password123")
	ownerToken := LoginUser(t, env, ownerEmail, "password123")

	// Viewer
	viewerEmail := fmt.Sprintf("rbac-viewer-%d@test.com", uniqueID())
	RegisterUser(t, env, viewerEmail, "password123")
	viewerToken := LoginUser(t, env, viewerEmail, "password123")

	// Create org + invite viewer
	resp := DoRequest(t, env, "POST", "/api/v1/orgs", map[string]string{
		"name": "RBAC Test",
		"slug": fmt.Sprintf("rbac-org-%d", uniqueID()),
	}, ownerToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	orgID := ParseResponse(t, resp)["data"].(map[string]any)["id"].(string)

	resp = DoRequest(t, env, "POST", "/api/v1/orgs/"+orgID+"/invites", map[string]string{
		"email": viewerEmail,
		"role":  "viewer",
	}, ownerToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	invToken := ParseResponse(t, resp)["data"].(map[string]any)["token"].(string)

	resp = DoRequest(t, env, "POST", "/api/v1/invites/"+invToken+"/accept", nil, viewerToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Viewer CAN read org
	resp = DoRequest(t, env, "GET", "/api/v1/orgs/"+orgID, nil, viewerToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Viewer CANNOT update org (requires admin+)
	resp = DoRequest(t, env, "PUT", "/api/v1/orgs/"+orgID, map[string]string{
		"name": "Hacked",
	}, viewerToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	// Viewer CANNOT delete org (requires owner)
	resp = DoRequest(t, env, "DELETE", "/api/v1/orgs/"+orgID, nil, viewerToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	// Viewer CANNOT create invites (requires admin+)
	resp = DoRequest(t, env, "POST", "/api/v1/orgs/"+orgID+"/invites", map[string]string{
		"email": "nope@test.com",
		"role":  "viewer",
	}, viewerToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestOrgs_OwnershipIsolation(t *testing.T) {
	env := SetupTestEnv(t)

	// User1
	email1 := fmt.Sprintf("iso-user1-%d@test.com", uniqueID())
	RegisterUser(t, env, email1, "password123")
	token1 := LoginUser(t, env, email1, "password123")

	// User2 (not a member)
	email2 := fmt.Sprintf("iso-user2-%d@test.com", uniqueID())
	RegisterUser(t, env, email2, "password123")
	token2 := LoginUser(t, env, email2, "password123")

	// User1 creates an org
	resp := DoRequest(t, env, "POST", "/api/v1/orgs", map[string]string{
		"name": "Private Org",
		"slug": fmt.Sprintf("private-org-%d", uniqueID()),
	}, token1)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	orgID := ParseResponse(t, resp)["data"].(map[string]any)["id"].(string)

	// User2 cannot access the org (not a member)
	resp = DoRequest(t, env, "GET", "/api/v1/orgs/"+orgID, nil, token2)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestOrgs_DuplicateSlug(t *testing.T) {
	env := SetupTestEnv(t)

	email := fmt.Sprintf("slug-user-%d@test.com", uniqueID())
	RegisterUser(t, env, email, "password123")
	token := LoginUser(t, env, email, "password123")

	slug := fmt.Sprintf("dup-slug-%d", uniqueID())

	// First org with slug
	resp := DoRequest(t, env, "POST", "/api/v1/orgs", map[string]string{
		"name": "First",
		"slug": slug,
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Second org with same slug should fail
	resp = DoRequest(t, env, "POST", "/api/v1/orgs", map[string]string{
		"name": "Second",
		"slug": slug,
	}, token)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()
}

var counter int64

func uniqueID() int64 {
	counter++
	return counter
}
