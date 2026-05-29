ALTER TABLE tenant_memberships DROP COLUMN IF EXISTS last_active_at;
ALTER TABLE users DROP COLUMN IF EXISTS settings;
