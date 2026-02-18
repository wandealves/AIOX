CREATE TABLE IF NOT EXISTS executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    input TEXT,
    output TEXT,
    tokens_used INT DEFAULT 0,
    tools_called JSONB DEFAULT '[]'::jsonb,
    worker_id TEXT,
    duration_ms INT DEFAULT 0,
    go_latency_ms INT DEFAULT 0,
    python_latency_ms INT DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'completed',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_executions_owner_created ON executions (owner_user_id, created_at DESC);
CREATE INDEX idx_executions_agent ON executions (agent_id, created_at DESC);
