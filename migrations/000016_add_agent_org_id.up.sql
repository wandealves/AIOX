ALTER TABLE agents ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);
CREATE INDEX IF NOT EXISTS idx_agents_org ON agents (org_id) WHERE org_id IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE user_quotas ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);
