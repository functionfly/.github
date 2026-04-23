-- Migration: Remove trigger execution engine tables
-- Reverses the changes made by 20260419214800_trigger_execution_engine.up.sql

-- Drop triggers
DROP TRIGGER IF EXISTS update_trigger_event_queue_updated_at ON trigger_event_queue;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS cleanup_old_trigger_logs(INTEGER);
DROP FUNCTION IF EXISTS retry_dead_letter_event(UUID);
DROP FUNCTION IF EXISTS get_trigger_execution_stats(UUID, INTEGER);
DROP FUNCTION IF EXISTS get_tenant_trigger_queue_stats(UUID);

-- Drop tables (cascade to remove indexes)
DROP TABLE IF EXISTS trigger_dead_letter;
DROP TABLE IF EXISTS trigger_execution_logs;
DROP TABLE IF EXISTS trigger_event_queue;
