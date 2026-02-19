package governance

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGovernance_Nil(t *testing.T) {
	cfg := ParseGovernance(nil)
	assert.False(t, cfg.Blocked)
	assert.Nil(t, cfg.AllowedDomains)
	assert.Nil(t, cfg.AllowedProviders)
	assert.Equal(t, 0, cfg.MaxTokensPerRequest)
}

func TestParseGovernance_Empty(t *testing.T) {
	cfg := ParseGovernance([]byte{})
	assert.False(t, cfg.Blocked)
}

func TestParseGovernance_InvalidJSON(t *testing.T) {
	cfg := ParseGovernance([]byte("not json"))
	assert.False(t, cfg.Blocked)
}

func TestParseGovernance_Blocked(t *testing.T) {
	data, _ := json.Marshal(GovernanceConfig{Blocked: true})
	cfg := ParseGovernance(data)
	assert.True(t, cfg.Blocked)
}

func TestParseGovernance_AllowedProviders(t *testing.T) {
	data, _ := json.Marshal(GovernanceConfig{AllowedProviders: []string{"openai", "anthropic"}})
	cfg := ParseGovernance(data)
	assert.Equal(t, []string{"openai", "anthropic"}, cfg.AllowedProviders)
	assert.False(t, cfg.Blocked)
}

func TestParseGovernance_MaxTokensPerRequest(t *testing.T) {
	data, _ := json.Marshal(GovernanceConfig{MaxTokensPerRequest: 4096})
	cfg := ParseGovernance(data)
	assert.Equal(t, 4096, cfg.MaxTokensPerRequest)
}

func TestParseGovernance_Partial(t *testing.T) {
	data := []byte(`{"blocked": true, "allowed_domains": ["example.com"]}`)
	cfg := ParseGovernance(data)
	assert.True(t, cfg.Blocked)
	assert.Equal(t, []string{"example.com"}, cfg.AllowedDomains)
	assert.Nil(t, cfg.AllowedProviders)
	assert.Equal(t, 0, cfg.MaxTokensPerRequest)
}

func TestParseGovernance_FullConfig(t *testing.T) {
	data := []byte(`{
		"allowed_domains": ["agents.aiox.local"],
		"max_tokens_per_request": 2048,
		"allowed_providers": ["openai"],
		"blocked": false
	}`)
	cfg := ParseGovernance(data)
	assert.False(t, cfg.Blocked)
	assert.Equal(t, []string{"agents.aiox.local"}, cfg.AllowedDomains)
	assert.Equal(t, 2048, cfg.MaxTokensPerRequest)
	assert.Equal(t, []string{"openai"}, cfg.AllowedProviders)
}
