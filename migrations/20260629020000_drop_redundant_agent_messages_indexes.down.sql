-- Rollback: 20260629020000_drop_redundant_agent_messages_indexes
--
-- This rollback recreates indexes that were dropped to fix slow INSERTs.
-- Only for emergency rollback - index proliferation causes 10+ second INSERT latency.

BEGIN;

-- Recreate the redundant indexes
CREATE INDEX IF NOT EXISTS idx_agent_messages_created_at_brin
    ON agent_messages USING brin (created_at);

CREATE INDEX IF NOT EXISTS idx_agent_messages_to_agent_brin
    ON agent_messages USING brin (to_agent_id, created_at)
    WITH (pages_per_range = '4');

CREATE INDEX IF NOT EXISTS idx_agent_messages_pending_brin
    ON agent_messages USING brin (to_agent_id, created_at)
    WHERE status = ANY (ARRAY['pending'::text, 'delivered'::text]);

CREATE INDEX IF NOT EXISTS idx_agent_messages_id_status
    ON agent_messages USING btree (id, status)
    WHERE status = 'pending'::text;

-- Update statistics
ANALYZE agent_messages;

COMMIT;
