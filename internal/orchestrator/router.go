package orchestrator

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/agents"
	ixmpp "github.com/aiox-platform/aiox/internal/xmpp"
)

// RouteResult contains the resolved agent information for a message.
type RouteResult struct {
	AgentID     uuid.UUID
	OwnerUserID uuid.UUID
	AgentName   string
	AgentJID    string
	Visibility  string
	Governance  []byte
}

// Router resolves JIDs to agents using the agents repository.
type Router struct {
	agentRepo agents.Repository
}

// NewRouter creates a new Router.
func NewRouter(agentRepo agents.Repository) *Router {
	return &Router{agentRepo: agentRepo}
}

// Route resolves a message's ToJID to the target agent.
func (r *Router) Route(ctx context.Context, toJID string) (*RouteResult, error) {
	agentID, err := ixmpp.ExtractAgentID(toJID)
	if err != nil {
		return nil, fmt.Errorf("extracting agent ID: %w", err)
	}

	row, err := r.agentRepo.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("looking up agent: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	// Parse profile to get agent name
	name := "unknown"
	profile, err := agents.ParseProfile(row.Profile)
	if err == nil {
		name = profile.Name
	}

	return &RouteResult{
		AgentID:     row.ID,
		OwnerUserID: row.OwnerUserID,
		AgentName:   name,
		AgentJID:    row.JID,
		Visibility:  row.Visibility,
		Governance:  row.Governance,
	}, nil
}
