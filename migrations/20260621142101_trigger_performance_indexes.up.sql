-- Migration: 20260621142101_trigger_performance_indexes.up.sql
-- Description: Add performance indexes for trigger_event_queue polling and agent_identities queries
-- Created: 2026-06-21

-- ============================================================================
-- idx_trigger_queue_pending_ready
-- Purpose: Optimize the trigger event queue polling query that selects pending
-- events ordered by next_attempt_at. The existing idx_trigger_queue_poll uses
-- NOW() in its predicate which causes a mismatch with the application's
-- time.Now() value, leading to suboptimal index usage.
-- Query: SELECT * FROM trigger_event_queue WHERE status = 'pending'
--        AND next_attempt_at <= '...' ORDER BY next_attempt_at ASC LIMIT 100
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_trigger_queue_pending_ready
ON trigger_event_queue(next_attempt_at ASC)
WHERE status = 'pending' AND next_attempt_at IS NOT NULL;

-- ============================================================================
-- idx_agent_identities_parent_covering
-- Purpose: Avoid heap fetches when querying agent_identities by parent_agent_id.
-- The table has large JSONB columns (daemon_config, state_snapshot, capabilities)
-- that bloat row size. This covering index allows index-only scans for common
-- query columns.
-- Query: SELECT * FROM agent_identities WHERE parent_agent_id = '...'
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_agent_identities_parent_covering
ON agent_identities(parent_agent_id)
INCLUDE (tenant_id, agent_id, name, swarm_role, status, capabilities);

-- ============================================================================
-- idx_state_triggers_active_covering
-- Purpose: Improve performance of "SELECT * FROM state_triggers WHERE is_active = true"
-- The existing idx_state_triggers_is_active is ineffective because boolean
-- columns with skewed distributions (most triggers are active) cause the
-- planner to prefer sequential scans. A covering index helps.
-- Query: SELECT * FROM state_triggers WHERE is_active = true
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_state_triggers_active_covering
ON state_triggers(is_active)
INCLUDE (tenant_id, trigger_type, target_function_id, target_function,
         source_state_id, key_pattern, condition, include_previous,
         include_new, max_invocations_per_minute, last_triggered_at);
