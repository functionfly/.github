-- Remove hash chaining columns from audit_events
ALTER TABLE audit_events DROP COLUMN IF EXISTS previous_hash;
ALTER TABLE audit_events DROP COLUMN IF EXISTS event_hash;

DROP INDEX IF EXISTS idx_audit_events_hash_chain;
