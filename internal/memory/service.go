package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Service orchestrates short-term (Redis) and long-term (pgvector) memory operations.
type Service struct {
	repo       Repository
	shortTerm  *ShortTermStore
}

// NewService creates a new memory service.
func NewService(repo Repository, shortTerm *ShortTermStore) *Service {
	return &Service{
		repo:      repo,
		shortTerm: shortTerm,
	}
}

// GetConversationContext builds the memory context payload for a task request.
// It fetches short-term messages from Redis and searches long-term memories from pgvector.
func (s *Service) GetConversationContext(
	ctx context.Context,
	agentID, ownerUserID uuid.UUID,
	userJID string,
	cfg MemoryConfig,
	queryEmbedding []float32,
) (*ContextPayload, error) {
	payload := &ContextPayload{}

	// Short-term: recent conversation messages
	if cfg.ShortTermEnabled && s.shortTerm != nil {
		msgs, err := s.shortTerm.GetRecentMessages(ctx, agentID, userJID, cfg.MaxShortTermMsgs)
		if err != nil {
			slog.Warn("memory: failed to get short-term messages", "error", err, "agent_id", agentID)
		} else {
			payload.RecentMessages = msgs
		}
	}

	// Long-term: semantic similarity search (only if we have a query embedding)
	if cfg.LongTermEnabled && len(queryEmbedding) > 0 {
		results, err := s.repo.SearchSimilar(ctx, agentID, ownerUserID, queryEmbedding, cfg.MaxLongTermResults, cfg.SimilarityThreshold)
		if err != nil {
			slog.Warn("memory: failed to search long-term memories", "error", err, "agent_id", agentID)
		} else {
			for _, r := range results {
				payload.RelevantMemories = append(payload.RelevantMemories, RelevantMemory{
					Content:    r.Memory.Content,
					MemoryType: r.Memory.MemoryType,
					Similarity: r.Similarity,
				})
			}
		}
	}

	return payload, nil
}

// StoreConversationTurn appends user and assistant messages to the short-term Redis store.
func (s *Service) StoreConversationTurn(
	ctx context.Context,
	agentID uuid.UUID,
	userJID string,
	userMsg, assistantResp string,
	cfg MemoryConfig,
) error {
	if !cfg.ShortTermEnabled || s.shortTerm == nil {
		return nil
	}

	now := time.Now()

	// Append user message
	userEntry := ConversationEntry{
		Role:      "user",
		Content:   userMsg,
		Timestamp: now,
	}
	if err := s.shortTerm.AppendMessage(ctx, agentID, userJID, userEntry, cfg.MaxShortTermMsgs, cfg.ShortTermTTLSec); err != nil {
		return fmt.Errorf("appending user message: %w", err)
	}

	// Append assistant response
	assistantEntry := ConversationEntry{
		Role:      "assistant",
		Content:   assistantResp,
		Timestamp: now,
	}
	if err := s.shortTerm.AppendMessage(ctx, agentID, userJID, assistantEntry, cfg.MaxShortTermMsgs, cfg.ShortTermTTLSec); err != nil {
		return fmt.Errorf("appending assistant message: %w", err)
	}

	return nil
}

// StoreLongTermMemory persists a memory with its embedding to pgvector.
func (s *Service) StoreLongTermMemory(ctx context.Context, mem *Memory) error {
	return s.repo.Create(ctx, mem)
}

// List returns paginated memories for an agent.
func (s *Service) List(ctx context.Context, agentID, ownerUserID uuid.UUID, page, pageSize int) ([]Memory, int64, error) {
	memories, err := s.repo.ListByAgent(ctx, agentID, ownerUserID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.repo.CountByAgent(ctx, agentID, ownerUserID)
	if err != nil {
		return nil, 0, err
	}
	return memories, count, nil
}

// Create creates a new memory.
func (s *Service) Create(ctx context.Context, agentID, ownerUserID uuid.UUID, req *CreateMemoryRequest) (*Memory, error) {
	mem := &Memory{
		ID:          uuid.New(),
		OwnerUserID: ownerUserID,
		AgentID:     agentID,
		Content:     req.Content,
		MemoryType:  req.MemoryType,
		Embedding:   req.Embedding,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}
	if len(mem.Metadata) == 0 {
		mem.Metadata = json.RawMessage(`{}`)
	}
	if err := s.repo.Create(ctx, mem); err != nil {
		return nil, err
	}
	return mem, nil
}

// Search performs a similarity search on agent memories.
func (s *Service) Search(ctx context.Context, agentID, ownerUserID uuid.UUID, req *SearchMemoryRequest) ([]SearchResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 5
	}
	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 0.7
	}
	return s.repo.SearchSimilar(ctx, agentID, ownerUserID, req.Embedding, limit, threshold)
}

// Delete deletes a single memory.
func (s *Service) Delete(ctx context.Context, id, ownerUserID uuid.UUID) error {
	return s.repo.Delete(ctx, id, ownerUserID)
}

// DeleteByAgent deletes all memories for an agent.
func (s *Service) DeleteByAgent(ctx context.Context, agentID, ownerUserID uuid.UUID) error {
	return s.repo.DeleteByAgent(ctx, agentID, ownerUserID)
}
