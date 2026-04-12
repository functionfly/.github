-- Allow NULL in audit_events actor_email column and set default
ALTER TABLE audit_events ALTER COLUMN actor_email DROP NOT NULL;

-- Update existing NULL values to empty string to prevent scan errors
UPDATE audit_events SET actor_email = '' WHERE actor_email IS NULL;
