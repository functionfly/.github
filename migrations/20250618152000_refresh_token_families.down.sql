-- Remove family_id column for refresh token rotation

DROP INDEX IF EXISTS idx_refresh_tokens_family_active;
DROP INDEX IF EXISTS idx_refresh_tokens_family;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS family_id;
