BEGIN;

-- Default invite role per team
ALTER TABLE teams ADD COLUMN IF NOT EXISTS default_invite_role VARCHAR(50) DEFAULT 'member';

-- Team API keys
CREATE TABLE IF NOT EXISTS team_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(12) NOT NULL,
    scopes TEXT[] DEFAULT '{}',
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_team_api_keys_team_id ON team_api_keys(team_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_team_api_keys_key_hash ON team_api_keys(key_hash) WHERE revoked_at IS NULL;

-- Team audit log
CREATE TABLE IF NOT EXISTS team_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id UUID,
    details JSONB DEFAULT '{}',
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_team_audit_log_team_id ON team_audit_log(team_id);
CREATE INDEX IF NOT EXISTS idx_team_audit_log_created_at ON team_audit_log(created_at);

-- Team resource quotas
CREATE TABLE IF NOT EXISTS team_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL,
    max_count INTEGER NOT NULL DEFAULT 0,
    current_count INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(team_id, resource_type)
);
CREATE INDEX IF NOT EXISTS idx_team_quotas_team_id ON team_quotas(team_id);

-- Seed default quotas for existing teams
INSERT INTO team_quotas (team_id, resource_type, max_count)
SELECT t.id, rt, 100
FROM teams t
CROSS JOIN (VALUES ('functions'), ('deployments'), ('api_keys')) AS rt(resource_type)
ON CONFLICT (team_id, resource_type) DO NOTHING;

COMMIT;
