CREATE TABLE IF NOT EXISTS ai_workers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    worker_id TEXT NOT NULL,
    host TEXT NOT NULL,
    port INT NOT NULL,
    status TEXT NOT NULL DEFAULT 'healthy',
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active_requests INT NOT NULL DEFAULT 0,
    avg_latency_ms INT NOT NULL DEFAULT 0,
    memory_usage_mb INT NOT NULL DEFAULT 0,
    capabilities JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_ai_workers_worker_id ON ai_workers (worker_id);
CREATE INDEX idx_ai_workers_status ON ai_workers (status);
