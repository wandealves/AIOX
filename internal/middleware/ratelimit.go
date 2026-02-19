package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter provides per-IP sliding-window rate limiting backed by Redis sorted sets.
type RateLimiter struct {
	client    redis.Cmdable
	maxReqs   int
	windowSec int
}

// NewRateLimiter creates a rate limiter that allows maxReqs per windowSec seconds.
func NewRateLimiter(client redis.Cmdable, maxReqs, windowSec int) *RateLimiter {
	return &RateLimiter{client: client, maxReqs: maxReqs, windowSec: windowSec}
}

// Middleware returns an HTTP middleware that enforces the rate limit.
// On Redis errors it fails open (allows the request through).
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		key := "ratelimit:auth:" + ip

		allowed, err := rl.allow(r.Context(), key)
		if err != nil {
			slog.Warn("rate limiter: redis error, failing open", "error", err, "ip", ip)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(rl.windowSec))
			http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ctx context.Context, key string) (bool, error) {
	now := time.Now()
	windowStart := float64(now.Add(-time.Duration(rl.windowSec) * time.Second).UnixMilli())
	member := fmt.Sprintf("%d", now.UnixNano())
	score := float64(now.UnixMilli())

	pipe := rl.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%f", windowStart))
	countCmd := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: member})
	pipe.Expire(ctx, key, time.Duration(rl.windowSec)*time.Second+time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return countCmd.Val() < int64(rl.maxReqs), nil
}

func clientIP(r *http.Request) string {
	// Check X-Forwarded-For first (trusted reverse proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
