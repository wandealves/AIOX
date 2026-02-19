package memory

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMiniredis(t *testing.T) (*ShortTermStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })
	return NewShortTermStore(client), mr
}

func TestShortTermStore_AppendAndGet(t *testing.T) {
	store, _ := setupMiniredis(t)
	ctx := context.Background()
	agentID := uuid.New()
	userJID := "user@example.com"

	// Append two messages
	err := store.AppendMessage(ctx, agentID, userJID, ConversationEntry{
		Role:      "user",
		Content:   "Hello",
		Timestamp: time.Now(),
	}, 20, 3600)
	require.NoError(t, err)

	err = store.AppendMessage(ctx, agentID, userJID, ConversationEntry{
		Role:      "assistant",
		Content:   "Hi there!",
		Timestamp: time.Now(),
	}, 20, 3600)
	require.NoError(t, err)

	// Retrieve
	msgs, err := store.GetRecentMessages(ctx, agentID, userJID, 10)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)
	assert.Equal(t, "user", msgs[0].Role)
	assert.Equal(t, "Hello", msgs[0].Content)
	assert.Equal(t, "assistant", msgs[1].Role)
	assert.Equal(t, "Hi there!", msgs[1].Content)
}

func TestShortTermStore_Trim(t *testing.T) {
	store, _ := setupMiniredis(t)
	ctx := context.Background()
	agentID := uuid.New()
	userJID := "user@example.com"

	// Append 5 messages with max 3
	for i := 0; i < 5; i++ {
		err := store.AppendMessage(ctx, agentID, userJID, ConversationEntry{
			Role:      "user",
			Content:   string(rune('A' + i)),
			Timestamp: time.Now(),
		}, 3, 3600)
		require.NoError(t, err)
	}

	// Should only have the last 3
	msgs, err := store.GetRecentMessages(ctx, agentID, userJID, 10)
	require.NoError(t, err)
	assert.Len(t, msgs, 3)
	assert.Equal(t, "C", msgs[0].Content)
	assert.Equal(t, "D", msgs[1].Content)
	assert.Equal(t, "E", msgs[2].Content)
}

func TestShortTermStore_TTL(t *testing.T) {
	store, mr := setupMiniredis(t)
	ctx := context.Background()
	agentID := uuid.New()
	userJID := "user@example.com"

	err := store.AppendMessage(ctx, agentID, userJID, ConversationEntry{
		Role:    "user",
		Content: "Hello",
	}, 20, 60)
	require.NoError(t, err)

	// Fast-forward time past TTL
	mr.FastForward(61 * time.Second)

	msgs, err := store.GetRecentMessages(ctx, agentID, userJID, 10)
	require.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestShortTermStore_Clear(t *testing.T) {
	store, _ := setupMiniredis(t)
	ctx := context.Background()
	agentID := uuid.New()
	userJID := "user@example.com"

	err := store.AppendMessage(ctx, agentID, userJID, ConversationEntry{
		Role:    "user",
		Content: "Hello",
	}, 20, 3600)
	require.NoError(t, err)

	err = store.ClearConversation(ctx, agentID, userJID)
	require.NoError(t, err)

	msgs, err := store.GetRecentMessages(ctx, agentID, userJID, 10)
	require.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestShortTermStore_GetEmptyReturnsEmpty(t *testing.T) {
	store, _ := setupMiniredis(t)
	ctx := context.Background()
	agentID := uuid.New()
	userJID := "user@example.com"

	msgs, err := store.GetRecentMessages(ctx, agentID, userJID, 10)
	require.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestShortTermStore_IsolatedByAgentAndUser(t *testing.T) {
	store, _ := setupMiniredis(t)
	ctx := context.Background()
	agent1 := uuid.New()
	agent2 := uuid.New()
	user1 := "user1@example.com"
	user2 := "user2@example.com"

	// Agent1 + User1
	err := store.AppendMessage(ctx, agent1, user1, ConversationEntry{
		Role: "user", Content: "A1U1",
	}, 20, 3600)
	require.NoError(t, err)

	// Agent1 + User2
	err = store.AppendMessage(ctx, agent1, user2, ConversationEntry{
		Role: "user", Content: "A1U2",
	}, 20, 3600)
	require.NoError(t, err)

	// Agent2 + User1
	err = store.AppendMessage(ctx, agent2, user1, ConversationEntry{
		Role: "user", Content: "A2U1",
	}, 20, 3600)
	require.NoError(t, err)

	msgs, _ := store.GetRecentMessages(ctx, agent1, user1, 10)
	assert.Len(t, msgs, 1)
	assert.Equal(t, "A1U1", msgs[0].Content)

	msgs, _ = store.GetRecentMessages(ctx, agent1, user2, 10)
	assert.Len(t, msgs, 1)
	assert.Equal(t, "A1U2", msgs[0].Content)

	msgs, _ = store.GetRecentMessages(ctx, agent2, user1, 10)
	assert.Len(t, msgs, 1)
	assert.Equal(t, "A2U1", msgs[0].Content)
}
