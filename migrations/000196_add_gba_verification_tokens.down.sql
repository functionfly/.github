-- Remove verification tokens table
DROP INDEX IF EXISTS idx_gba_verification_tokens_identifier;
DROP INDEX IF EXISTS idx_gba_verification_tokens_token;
DROP INDEX IF EXISTS idx_gba_verification_tokens_tenant;
DROP INDEX IF EXISTS idx_gba_verification_tokens_expires;
DROP TABLE IF EXISTS gba_verification_tokens;