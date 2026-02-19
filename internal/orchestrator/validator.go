package orchestrator

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/governance"
)

// Validator checks ownership and governance rules for message routing.
type Validator struct{}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks that the route result is valid for processing.
func (v *Validator) Validate(route *RouteResult) error {
	if route.AgentID == uuid.Nil {
		return fmt.Errorf("agent not found")
	}
	if route.OwnerUserID == uuid.Nil {
		return fmt.Errorf("agent has no owner")
	}
	return v.checkGovernance(route)
}

// ValidateOwnership checks that the requesting user owns the agent.
// This will be used in Phase 3 when XMPP user → platform user mapping exists.
func (v *Validator) ValidateOwnership(fromUserID, ownerUserID uuid.UUID) error {
	if fromUserID != ownerUserID {
		return fmt.Errorf("ownership violation: user %s does not own agent (owner: %s)", fromUserID, ownerUserID)
	}
	return nil
}

func (v *Validator) checkGovernance(route *RouteResult) error {
	if len(route.Governance) == 0 || string(route.Governance) == "null" {
		return nil
	}

	gov := governance.ParseGovernance(route.Governance)

	// Check if agent is blocked
	if gov.Blocked {
		return fmt.Errorf("agent is blocked by governance policy")
	}

	// If allowed_domains is configured, validate the agent's JID domain
	if len(gov.AllowedDomains) > 0 {
		jidDomain := extractDomain(route.AgentJID)
		if !domainAllowed(jidDomain, gov.AllowedDomains) {
			return fmt.Errorf("agent JID domain %q not in allowed domains", jidDomain)
		}
	}

	return nil
}

func extractDomain(jid string) string {
	// Strip resource
	bare := jid
	if idx := strings.Index(jid, "/"); idx >= 0 {
		bare = jid[:idx]
	}
	// Get domain after @
	if idx := strings.Index(bare, "@"); idx >= 0 {
		return bare[idx+1:]
	}
	return bare
}

func domainAllowed(domain string, allowed []string) bool {
	for _, d := range allowed {
		if strings.EqualFold(d, domain) {
			return true
		}
	}
	return false
}
