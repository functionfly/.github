-- Rollback: legal_holds_active_expires_partial_index
-- Created at: 2026-06-10T19:03:21-05:00
-- Purpose: Remove partial indexes for legal holds active query

-- Down migration (reverses the up migration)
BEGIN;

DROP INDEX IF EXISTS idx_legal_holds_active_not_expired;
DROP INDEX IF EXISTS idx_legal_holds_active_no_expiry;
DROP INDEX IF EXISTS idx_legal_holds_active_any_expiry;
DROP INDEX IF EXISTS idx_legal_holds_active_covering;

COMMIT;