ALTER TABLE agent_tools
    DROP COLUMN IF EXISTS utcp_manual_id,
    DROP COLUMN IF EXISTS tool_type;
