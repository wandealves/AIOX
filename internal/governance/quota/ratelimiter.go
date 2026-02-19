package quota

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	rateLimitKeyPrefix = "quota:minute:"
	windowDuration     = 60 * time.Second
	keyTTL             = 90 * time.Second
)

// RateLimiter implements a Redis sorted-set sliding window for per-minute rate limiting.
type RateLimiter struct {
	rdb redis.Cmdable
}

// NewRateLimiter creates a new Redis-based rate limiter.
func NewRateLimiter(rdb redis.Cmdable) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// CheckAndIncrement checks whether the user is under the per-minute limit.
// If under limit, it increments the counter and returns true (allowed).
// If over limit, it returns false (denied).
func (rl *RateLimiter) CheckAndIncrement(ctx context.Context, userID uuid.UUID, maxPerMinute int) (bool, error) {
	key := rateLimitKeyPrefix + userID.String()
	now := time.Now()
	nowMs := float64(now.UnixMilli())
	windowStart := float64(now.Add(-windowDuration).UnixMilli())

	pipe := rl.rdb.Pipeline()

	// Remove entries older than the window
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatFloat(windowStart, 'f', 0, 64))

	// Count current entries in the window
	countCmd := pipe.ZCard(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("rate limiter pipeline (clean+count): %w", err)
	}

	count := countCmd.Val()
	if count >= int64(maxPerMinute) {
		return false, nil
	}

	// Under limit: add new entry and set TTL
	pipe2 := rl.rdb.Pipeline()
	member := fmt.Sprintf("%d:%d", now.UnixNano(), count)
	pipe2.ZAdd(ctx, key, redis.Z{Score: nowMs, Member: member})
	pipe2.Expire(ctx, key, keyTTL)

	_, err = pipe2.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("rate limiter pipeline (add): %w", err)
	}

	return true, nil
}

// GetMinuteUsage returns the current number of requests in the sliding window.
func (rl *RateLimiter) GetMinuteUsage(ctx context.Context, userID uuid.UUID) (int, error) {
	key := rateLimitKeyPrefix + userID.String()
	now := time.Now()
	windowStart := float64(now.Add(-windowDuration).UnixMilli())
	nowMs := float64(now.UnixMilli())

	count, err := rl.rdb.ZCount(ctx, key, strconv.FormatFloat(windowStart, 'f', 0, 64), strconv.FormatFloat(nowMs, 'f', 0, 64)).Result()
	if err != nil {
		return 0, fmt.Errorf("getting minute usage: %w", err)
	}
	return int(count), nil
}
