-- Migration: secret_versions_table (down)
-- Description: Remove secret versioning and history table

-- Remove trigger first
DROP TRIGGER IF EXISTS trigger_secret_version_on_create ON secrets_vault;

-- Remove trigger function
DROP FUNCTION IF EXISTS auto_create_initial_version();

-- Remove helper functions
DROP FUNCTION IF EXISTS get_next_secret_version(UUID);
DROP FUNCTION IF EXISTS get_secret_current_version(UUID);

-- Remove columns from secrets_vault
ALTER TABLE secrets_vault 
    DROP COLUMN IF EXISTS current_version,
    DROP COLUMN IF EXISTS last_modified_by,
    DROP COLUMN IF EXISTS last_modified_at;

-- Drop the secret_versions table
DROP TABLE IF EXISTS secret_versions;
