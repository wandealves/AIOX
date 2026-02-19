package memory

import "encoding/json"

// MemoryConfig holds agent-level memory settings parsed from agents.memory_config JSONB.
type MemoryConfig struct {
	Enabled             bool    `json:"enabled"`
	ShortTermEnabled    bool    `json:"short_term_enabled"`
	LongTermEnabled     bool    `json:"long_term_enabled"`
	MaxShortTermMsgs    int     `json:"max_short_term_msgs"`
	ShortTermTTLSec     int     `json:"short_term_ttl_sec"`
	MaxLongTermResults  int     `json:"max_long_term_results"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
}

// DefaultConfig returns a MemoryConfig with sensible defaults.
func DefaultConfig() MemoryConfig {
	return MemoryConfig{
		Enabled:             false,
		ShortTermEnabled:    true,
		LongTermEnabled:     true,
		MaxShortTermMsgs:    20,
		ShortTermTTLSec:     3600,
		MaxLongTermResults:  5,
		SimilarityThreshold: 0.7,
	}
}

// ParseConfig parses agent memory_config JSONB into MemoryConfig.
// Returns defaults on nil, empty, or invalid input. Partial JSON is merged over defaults.
func ParseConfig(data []byte) MemoryConfig {
	cfg := DefaultConfig()
	if len(data) == 0 {
		return cfg
	}

	// Parse into a map first to detect which fields were explicitly set
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return cfg
	}
	if len(raw) == 0 {
		return cfg
	}

	// Unmarshal over defaults so only provided fields are overwritten
	_ = json.Unmarshal(data, &cfg)
	return cfg
}
