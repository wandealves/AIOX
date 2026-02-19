package quota

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/config"
)

// Service orchestrates Redis rate limiting and PostgreSQL quota tracking.
type Service struct {
	repo    *Repository
	limiter *RateLimiter
	cfg     config.GovernanceCfg
}

// NewService creates a new quota Service.
func NewService(repo *Repository, limiter *RateLimiter, cfg config.GovernanceCfg) *Service {
	return &Service{
		repo:    repo,
		limiter: limiter,
		cfg:     cfg,
	}
}

// CheckQuota verifies that the user has not exceeded rate or daily limits.
// Returns nil if allowed, or an error describing the exceeded limit.
func (s *Service) CheckQuota(ctx context.Context, userID uuid.UUID) error {
	// 1. Redis sliding-window per-minute rate limit (fast path)
	allowed, err := s.limiter.CheckAndIncrement(ctx, userID, s.cfg.MaxTokensPerMinute)
	if err != nil {
		slog.Warn("quota: rate limiter check failed, allowing request", "error", err)
		// Fail open on Redis errors to not block the user
	} else if !allowed {
		_ = s.repo.RecordViolation(ctx, userID, "rate_limit_minute")
		return fmt.Errorf("rate limit exceeded: max %d requests per minute", s.cfg.MaxTokensPerMinute)
	}

	// 2. PostgreSQL daily limits
	// Reset daily counters if stale
	if _, err := s.repo.ResetDailyIfStale(ctx, userID); err != nil {
		slog.Warn("quota: daily reset check failed", "error", err)
	}

	quota, err := s.repo.GetOrCreate(ctx, userID)
	if err != nil {
		slog.Warn("quota: failed to get quota, allowing request", "error", err)
		return nil // Fail open
	}

	if quota.TokensUsedToday >= s.cfg.MaxTokensPerDay {
		_ = s.repo.RecordViolation(ctx, userID, "daily_token_limit")
		return fmt.Errorf("daily token limit exceeded: %d/%d tokens used", quota.TokensUsedToday, s.cfg.MaxTokensPerDay)
	}

	if quota.RequestsToday >= s.cfg.MaxRequestsPerDay {
		_ = s.repo.RecordViolation(ctx, userID, "daily_request_limit")
		return fmt.Errorf("daily request limit exceeded: %d/%d requests", quota.RequestsToday, s.cfg.MaxRequestsPerDay)
	}

	return nil
}

// DeductTokens records token usage after a successful worker response.
func (s *Service) DeductTokens(ctx context.Context, userID uuid.UUID, tokensUsed int) error {
	return s.repo.IncrementDaily(ctx, userID, tokensUsed)
}

// GetQuota returns the user's current quota status for API display.
func (s *Service) GetQuota(ctx context.Context, userID uuid.UUID) (*QuotaStatus, error) {
	// Reset if stale before reading
	if _, err := s.repo.ResetDailyIfStale(ctx, userID); err != nil {
		slog.Warn("quota: daily reset check failed", "error", err)
	}

	quota, err := s.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting quota: %w", err)
	}

	minuteUsage, err := s.limiter.GetMinuteUsage(ctx, userID)
	if err != nil {
		slog.Warn("quota: failed to get minute usage", "error", err)
		minuteUsage = 0
	}

	return &QuotaStatus{
		TokensUsedToday:   quota.TokensUsedToday,
		TokensLimitDay:    s.cfg.MaxTokensPerDay,
		RequestsToday:     quota.RequestsToday,
		RequestsLimitDay:  s.cfg.MaxRequestsPerDay,
		TokensUsedMinute:  minuteUsage,
		TokensLimitMinute: s.cfg.MaxTokensPerMinute,
	}, nil
}
