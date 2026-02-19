package audit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inats "github.com/aiox-platform/aiox/internal/nats"
)

func TestAuditEventDeserialization(t *testing.T) {
	ownerID := uuid.New()
	agentID := uuid.New()

	event := inats.AuditEvent{
		OwnerUserID:  ownerID,
		EventType:    "task_completed",
		Severity:     "info",
		ResourceType: "agent",
		ResourceID:   agentID.String(),
		Details:      "Task processed by worker w1",
		Timestamp:    time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded inats.AuditEvent
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, ownerID, decoded.OwnerUserID)
	assert.Equal(t, "task_completed", decoded.EventType)
	assert.Equal(t, "info", decoded.Severity)
	assert.Equal(t, "agent", decoded.ResourceType)
	assert.Equal(t, agentID.String(), decoded.ResourceID)
	assert.Equal(t, "Task processed by worker w1", decoded.Details)
}

func TestAuditEventToLog_ValidResourceID(t *testing.T) {
	agentID := uuid.New()
	event := inats.AuditEvent{
		OwnerUserID:  uuid.New(),
		EventType:    "message_routed",
		Severity:     "info",
		ResourceType: "agent",
		ResourceID:   agentID.String(),
		Details:      "Message routed from user@test.com",
		Timestamp:    time.Now().UTC(),
	}

	log := convertEventToLog(event)

	assert.Equal(t, event.OwnerUserID, log.OwnerUserID)
	assert.Equal(t, "message_routed", log.EventType)
	assert.Equal(t, "info", log.Severity)
	assert.Equal(t, "agent", log.ResourceType)
	require.NotNil(t, log.ResourceID)
	assert.Equal(t, agentID, *log.ResourceID)

	var details map[string]string
	require.NoError(t, json.Unmarshal(log.Details, &details))
	assert.Equal(t, "Message routed from user@test.com", details["message"])
}

func TestAuditEventToLog_InvalidResourceID(t *testing.T) {
	event := inats.AuditEvent{
		OwnerUserID:  uuid.New(),
		EventType:    "custom_event",
		Severity:     "warn",
		ResourceType: "custom",
		ResourceID:   "not-a-uuid",
		Details:      "Some details",
		Timestamp:    time.Now().UTC(),
	}

	log := convertEventToLog(event)
	assert.Nil(t, log.ResourceID)
}

func TestAuditEventToLog_EmptyResourceID(t *testing.T) {
	event := inats.AuditEvent{
		OwnerUserID: uuid.New(),
		EventType:   "system_event",
		Severity:    "info",
		Details:     "System started",
		Timestamp:   time.Now().UTC(),
	}

	log := convertEventToLog(event)
	assert.Nil(t, log.ResourceID)
}

// convertEventToLog mirrors the consumer's conversion logic for testing.
func convertEventToLog(event inats.AuditEvent) *AuditLog {
	log := &AuditLog{
		ID:           uuid.New(),
		OwnerUserID:  event.OwnerUserID,
		EventType:    event.EventType,
		Severity:     event.Severity,
		ResourceType: event.ResourceType,
		CreatedAt:    event.Timestamp,
	}

	if event.ResourceID != "" {
		if parsed, err := uuid.Parse(event.ResourceID); err == nil {
			log.ResourceID = &parsed
		}
	}

	detailsMap := map[string]string{"message": event.Details}
	if data, err := json.Marshal(detailsMap); err == nil {
		log.Details = data
	}

	return log
}
