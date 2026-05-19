-- Remove environment column from studio_collab_events

DROP INDEX IF EXISTS idx_studio_collab_events_tenant_environment;
DROP INDEX IF EXISTS idx_studio_collab_events_environment;

ALTER TABLE studio_collab_events DROP COLUMN IF EXISTS environment;