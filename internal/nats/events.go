package nats

import (
	"time"

	"github.com/google/uuid"
)

// FetchTimeout is the default timeout for batch fetching messages from consumers.
const FetchTimeout = 2 * time.Second

// Stream names.
const (
	StreamMessages = "AIOX_MESSAGES"
	StreamTasks    = "AIOX_TASKS"
	StreamEvents   = "AIOX_EVENTS"
)

// Subject constants.
const (
	SubjectInboundMessage  = "aiox.messages.inbound"
	SubjectOutboundMessage = "aiox.messages.outbound"
	SubjectTaskPrefix      = "aiox.tasks"     // aiox.tasks.{agent_id}
	SubjectAgentEvent      = "aiox.events.agent"
	SubjectAuditEvent      = "aiox.events.audit"
)

// InboundMessage is published when an XMPP message arrives at the component.
type InboundMessage struct {
	ID         string    `json:"id"`
	FromJID    string    `json:"from_jid"`
	ToJID      string    `json:"to_jid"`
	Body       string    `json:"body"`
	StanzaType string    `json:"stanza_type"`
	ReceivedAt time.Time `json:"received_at"`
}

// OutboundMessage is published to send a message back via XMPP.
type OutboundMessage struct {
	ID        string `json:"id"`
	ToJID     string `json:"to_jid"`
	FromJID   string `json:"from_jid"`
	Body      string `json:"body"`
	InReplyTo string `json:"in_reply_to,omitempty"`
}

// TaskMessage is published for agent task processing via Python workers.
type TaskMessage struct {
	RequestID   string    `json:"request_id"`
	AgentID     uuid.UUID `json:"agent_id"`
	OwnerUserID uuid.UUID `json:"owner_user_id"`
	Message     string    `json:"message"`
	FromJID     string    `json:"from_jid"`
	AgentJID    string    `json:"agent_jid"`
	AgentName   string    `json:"agent_name"`
}

// AgentEvent is published for agent lifecycle events.
type AgentEvent struct {
	AgentID     uuid.UUID `json:"agent_id"`
	OwnerUserID uuid.UUID `json:"owner_user_id"`
	JID         string    `json:"jid"`
	EventType   string    `json:"event_type"` // e.g., "message_received", "message_sent"
	Timestamp   time.Time `json:"timestamp"`
}

// AuditEvent is published for compliance/audit logging.
type AuditEvent struct {
	OwnerUserID  uuid.UUID `json:"owner_user_id"`
	EventType    string    `json:"event_type"`
	Severity     string    `json:"severity"` // info, warn, error
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Details      string    `json:"details"`
	Timestamp    time.Time `json:"timestamp"`
}
