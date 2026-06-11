CREATE TABLE IF NOT EXISTS connector_oauth_states (
    state       VARCHAR(255) PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    connector_id UUID NOT NULL REFERENCES connectors(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_connector_oauth_states_tenant ON connector_oauth_states(tenant_id);
CREATE INDEX IF NOT EXISTS idx_connector_oauth_states_created ON connector_oauth_states(created_at);
