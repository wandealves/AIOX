ALTER TABLE system_tools DROP COLUMN IF EXISTS max_chain_depth;
ALTER TABLE system_tools DROP COLUMN IF EXISTS depends_on;
ALTER TABLE agent_tools DROP COLUMN IF EXISTS max_chain_depth;
ALTER TABLE agent_tools DROP COLUMN IF EXISTS depends_on;
ALTER TABLE system_tools DROP COLUMN IF EXISTS plugin_id;

DROP TABLE IF EXISTS plugins;
