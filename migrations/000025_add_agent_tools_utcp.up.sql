ALTER TABLE agent_tools
    ADD COLUMN IF NOT EXISTS utcp_manual_id UUID REFERENCES utcp_manuals(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS tool_type TEXT DEFAULT 'http';
