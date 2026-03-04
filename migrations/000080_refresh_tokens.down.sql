-- Drop refresh tokens table and related objects
DROP TRIGGER IF EXISTS trigger_refresh_tokens_updated_at ON refresh_tokens;
DROP FUNCTION IF EXISTS update_refresh_tokens_updated_at();
DROP TABLE IF EXISTS refresh_tokens;