-- Add hash chaining to audit_events for integrity verification
-- This enables detection of tampering, deletion, or insertion of audit events

-- Add columns for hash chaining
ALTER TABLE audit_events ADD COLUMN IF NOT EXISTS previous_hash VARCHAR(64);
ALTER TABLE audit_events ADD COLUMN IF NOT EXISTS event_hash VARCHAR(64);

-- Create index for hash verification queries
CREATE INDEX IF NOT EXISTS idx_audit_events_hash_chain ON audit_events(tenant_id, timestamp DESC, event_hash);
