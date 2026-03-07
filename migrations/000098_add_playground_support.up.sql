-- Add playground support and app_id to functions table
ALTER TABLE functions ADD COLUMN IF NOT EXISTS app_id UUID REFERENCES apps(id) ON DELETE SET NULL;
ALTER TABLE functions ADD COLUMN IF NOT EXISTS playground_enabled BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE functions ADD COLUMN IF NOT EXISTS playground_config JSONB DEFAULT '{}'::jsonb;

-- Create indexes for playground lookups
CREATE INDEX IF NOT EXISTS idx_functions_playground_enabled ON functions(playground_enabled) WHERE playground_enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_functions_app_id ON functions(app_id);
CREATE INDEX IF NOT EXISTS idx_apps_slug ON apps(slug);
