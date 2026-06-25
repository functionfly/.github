-- Migration: 20260619182100_fix_agent_messages_invalid_indexes
--
-- Problem: Slow queries (3.4-3.8s) on agent_messages despite index existing
-- Root cause: CREATE INDEX CONCURRENTLY IF NOT EXISTS leaves indexes in
--             'invalid' state after failure, but subsequent runs do nothing.
--
-- This migration:
-- 1. Checks for invalid indexes on agent_messages
-- 2. Drops and recreates them if invalid
-- 3. Updates statistics for query planner

-- Check if idx_agent_messages_inbox_btree exists and is valid
-- If invalid or missing, recreate it
DO $$
DECLARE
    idx_invalid boolean;
    idx_exists boolean;
BEGIN
    -- Check if index exists
    SELECT EXISTS (
        SELECT 1 FROM pg_indexes WHERE indexname = 'idx_agent_messages_inbox_btree'
    ) INTO idx_exists;

    IF idx_exists THEN
        -- Check if index is valid
        SELECT NOT indisvalid INTO idx_invalid
        FROM pg_index
        WHERE indexrelid = 'idx_agent_messages_inbox_btree'::regclass;

        IF idx_invalid THEN
            RAISE NOTICE 'Index idx_agent_messages_inbox_btree is INVALID, dropping and recreating...';
            DROP INDEX idx_agent_messages_inbox_btree;
            CREATE INDEX idx_agent_messages_inbox_btree
                ON agent_messages (to_agent_id, status, created_at ASC)
                WHERE status IN ('pending', 'delivered');
        ELSE
            RAISE NOTICE 'Index idx_agent_messages_inbox_btree is valid';
        END IF;
    ELSE
        RAISE NOTICE 'Index idx_agent_messages_inbox_btree does not exist, creating...';
        CREATE INDEX idx_agent_messages_inbox_btree
            ON agent_messages (to_agent_id, status, created_at ASC)
            WHERE status IN ('pending', 'delivered');
    END IF;
END $$;

-- Also ensure idx_agent_messages_outbox_btree is valid (for completeness)
DO $$
DECLARE
    idx_invalid boolean;
    idx_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM pg_indexes WHERE indexname = 'idx_agent_messages_outbox_btree'
    ) INTO idx_exists;

    IF idx_exists THEN
        SELECT NOT indisvalid INTO idx_invalid
        FROM pg_index
        WHERE indexrelid = 'idx_agent_messages_outbox_btree'::regclass;

        IF idx_invalid THEN
            RAISE NOTICE 'Index idx_agent_messages_outbox_btree is INVALID, dropping and recreating...';
            DROP INDEX idx_agent_messages_outbox_btree;
            CREATE INDEX idx_agent_messages_outbox_btree
                ON agent_messages (from_agent_id, created_at DESC);
        ELSE
            RAISE NOTICE 'Index idx_agent_messages_outbox_btree is valid';
        END IF;
    ELSE
        RAISE NOTICE 'Index idx_agent_messages_outbox_btree does not exist, creating...';
        CREATE INDEX idx_agent_messages_outbox_btree
            ON agent_messages (from_agent_id, created_at DESC);
    END IF;
END $$;

-- Update statistics for query planner to recognize the indexes
ANALYZE agent_messages;

-- Verify indexes are being used (this is informational)
-- The slow query should now use idx_agent_messages_inbox_btree
