BEGIN;

-- Add optional team_id to existing api_keys table
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES teams(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_team_id ON api_keys(team_id) WHERE team_id IS NOT NULL;

-- Remove redundant team_api_keys table (superseded by api_keys.team_id)
DROP TABLE IF EXISTS team_api_keys;

COMMIT;
