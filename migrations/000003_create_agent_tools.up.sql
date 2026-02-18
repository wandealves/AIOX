CREATE TABLE IF NOT EXISTS agent_tools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    tool_name TEXT NOT NULL,
    policy JSONB NOT NULL DEFAULT '{}'::jsonb,
    worker_type TEXT NOT NULL DEFAULT 'python',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_agent_tools_unique ON agent_tools (agent_id, tool_name);
CREATE INDEX idx_agent_tools_agent ON agent_tools (agent_id);
