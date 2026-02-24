ALTER TABLE agent_tools
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS parameters,
    DROP COLUMN IF EXISTS endpoint_url,
    DROP COLUMN IF EXISTS http_method,
    DROP COLUMN IF EXISTS headers,
    DROP COLUMN IF EXISTS auth_type,
    DROP COLUMN IF EXISTS auth_config,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS timeout_sec,
    DROP COLUMN IF EXISTS updated_at;
