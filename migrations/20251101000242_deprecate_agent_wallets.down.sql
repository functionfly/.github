-- Down migration for 000256_deprecate_agent_wallets
-- Reverses the deprecation of agent_wallets table

BEGIN;

-- Drop the check constraint
ALTER TABLE agent_wallets DROP CONSTRAINT IF EXISTS agent_wallets_deprecated_check;

-- Drop the deprecated_at column
ALTER TABLE agent_wallets DROP COLUMN IF EXISTS deprecated_at;

-- Remove the table comment
COMMENT ON TABLE agent_wallets IS NULL;

COMMIT;