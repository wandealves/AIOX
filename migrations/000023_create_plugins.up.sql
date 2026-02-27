CREATE TABLE plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    manifest_url TEXT DEFAULT '',
    is_active BOOLEAN DEFAULT true,
    auto_sync BOOLEAN DEFAULT false,
    last_synced_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Link system_tools to plugins
ALTER TABLE system_tools ADD COLUMN IF NOT EXISTS plugin_id UUID REFERENCES plugins(id) ON DELETE CASCADE;

-- Tool chaining support
ALTER TABLE agent_tools ADD COLUMN IF NOT EXISTS depends_on TEXT[] DEFAULT '{}';
ALTER TABLE agent_tools ADD COLUMN IF NOT EXISTS max_chain_depth INT DEFAULT 5;
ALTER TABLE system_tools ADD COLUMN IF NOT EXISTS depends_on TEXT[] DEFAULT '{}';
ALTER TABLE system_tools ADD COLUMN IF NOT EXISTS max_chain_depth INT DEFAULT 5;
