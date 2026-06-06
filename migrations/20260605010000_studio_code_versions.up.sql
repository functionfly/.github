-- +migrate Up
CREATE TABLE IF NOT EXISTS studio_code_versions (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36),
    environment VARCHAR(50),
    file_path VARCHAR(1024) NOT NULL,
    content TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    action VARCHAR(50) NOT NULL DEFAULT 'save',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_studio_code_versions_tenant_user_env
ON studio_code_versions (tenant_id, user_id, environment, file_path, version DESC);

CREATE INDEX IF NOT EXISTS idx_studio_code_versions_lookup
ON studio_code_versions (tenant_id, COALESCE(user_id, ''), COALESCE(environment, ''), file_path, version);

COMMENT ON TABLE studio_code_versions IS 'Stores versioned snapshots of code for studio editor undo/redo and history';
COMMENT ON COLUMN studio_code_versions.action IS 'Action type: save, undo, redo, format';
COMMENT ON COLUMN studio_code_versions.version IS 'Auto-incrementing version per tenant/user/environment/file_path';

-- +migrate Down
DROP INDEX IF EXISTS idx_studio_code_versions_lookup;
DROP INDEX IF EXISTS idx_studio_code_versions_tenant_user_env;
DROP TABLE IF EXISTS studio_code_versions;