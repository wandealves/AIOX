CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs (resource_id, created_at DESC) WHERE resource_id IS NOT NULL;
