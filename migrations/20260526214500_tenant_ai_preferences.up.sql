CREATE TABLE IF NOT EXISTS tenant_ai_preferences (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    profile VARCHAR(32) NOT NULL DEFAULT 'balanced',
    global_default JSONB NOT NULL DEFAULT '{}'::jsonb,
    use_same_model_everywhere BOOLEAN NOT NULL DEFAULT false,
    defaults JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled_models JSONB NOT NULL DEFAULT '[]'::jsonb,
    allow_user_overrides BOOLEAN NOT NULL DEFAULT true,
    routing_strategy VARCHAR(32) NOT NULL DEFAULT 'quality_first',
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (profile IN ('fast', 'balanced', 'premium', 'custom')),
    CHECK (routing_strategy IN ('quality_first', 'balanced', 'cost_optimized', 'cost_first'))
);

CREATE INDEX IF NOT EXISTS idx_tenant_ai_preferences_updated_at
    ON tenant_ai_preferences (updated_at DESC);
