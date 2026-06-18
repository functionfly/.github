-- Add TTL fields to state_fabrics table for expiration cleanup
-- This enables the cleanup worker to delete expired state fabrics

ALTER TABLE state_fabrics ADD COLUMN IF NOT EXISTS ttl_days INTEGER NOT NULL DEFAULT 0;
ALTER TABLE state_fabrics ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_state_fabrics_expires_at ON state_fabrics(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_state_fabrics_ttl_days ON state_fabrics(ttl_days) WHERE ttl_days > 0;
