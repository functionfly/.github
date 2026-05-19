-- Add environment column to studio_collab_events for environment-scoped filtering

ALTER TABLE studio_collab_events ADD COLUMN IF NOT EXISTS environment TEXT NOT NULL DEFAULT '';

-- Backfill empty strings for existing rows (they were created without environment context)
-- No action needed as default is already empty string

-- Add index for environment-filtered queries
CREATE INDEX IF NOT EXISTS idx_studio_collab_events_tenant_environment ON studio_collab_events(tenant_id, environment, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_studio_collab_events_environment ON studio_collab_events(environment, created_at DESC);