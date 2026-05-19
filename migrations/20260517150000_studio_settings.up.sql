-- Studio Settings table for persisting user studio preferences
CREATE TABLE IF NOT EXISTS studio_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID DEFAULT NULL,
    environment VARCHAR(100) DEFAULT '',
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, environment)
);

CREATE INDEX IF NOT EXISTS idx_studio_settings_tenant ON studio_settings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_studio_settings_user ON studio_settings(user_id);