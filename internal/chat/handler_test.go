package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiox-platform/aiox/internal/agents"
	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
)

func TestSendMessage_EmptyBody(t *testing.T) {
	h := &Handler{}

	userID := uuid.New().String()
	agent := &agents.Agent{
		ID:          uuid.New(),
		OwnerUserID: uuid.MustParse(userID),
		JID:         "agent-test@agents.local",
	}

	body := `{"body":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/"+agent.ID.String()+"/messages", bytes.NewBufferString(body))
	ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &auth.AccessClaims{UserID: userID, Email: "test@test.com"})
	ctx = agents.SetAgentInContext(ctx, agent)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.SendMessage(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp api.Response
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Contains(t, resp.Error, "body is required")
}

func TestSendMessage_NoAuth(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/123/messages", nil)
	w := httptest.NewRecorder()
	h.SendMessage(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSendMessage_NoAgent(t *testing.T) {
	h := &Handler{}

	userID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/123/messages", bytes.NewBufferString(`{"body":"hi"}`))
	ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &auth.AccessClaims{UserID: userID, Email: "test@test.com"})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.SendMessage(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListConversations_InvalidAgentID(t *testing.T) {
	h := &Handler{}

	userID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/not-uuid/conversations", nil)

	// Setup chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agentID", "not-uuid")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, auth.UserClaimsKey, &auth.AccessClaims{UserID: userID, Email: "test@test.com"})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.ListConversations(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListConversations_NoAuth(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/123/conversations", nil)
	w := httptest.NewRecorder()
	h.ListConversations(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
