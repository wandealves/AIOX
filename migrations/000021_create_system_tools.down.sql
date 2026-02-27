ALTER TABLE agent_tools DROP COLUMN IF EXISTS output_schema;
ALTER TABLE agent_tools DROP COLUMN IF EXISTS category;
ALTER TABLE agent_tools DROP COLUMN IF EXISTS version;

DROP TABLE IF EXISTS system_tools;
