-- Migration: Deprecate agent_wallets table
--
-- This migration marks the agent_wallets table as deprecated.
-- The unified wallets table now serves as the source of truth for agent wallets.

BEGIN;

-- Add deprecation flag to agent_wallets
ALTER TABLE agent_wallets ADD COLUMN deprecated_at TIMESTAMPTZ;

-- Mark all existing records as deprecated
UPDATE agent_wallets SET deprecated_at = NOW() WHERE deprecated_at IS NULL;

-- Add check constraint to prevent new writes (deprecated_at must be NOT NULL)
ALTER TABLE agent_wallets ADD CONSTRAINT agent_wallets_deprecated_check CHECK (deprecated_at IS NOT NULL);

COMMENT ON TABLE agent_wallets IS 'DEPRECATED: Use wallets table instead (owner_type=agent). This table is retained for backwards compatibility only.';

COMMIT;