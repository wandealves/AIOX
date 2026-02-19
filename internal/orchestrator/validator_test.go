package orchestrator

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Validate(t *testing.T) {
	v := NewValidator()

	t.Run("valid route passes", func(t *testing.T) {
		route := &RouteResult{
			AgentID:     uuid.New(),
			OwnerUserID: uuid.New(),
			AgentJID:    "agent-123@agents.aiox.local",
			Visibility:  "private",
		}
		assert.NoError(t, v.Validate(route))
	})

	t.Run("nil agent ID fails", func(t *testing.T) {
		route := &RouteResult{
			AgentID:     uuid.Nil,
			OwnerUserID: uuid.New(),
		}
		assert.Error(t, v.Validate(route))
	})

	t.Run("nil owner ID fails", func(t *testing.T) {
		route := &RouteResult{
			AgentID:     uuid.New(),
			OwnerUserID: uuid.Nil,
		}
		assert.Error(t, v.Validate(route))
	})

	t.Run("empty governance passes", func(t *testing.T) {
		route := &RouteResult{
			AgentID:     uuid.New(),
			OwnerUserID: uuid.New(),
			AgentJID:    "agent-123@agents.aiox.local",
			Governance:  nil,
		}
		assert.NoError(t, v.Validate(route))
	})

	t.Run("null governance passes", func(t *testing.T) {
		route := &RouteResult{
			AgentID:     uuid.New(),
			OwnerUserID: uuid.New(),
			AgentJID:    "agent-123@agents.aiox.local",
			Governance:  []byte("null"),
		}
		assert.NoError(t, v.Validate(route))
	})

	t.Run("allowed domain passes", func(t *testing.T) {
		gov, _ := json.Marshal(governanceConfig{AllowedDomains: []string{"agents.aiox.local"}})
		route := &RouteResult{
			AgentID:     uuid.New(),
			OwnerUserID: uuid.New(),
			AgentJID:    "agent-123@agents.aiox.local",
			Governance:  gov,
		}
		assert.NoError(t, v.Validate(route))
	})

	t.Run("disallowed domain fails", func(t *testing.T) {
		gov, _ := json.Marshal(governanceConfig{AllowedDomains: []string{"other.domain.com"}})
		route := &RouteResult{
			AgentID:     uuid.New(),
			OwnerUserID: uuid.New(),
			AgentJID:    "agent-123@agents.aiox.local",
			Governance:  gov,
		}
		assert.Error(t, v.Validate(route))
	})

	t.Run("domain check is case insensitive", func(t *testing.T) {
		gov, _ := json.Marshal(governanceConfig{AllowedDomains: []string{"AGENTS.AIOX.LOCAL"}})
		route := &RouteResult{
			AgentID:     uuid.New(),
			OwnerUserID: uuid.New(),
			AgentJID:    "agent-123@agents.aiox.local",
			Governance:  gov,
		}
		assert.NoError(t, v.Validate(route))
	})
}

func TestValidator_ValidateOwnership(t *testing.T) {
	v := NewValidator()

	t.Run("matching IDs pass", func(t *testing.T) {
		id := uuid.New()
		require.NoError(t, v.ValidateOwnership(id, id))
	})

	t.Run("different IDs fail", func(t *testing.T) {
		assert.Error(t, v.ValidateOwnership(uuid.New(), uuid.New()))
	})
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		jid    string
		domain string
	}{
		{"user@domain.com", "domain.com"},
		{"user@domain.com/resource", "domain.com"},
		{"agent-123@agents.aiox.local", "agents.aiox.local"},
		{"nodomain", "nodomain"},
	}

	for _, tt := range tests {
		t.Run(tt.jid, func(t *testing.T) {
			assert.Equal(t, tt.domain, extractDomain(tt.jid))
		})
	}
}
