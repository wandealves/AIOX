CREATE TABLE IF NOT EXISTS user_quotas (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    tokens_used_today INT NOT NULL DEFAULT 0,
    tokens_used_minute INT NOT NULL DEFAULT 0,
    requests_today INT NOT NULL DEFAULT 0,
    last_minute_reset TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_daily_reset TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    violations JSONB DEFAULT '[]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
