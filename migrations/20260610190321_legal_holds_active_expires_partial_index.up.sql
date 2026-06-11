-- Migration: legal_holds_active_expires_partial_index
-- Created at: 2026-06-10T19:03:21-05:00
-- Purpose: Add partial index for HasActiveLegalHolds query pattern
-- The query SELECT 1 FROM legal_holds WHERE status = 'active' AND (expires_at IS NULL OR expires_at > NOW())
-- needs a partial index to avoid full table scans under heavy load.

-- Up migration
BEGIN;

-- Partial index for active legal holds that haven't expired
-- This supports the query: WHERE status = 'active' AND (expires_at IS NULL OR expires_at > NOW())
CREATE INDEX IF NOT EXISTS idx_legal_holds_active_not_expired
    ON legal_holds(expires_at)
    WHERE status = 'active' AND expires_at IS NOT NULL;

-- Partial index for active legal holds with no expiration (never expire)
-- This supports the query branch: WHERE status = 'active' AND expires_at IS NULL
CREATE INDEX IF NOT EXISTS idx_legal_holds_active_no_expiry
    ON legal_holds(id)
    WHERE status = 'active' AND expires_at IS NULL;

-- Composite index for the exact query pattern (covers both NULL and non-NULL expires_at)
-- PostgreSQL can use this for both branches via index skipping
CREATE INDEX IF NOT EXISTS idx_legal_holds_active_any_expiry
    ON legal_holds(status, expires_at)
    WHERE status = 'active';

-- Covering index to avoid heap lookup entirely for the EXISTS query
CREATE INDEX IF NOT EXISTS idx_legal_holds_active_covering
    ON legal_holds(id)
    WHERE status = 'active';

COMMIT;