package quota

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMiniredis(t *testing.T) *redis.Client {
	t.Helper()
	s := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: s.Addr()})
}

func TestRateLimiter_UnderLimit(t *testing.T) {
	rdb := setupMiniredis(t)
	rl := NewRateLimiter(rdb)
	ctx := context.Background()
	userID := uuid.New()

	allowed, err := rl.CheckAndIncrement(ctx, userID, 10)
	require.NoError(t, err)
	assert.True(t, allowed)

	usage, err := rl.GetMinuteUsage(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 1, usage)
}

func TestRateLimiter_AtLimit(t *testing.T) {
	rdb := setupMiniredis(t)
	rl := NewRateLimiter(rdb)
	ctx := context.Background()
	userID := uuid.New()

	// Fill up to the limit
	for i := 0; i < 5; i++ {
		allowed, err := rl.CheckAndIncrement(ctx, userID, 5)
		require.NoError(t, err)
		assert.True(t, allowed, "request %d should be allowed", i+1)
	}

	// Next should be denied
	allowed, err := rl.CheckAndIncrement(ctx, userID, 5)
	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestRateLimiter_DifferentUsers(t *testing.T) {
	rdb := setupMiniredis(t)
	rl := NewRateLimiter(rdb)
	ctx := context.Background()

	user1 := uuid.New()
	user2 := uuid.New()

	// Fill user1 to the limit
	for i := 0; i < 3; i++ {
		allowed, err := rl.CheckAndIncrement(ctx, user1, 3)
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// user1 should be denied
	allowed, err := rl.CheckAndIncrement(ctx, user1, 3)
	require.NoError(t, err)
	assert.False(t, allowed)

	// user2 should still be allowed
	allowed, err = rl.CheckAndIncrement(ctx, user2, 3)
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestRateLimiter_SlidingWindow(t *testing.T) {
	rdb := setupMiniredis(t)
	rl := NewRateLimiter(rdb)
	ctx := context.Background()
	userID := uuid.New()

	// Manually add entries with old timestamps (>60s ago) to simulate expired window entries
	key := rateLimitKeyPrefix + userID.String()
	oldTime := float64(time.Now().Add(-70 * time.Second).UnixMilli())
	for i := 0; i < 3; i++ {
		rdb.ZAdd(ctx, key, redis.Z{
			Score:  oldTime + float64(i),
			Member: fmt.Sprintf("old:%d", i),
		})
	}

	// Verify we have 3 entries
	count, err := rdb.ZCard(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// CheckAndIncrement should clean old entries and allow the request
	allowed, err := rl.CheckAndIncrement(ctx, userID, 3)
	require.NoError(t, err)
	assert.True(t, allowed, "old entries should be cleaned, allowing new request")

	// Verify usage is now 1 (old entries removed, 1 new added)
	usage, err := rl.GetMinuteUsage(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 1, usage)
}

func TestRateLimiter_GetMinuteUsageEmpty(t *testing.T) {
	rdb := setupMiniredis(t)
	rl := NewRateLimiter(rdb)
	ctx := context.Background()

	usage, err := rl.GetMinuteUsage(ctx, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 0, usage)
}
