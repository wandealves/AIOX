CREATE TABLE tool_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_id UUID NOT NULL,
    tool_source VARCHAR(10) NOT NULL,
    grantee_type VARCHAR(20) NOT NULL,
    grantee_id VARCHAR(255) NOT NULL,
    permission VARCHAR(20) DEFAULT 'execute',
    granted_by UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tool_id, tool_source, grantee_type, grantee_id)
);

CREATE INDEX idx_tool_permissions_tool ON tool_permissions (tool_id, tool_source);
CREATE INDEX idx_tool_permissions_grantee ON tool_permissions (grantee_type, grantee_id);
