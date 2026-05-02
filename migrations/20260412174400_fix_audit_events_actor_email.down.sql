-- Migration: Revert audit_events actor_email changes
-- Reverses: 20260412174400_fix_audit_events_actor_email

-- Make actor_email NOT NULL again (requires all values to be non-NULL)
ALTER TABLE audit_events ALTER COLUMN actor_email SET NOT NULL;