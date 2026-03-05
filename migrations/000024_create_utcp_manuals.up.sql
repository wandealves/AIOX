CREATE TABLE IF NOT EXISTS utcp_manuals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    manual_url TEXT NOT NULL,
    utcp_version TEXT DEFAULT '',
    auth_template JSONB DEFAULT '{}'::jsonb,
    auth_values TEXT DEFAULT '',
    is_active BOOLEAN DEFAULT true,
    last_synced_at TIMESTAMPTZ,
    tools_count INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_utcp_manuals_agent_name ON utcp_manuals (agent_id, name);
