-- system_tools: global tools available to all agents
CREATE TABLE system_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    category VARCHAR(50) DEFAULT 'utilities',
    parameters JSONB DEFAULT '{}'::jsonb,
    output_schema JSONB DEFAULT '{}'::jsonb,
    tool_type VARCHAR(20) DEFAULT 'builtin',
    endpoint_url TEXT DEFAULT '',
    http_method TEXT DEFAULT 'GET',
    headers JSONB DEFAULT '{}'::jsonb,
    auth_type TEXT DEFAULT '',
    auth_config JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    timeout_sec INT DEFAULT 30,
    version VARCHAR(20) DEFAULT '1.0.0',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_system_tools_category ON system_tools (category);
CREATE INDEX idx_system_tools_active ON system_tools (is_active) WHERE is_active = true;

-- Add versioning and category to existing agent_tools
ALTER TABLE agent_tools ADD COLUMN IF NOT EXISTS version VARCHAR(20) DEFAULT '1.0.0';
ALTER TABLE agent_tools ADD COLUMN IF NOT EXISTS category VARCHAR(50) DEFAULT 'custom';
ALTER TABLE agent_tools ADD COLUMN IF NOT EXISTS output_schema JSONB DEFAULT '{}'::jsonb;
