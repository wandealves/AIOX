CREATE TABLE scheduled_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id),
    agent_id UUID REFERENCES agents(id),
    pipeline_id UUID REFERENCES pipelines(id),
    name TEXT NOT NULL,
    cron_expression TEXT NOT NULL,
    input_template TEXT NOT NULL DEFAULT '',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_target CHECK (agent_id IS NOT NULL OR pipeline_id IS NOT NULL)
);

CREATE INDEX idx_scheduled_tasks_owner ON scheduled_tasks (owner_user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_scheduled_tasks_next_run ON scheduled_tasks (next_run_at) WHERE is_active = true AND deleted_at IS NULL;
