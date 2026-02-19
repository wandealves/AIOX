package memory

import "time"

// ConversationEntry is a single message in the short-term conversation history.
type ConversationEntry struct {
	Role      string    `json:"role"`    // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ContextPayload is the memory context injected into TaskRequest for the Python worker.
type ContextPayload struct {
	RecentMessages   []ConversationEntry `json:"recent_messages"`
	RelevantMemories []RelevantMemory    `json:"relevant_memories"`
}

// RelevantMemory is a long-term memory returned from pgvector similarity search.
type RelevantMemory struct {
	Content    string  `json:"content"`
	MemoryType string  `json:"memory_type"`
	Similarity float64 `json:"similarity"`
}
