-- Add soft-delete support to factory_config (GORM DeletedAt)
ALTER TABLE factory_config
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
-- Partial index for active config lookups (agent_id + deleted_at IS NULL)
CREATE INDEX IF NOT EXISTS idx_factory_config_agent_id_deleted_at ON factory_config(agent_id) WHERE deleted_at IS NULL;
