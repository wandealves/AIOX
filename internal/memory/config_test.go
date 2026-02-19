package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig_Nil(t *testing.T) {
	cfg := ParseConfig(nil)
	assert.False(t, cfg.Enabled)
	assert.True(t, cfg.ShortTermEnabled)
	assert.True(t, cfg.LongTermEnabled)
	assert.Equal(t, 20, cfg.MaxShortTermMsgs)
	assert.Equal(t, 3600, cfg.ShortTermTTLSec)
	assert.Equal(t, 5, cfg.MaxLongTermResults)
	assert.Equal(t, 0.7, cfg.SimilarityThreshold)
}

func TestParseConfig_Empty(t *testing.T) {
	cfg := ParseConfig([]byte{})
	assert.Equal(t, DefaultConfig(), cfg)
}

func TestParseConfig_EmptyObject(t *testing.T) {
	cfg := ParseConfig([]byte(`{}`))
	assert.Equal(t, DefaultConfig(), cfg)
}

func TestParseConfig_InvalidJSON(t *testing.T) {
	cfg := ParseConfig([]byte(`not json`))
	assert.Equal(t, DefaultConfig(), cfg)
}

func TestParseConfig_FullValid(t *testing.T) {
	data := []byte(`{
		"enabled": true,
		"short_term_enabled": true,
		"long_term_enabled": true,
		"max_short_term_msgs": 50,
		"short_term_ttl_sec": 7200,
		"max_long_term_results": 10,
		"similarity_threshold": 0.5
	}`)
	cfg := ParseConfig(data)
	assert.True(t, cfg.Enabled)
	assert.True(t, cfg.ShortTermEnabled)
	assert.True(t, cfg.LongTermEnabled)
	assert.Equal(t, 50, cfg.MaxShortTermMsgs)
	assert.Equal(t, 7200, cfg.ShortTermTTLSec)
	assert.Equal(t, 10, cfg.MaxLongTermResults)
	assert.Equal(t, 0.5, cfg.SimilarityThreshold)
}

func TestParseConfig_Partial(t *testing.T) {
	data := []byte(`{"enabled": true, "max_short_term_msgs": 30}`)
	cfg := ParseConfig(data)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 30, cfg.MaxShortTermMsgs)
	// Defaults for unspecified fields
	assert.True(t, cfg.ShortTermEnabled)
	assert.True(t, cfg.LongTermEnabled)
	assert.Equal(t, 3600, cfg.ShortTermTTLSec)
	assert.Equal(t, 5, cfg.MaxLongTermResults)
	assert.Equal(t, 0.7, cfg.SimilarityThreshold)
}

func TestParseConfig_DisabledExplicitly(t *testing.T) {
	data := []byte(`{"enabled": false, "short_term_enabled": false, "long_term_enabled": false}`)
	cfg := ParseConfig(data)
	assert.False(t, cfg.Enabled)
	assert.False(t, cfg.ShortTermEnabled)
	assert.False(t, cfg.LongTermEnabled)
}
