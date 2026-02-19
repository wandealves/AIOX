package xmpp

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractAgentID(t *testing.T) {
	expected := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name    string
		jid     string
		wantID  uuid.UUID
		wantErr bool
	}{
		{
			name:   "valid bare JID",
			jid:    "agent-550e8400-e29b-41d4-a716-446655440000@agents.aiox.local",
			wantID: expected,
		},
		{
			name:   "valid JID with resource",
			jid:    "agent-550e8400-e29b-41d4-a716-446655440000@agents.aiox.local/resource",
			wantID: expected,
		},
		{
			name:    "missing agent- prefix",
			jid:     "550e8400-e29b-41d4-a716-446655440000@agents.aiox.local",
			wantErr: true,
		},
		{
			name:    "invalid UUID",
			jid:     "agent-not-a-uuid@agents.aiox.local",
			wantErr: true,
		},
		{
			name:    "empty JID",
			jid:     "",
			wantErr: true,
		},
		{
			name:    "no @ sign",
			jid:     "agent-550e8400-e29b-41d4-a716-446655440000",
			wantID:  expected,
		},
		{
			name:    "user without agent prefix",
			jid:     "user@aiox.local",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractAgentID(tt.jid)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got)
		})
	}
}
