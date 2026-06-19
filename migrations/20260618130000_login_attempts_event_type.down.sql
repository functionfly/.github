-- Remove event_type column from login_attempts
ALTER TABLE login_attempts DROP COLUMN IF EXISTS event_type;

-- Drop the event_type index
DROP INDEX IF EXISTS idx_login_attempts_event_type;
