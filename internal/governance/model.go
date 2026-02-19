package governance

import "encoding/json"

// GovernanceConfig represents the governance JSONB structure on an agent.
type GovernanceConfig struct {
	AllowedDomains      []string `json:"allowed_domains,omitempty"`
	MaxTokensPerRequest int      `json:"max_tokens_per_request,omitempty"`
	AllowedProviders    []string `json:"allowed_providers,omitempty"`
	Blocked             bool     `json:"blocked,omitempty"`
}

// ParseGovernance parses agent governance JSONB into GovernanceConfig.
// Returns zero-value config on nil, empty, or invalid input.
func ParseGovernance(data []byte) GovernanceConfig {
	var cfg GovernanceConfig
	if len(data) == 0 {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	return cfg
}
