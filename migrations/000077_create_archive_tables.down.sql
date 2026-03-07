-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_archive_updated_at ON archive_data;

-- Drop functions
DROP FUNCTION IF EXISTS update_archive_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_archive_data_metadata_gin;
DROP INDEX IF EXISTS idx_archive_cleanup_started_at;
DROP INDEX IF EXISTS idx_archive_cleanup_operation;
DROP INDEX IF EXISTS idx_archive_cleanup_archive_id;
DROP INDEX IF EXISTS idx_archive_data_storage_key;
DROP INDEX IF EXISTS idx_archive_data_created_at;
DROP INDEX IF EXISTS idx_archive_data_status;
DROP INDEX IF EXISTS idx_archive_data_type;

-- Drop tables
DROP TABLE IF EXISTS archive_cleanup_log;
DROP TABLE IF EXISTS archive_data;