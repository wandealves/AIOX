CREATE INDEX IF NOT EXISTS idx_executions_conversation ON executions (agent_id, owner_user_id, created_at DESC);
