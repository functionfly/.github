-- Migration: Rollback Automated Retention and Archival System

-- Drop views
DROP VIEW IF EXISTS retention_summary;

-- Drop functions (order matters - dependent functions first)
DROP FUNCTION IF EXISTS confirm_source_deleted(UUID);
DROP FUNCTION IF EXISTS verify_archive_batch(UUID, BOOLEAN, TEXT);
DROP FUNCTION IF EXISTS prepare_archive_batch(TEXT, TIMESTAMPTZ, TIMESTAMPTZ, BIGINT, BIGINT, TEXT, TEXT, TEXT, TEXT);
DROP FUNCTION IF EXISTS execute_retention_cleanup(BOOLEAN, TEXT[]);
DROP FUNCTION IF EXISTS cleanup_registry_executions(INTEGER, BOOLEAN);
DROP FUNCTION IF EXISTS cleanup_cost_allocation_entries(INTEGER, BOOLEAN);
DROP FUNCTION IF EXISTS get_retention_cutoff(TEXT, INTEGER);
DROP FUNCTION IF EXISTS lift_legal_hold(UUID, UUID, TEXT);
DROP FUNCTION IF EXISTS create_legal_hold(TEXT, TEXT, TEXT, UUID[], TEXT[], UUID[], DATE, DATE, UUID);
DROP FUNCTION IF EXISTS is_under_legal_hold(TEXT, UUID, UUID);

-- Drop tables (in order to avoid FK conflicts)
DROP TABLE IF EXISTS archive_batches;
DROP TABLE IF EXISTS retention_audit_log;
DROP TABLE IF EXISTS legal_holds;
DROP TABLE IF EXISTS archive_configurations;
