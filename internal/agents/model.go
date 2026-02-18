package agents

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID           uuid.UUID        `json:"id"`
	OwnerUserID  uuid.UUID        `json:"owner_user_id"`
	JID          string           `json:"jid"`
	Profile      AgentProfile     `json:"profile"`
	LLMConfig    json.RawMessage  `json:"llm_config"`
	Capabilities json.RawMessage  `json:"capabilities"`
	MemoryConfig json.RawMessage  `json:"memory_config"`
	Governance   json.RawMessage  `json:"governance"`
	Visibility   string           `json:"visibility"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	DeletedAt    *time.Time       `json:"deleted_at,omitempty"`
}

type AgentProfile struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	SystemPrompt      string   `json:"system_prompt"`
	PersonalityTraits []string `json:"personality_traits,omitempty"`
	Encrypted         bool     `json:"encrypted"`
}

// AgentRow is the database representation with JSONB fields as raw bytes.
type AgentRow struct {
	ID           uuid.UUID
	OwnerUserID  uuid.UUID
	JID          string
	Profile      []byte
	LLMConfig    []byte
	Capabilities []byte
	MemoryConfig []byte
	Governance   []byte
	Visibility   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type CreateAgentRequest struct {
	Name              string          `json:"name" validate:"required,min=1,max=255"`
	Description       string          `json:"description" validate:"max=1000"`
	SystemPrompt      string          `json:"system_prompt" validate:"required,min=1"`
	PersonalityTraits []string        `json:"personality_traits"`
	LLMConfig         json.RawMessage `json:"llm_config"`
	Capabilities      json.RawMessage `json:"capabilities"`
	MemoryConfig      json.RawMessage `json:"memory_config"`
	Governance        json.RawMessage `json:"governance"`
	Visibility        string          `json:"visibility" validate:"omitempty,oneof=private public"`
}

type UpdateAgentRequest struct {
	Name              *string          `json:"name" validate:"omitempty,min=1,max=255"`
	Description       *string          `json:"description" validate:"omitempty,max=1000"`
	SystemPrompt      *string          `json:"system_prompt" validate:"omitempty,min=1"`
	PersonalityTraits *[]string        `json:"personality_traits"`
	LLMConfig         *json.RawMessage `json:"llm_config"`
	Capabilities      *json.RawMessage `json:"capabilities"`
	MemoryConfig      *json.RawMessage `json:"memory_config"`
	Governance        *json.RawMessage `json:"governance"`
	Visibility        *string          `json:"visibility" validate:"omitempty,oneof=private public"`
}

type ListAgentsParams struct {
	Page     int
	PageSize int
}

func DefaultListParams() ListAgentsParams {
	return ListAgentsParams{
		Page:     1,
		PageSize: 20,
	}
}
