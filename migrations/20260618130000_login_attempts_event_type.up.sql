-- Add event_type column to login_attempts for MFA lockout tracking
ALTER TABLE login_attempts ADD COLUMN IF NOT EXISTS event_type VARCHAR(50) DEFAULT NULL;

-- Index for filtering by event_type (MFA lockout queries)
CREATE INDEX IF NOT EXISTS idx_login_attempts_event_type ON login_attempts(event_type) WHERE event_type IS NOT NULL;

-- Backfill existing MFA lockout records with appropriate event_type
-- Lockout records with non-null lockout_until that don't have an event_type are MFA lockouts
UPDATE login_attempts
SET event_type = 'mfa_lockout'
WHERE lockout_until IS NOT NULL
  AND event_type IS NULL
  AND successful = false
  AND id IN (
    SELECT id FROM login_attempts
    WHERE lockout_until IS NOT NULL
      AND event_type IS NULL
    ORDER BY attempted_at DESC
  );
