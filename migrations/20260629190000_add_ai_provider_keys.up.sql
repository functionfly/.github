CREATE TABLE IF NOT EXISTS ai_provider_keys (
    id              VARCHAR(255) PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    provider        VARCHAR(50) NOT NULL,
    encrypted_key   BYTEA NOT NULL,
    key_nonce       BYTEA NOT NULL,
    key_tag         BYTEA NOT NULL,
    key_version     INT NOT NULL DEFAULT 1,
    key_last4       VARCHAR(4) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    health_message  TEXT,
    last_health_check TIMESTAMPTZ,
    last_used_at    TIMESTAMPTZ,
    connected_by    UUID NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, provider)
);

CREATE INDEX idx_ai_provider_keys_tenant ON ai_provider_keys(tenant_id);
CREATE INDEX idx_ai_provider_keys_status ON ai_provider_keys(status) WHERE status != 'active';

DO $$ BEGIN
    ALTER TABLE ai_provider_keys ADD CONSTRAINT fk_ai_provider_keys_tenant
        FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE ai_provider_keys ADD CONSTRAINT fk_ai_provider_keys_connected_by
        FOREIGN KEY (connected_by) REFERENCES users(id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
