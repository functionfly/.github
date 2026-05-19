-- Plugin Permissions Table
-- Granular permission grants per plugin

CREATE TYPE plugin_permission_type AS ENUM ('network', 'filesystem', 'agents', 'memory', 'terminal', 'gpu', 'webhooks', 'api_keys', 'secrets');

CREATE TABLE IF NOT EXISTS plugin_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    permission_type plugin_permission_type NOT NULL,
    permission_action VARCHAR(50) NOT NULL,
    resource TEXT,
    granted BOOLEAN DEFAULT false,
    granted_at TIMESTAMPTZ,
    granted_by UUID,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(plugin_id, permission_type, resource)
);

CREATE INDEX idx_plugin_permissions_plugin ON plugin_permissions(plugin_id);
CREATE INDEX idx_plugin_permissions_type ON plugin_permissions(permission_type);
CREATE INDEX idx_plugin_permissions_granted ON plugin_permissions(granted);

COMMENT ON TABLE plugin_permissions IS 'Granular permission grants per plugin for capability isolation';