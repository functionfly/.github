-- Migration: Remove trusted device token expiry column

DROP INDEX IF EXISTS idx_sessions_trusted_token_expires;
ALTER TABLE sessions DROP COLUMN IF EXISTS trusted_device_token_expires_at;