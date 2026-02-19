package quota

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles user_quotas PostgreSQL operations.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new quota Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetOrCreate returns the user's quota row, creating one if it doesn't exist.
func (r *Repository) GetOrCreate(ctx context.Context, userID uuid.UUID) (*UserQuota, error) {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_quotas (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`, userID)
	if err != nil {
		return nil, fmt.Errorf("ensuring user quota: %w", err)
	}

	var q UserQuota
	err = r.pool.QueryRow(ctx,
		`SELECT user_id, tokens_used_today, tokens_used_minute, requests_today,
		        last_minute_reset, last_daily_reset, updated_at
		 FROM user_quotas WHERE user_id = $1`, userID,
	).Scan(&q.UserID, &q.TokensUsedToday, &q.TokensUsedMinute, &q.RequestsToday,
		&q.LastMinuteReset, &q.LastDailyReset, &q.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("fetching user quota: %w", err)
	}
	return &q, nil
}

// IncrementDaily adds tokens and increments the request count for the day.
func (r *Repository) IncrementDaily(ctx context.Context, userID uuid.UUID, tokens int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_quotas
		 SET tokens_used_today = tokens_used_today + $2,
		     requests_today = requests_today + 1,
		     updated_at = NOW()
		 WHERE user_id = $1`, userID, tokens)
	if err != nil {
		return fmt.Errorf("incrementing daily quota: %w", err)
	}
	return nil
}

// ResetDailyIfStale resets daily counters if last reset was more than 24h ago.
// Returns true if a reset was performed.
func (r *Repository) ResetDailyIfStale(ctx context.Context, userID uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE user_quotas
		 SET tokens_used_today = 0,
		     requests_today = 0,
		     last_daily_reset = NOW(),
		     updated_at = NOW()
		 WHERE user_id = $1 AND last_daily_reset < NOW() - INTERVAL '24 hours'`, userID)
	if err != nil {
		return false, fmt.Errorf("resetting daily quota: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// RecordViolation appends a violation entry to the violations JSONB array.
func (r *Repository) RecordViolation(ctx context.Context, userID uuid.UUID, violation string) error {
	entry := map[string]any{
		"type":      violation,
		"timestamp": time.Now().UTC(),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling violation: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`UPDATE user_quotas
		 SET violations = violations || $2::jsonb,
		     updated_at = NOW()
		 WHERE user_id = $1`, userID, string(data))
	if err != nil {
		return fmt.Errorf("recording violation: %w", err)
	}
	return nil
}
