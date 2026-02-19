package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ShortTermStore manages conversation context in Redis lists.
type ShortTermStore struct {
	client *redis.Client
}

// NewShortTermStore creates a new short-term memory store.
func NewShortTermStore(client *redis.Client) *ShortTermStore {
	return &ShortTermStore{client: client}
}

func convKey(agentID uuid.UUID, userJID string) string {
	return fmt.Sprintf("conv:%s:%s", agentID.String(), userJID)
}

// GetRecentMessages returns the last `limit` conversation entries for the given agent+user pair.
func (s *ShortTermStore) GetRecentMessages(ctx context.Context, agentID uuid.UUID, userJID string, limit int) ([]ConversationEntry, error) {
	key := convKey(agentID, userJID)

	// LRANGE key -limit -1 returns the last `limit` elements
	vals, err := s.client.LRange(ctx, key, int64(-limit), -1).Result()
	if err != nil {
		return nil, fmt.Errorf("lrange %s: %w", key, err)
	}

	entries := make([]ConversationEntry, 0, len(vals))
	for _, v := range vals {
		var entry ConversationEntry
		if err := json.Unmarshal([]byte(v), &entry); err != nil {
			continue // skip malformed entries
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// AppendMessage adds a conversation entry to the Redis list and trims to maxMsgs.
func (s *ShortTermStore) AppendMessage(ctx context.Context, agentID uuid.UUID, userJID string, entry ConversationEntry, maxMsgs int, ttlSec int) error {
	key := convKey(agentID, userJID)

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling entry: %w", err)
	}

	pipe := s.client.Pipeline()
	pipe.RPush(ctx, key, string(data))
	pipe.LTrim(ctx, key, int64(-maxMsgs), -1)
	pipe.Expire(ctx, key, time.Duration(ttlSec)*time.Second)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline exec for %s: %w", key, err)
	}
	return nil
}

// ClearConversation deletes the conversation history for the given agent+user pair.
func (s *ShortTermStore) ClearConversation(ctx context.Context, agentID uuid.UUID, userJID string) error {
	key := convKey(agentID, userJID)
	return s.client.Del(ctx, key).Err()
}
