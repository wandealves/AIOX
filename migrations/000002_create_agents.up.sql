CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    jid TEXT NOT NULL,
    profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    llm_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    capabilities JSONB NOT NULL DEFAULT '{}'::jsonb,
    memory_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    governance JSONB NOT NULL DEFAULT '{}'::jsonb,
    visibility TEXT NOT NULL DEFAULT 'private',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_agents_jid ON agents (jid);
CREATE INDEX idx_agents_owner ON agents (owner_user_id);
CREATE INDEX idx_agents_active ON agents (owner_user_id) WHERE deleted_at IS NULL;
