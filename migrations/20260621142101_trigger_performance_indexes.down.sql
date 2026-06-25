-- Migration: 20260621142101_trigger_performance_indexes.down.sql
-- Description: Rollback performance indexes added for trigger_event_queue and agent_identities

DROP INDEX IF EXISTS idx_trigger_queue_pending_ready;
DROP INDEX IF EXISTS idx_agent_identities_parent_covering;
DROP INDEX IF EXISTS idx_state_triggers_active_covering;
