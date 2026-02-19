package memory

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Memory represents a row in the agent_memories table.
type Memory struct {
	ID          uuid.UUID       `json:"id"`
	OwnerUserID uuid.UUID       `json:"owner_user_id"`
	AgentID     uuid.UUID       `json:"agent_id"`
	Content     string          `json:"content"`
	Embedding   []float32       `json:"embedding,omitempty"`
	MemoryType  string          `json:"memory_type"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
}

// CreateMemoryRequest is used by the API to create a new memory.
type CreateMemoryRequest struct {
	Content    string          `json:"content" validate:"required,min=1"`
	MemoryType string          `json:"memory_type" validate:"required,min=1"`
	Embedding  []float32       `json:"embedding,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

// SearchMemoryRequest is used by the API to search memories by embedding similarity.
type SearchMemoryRequest struct {
	Embedding []float32 `json:"embedding" validate:"required"`
	Limit     int       `json:"limit,omitempty"`
	Threshold float64   `json:"threshold,omitempty"`
}

// SearchResult wraps a Memory with its similarity score.
type SearchResult struct {
	Memory     Memory  `json:"memory"`
	Similarity float64 `json:"similarity"`
}
