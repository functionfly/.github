-- Plugin Versions Table
-- Tracks version history for rollback support

CREATE TABLE IF NOT EXISTS plugin_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    changelog TEXT,
    manifest JSONB NOT NULL,
    size_bytes INTEGER DEFAULT 0,
    signature TEXT,
    release_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(plugin_id, version)
);

CREATE INDEX idx_plugin_versions_plugin ON plugin_versions(plugin_id);
CREATE INDEX idx_plugin_versions_release_at ON plugin_versions(release_at DESC);

COMMENT ON TABLE plugin_versions IS 'Version history for plugins - supports rollback functionality';