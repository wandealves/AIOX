package quota

import (
	"time"

	"github.com/google/uuid"
)

// UserQuota matches the user_quotas table schema.
type UserQuota struct {
	UserID           uuid.UUID `json:"user_id"`
	TokensUsedToday  int       `json:"tokens_used_today"`
	TokensUsedMinute int       `json:"tokens_used_minute"`
	RequestsToday    int       `json:"requests_today"`
	LastMinuteReset  time.Time `json:"last_minute_reset"`
	LastDailyReset   time.Time `json:"last_daily_reset"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// QuotaStatus is the API response showing current quota usage and limits.
type QuotaStatus struct {
	TokensUsedToday    int `json:"tokens_used_today"`
	TokensLimitDay     int `json:"tokens_limit_day"`
	RequestsToday      int `json:"requests_today"`
	RequestsLimitDay   int `json:"requests_limit_day"`
	TokensUsedMinute   int `json:"tokens_used_minute"`
	TokensLimitMinute  int `json:"tokens_limit_minute"`
}
