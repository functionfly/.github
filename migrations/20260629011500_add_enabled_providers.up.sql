ALTER TABLE tenant_ai_preferences
  ADD COLUMN IF NOT EXISTS enabled_providers JSONB NOT NULL DEFAULT '[]'::jsonb;
