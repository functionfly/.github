-- Rollback database encryption support
-- +migrate Down

-- Drop indexes first
DROP INDEX IF EXISTS idx_encryption_keys_version_purpose;
DROP INDEX IF EXISTS idx_encryption_keys_active;
DROP INDEX IF EXISTS idx_encrypted_fields_table_column;

-- Drop tables
DROP TABLE IF EXISTS encrypted_fields;
DROP TABLE IF EXISTS encryption_keys;