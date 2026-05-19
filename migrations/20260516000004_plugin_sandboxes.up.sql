-- Plugin Sandbox Configuration Table
-- Multi-tier execution sandbox settings

CREATE TYPE sandbox_tier AS ENUM ('wasm', 'worker', 'microvm', 'enterprise');

CREATE TABLE IF NOT EXISTS plugin_sandboxes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE UNIQUE,
    tier sandbox_tier NOT NULL DEFAULT 'worker',
    cpu_limit DECIMAL(5,2) DEFAULT 0.5,
    memory_limit_mb INTEGER DEFAULT 256,
    timeout_seconds INTEGER DEFAULT 30,
    network_isolated BOOLEAN DEFAULT true,
    filesystem_scope VARCHAR(50) DEFAULT 'workspace',
    max_instances INTEGER DEFAULT 1,
    env_vars JSONB DEFAULT '{}',
    allowed_domains TEXT[],
    blocked_domains TEXT[],
    rate_limit_rpm INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_plugin_sandboxes_plugin ON plugin_sandboxes(plugin_id);
CREATE INDEX idx_plugin_sandboxes_tier ON plugin_sandboxes(tier);

COMMENT ON TABLE plugin_sandboxes IS 'Sandbox configuration per plugin - defines execution isolation tier and resource limits';