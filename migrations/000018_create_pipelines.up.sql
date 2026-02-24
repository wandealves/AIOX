CREATE TABLE pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id),
    org_id UUID REFERENCES organizations(id),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    steps JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_pipelines_owner ON pipelines (owner_user_id) WHERE deleted_at IS NULL;

CREATE TABLE pipeline_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES pipelines(id),
    owner_user_id UUID NOT NULL REFERENCES users(id),
    input TEXT NOT NULL DEFAULT '',
    output TEXT NOT NULL DEFAULT '',
    current_step INT NOT NULL DEFAULT 0,
    total_steps INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    step_results JSONB NOT NULL DEFAULT '[]',
    total_tokens INT NOT NULL DEFAULT 0,
    total_duration_ms INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pipeline_executions_pipeline ON pipeline_executions (pipeline_id);
CREATE INDEX idx_pipeline_executions_owner ON pipeline_executions (owner_user_id);
