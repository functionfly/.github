DROP INDEX IF EXISTS idx_factory_config_agent_id_deleted_at;
ALTER TABLE factory_config DROP COLUMN IF EXISTS deleted_at;
