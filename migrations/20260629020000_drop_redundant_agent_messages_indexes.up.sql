-- Migration: 20260629020000_drop_redundant_agent_messages_indexes
--
-- Problem: 10+ second INSERT latency on agent_messages due to 7 indexes
--          when only 3 are needed (pkey, inbox, outbox).
--
-- Every INSERT must update all indexes. Excessive indexes cause:
-- - Disk I/O amplification (writing to multiple index files)
-- - Index bloat over time
-- - Lock contention during heavy INSERT batches
--
-- Solution: Drop redundant indexes per 20260616220001_consolidate plan
--           that were never actually removed.
--
-- Keep:
--   1. agent_messages_pkey                    - primary key (required)
--   2. idx_agent_messages_inbox_btree        - inbox queries (to_agent_id, status, created_at)
--   3. idx_agent_messages_outbox_btree       - outbox queries (from_agent_id, created_at DESC)
--
-- Drop (redundant/overlapping):
--   1. idx_agent_messages_created_at_brin   - overlaps outbox_btree, unused
--   2. idx_agent_messages_to_agent_brin     - overlaps inbox_btree, unused
--   3. idx_agent_messages_pending_brin       - overlaps inbox_btree, unused
--   4. idx_agent_messages_id_status          - overlaps inbox_btree, unused

BEGIN;

-- Drop redundant BRIN indexes (overlap with inbox/outbox btree indexes)
DROP INDEX IF EXISTS idx_agent_messages_created_at_brin;
DROP INDEX IF EXISTS idx_agent_messages_to_agent_brin;
DROP INDEX IF EXISTS idx_agent_messages_pending_brin;

-- Drop redundant partial btree (overlaps with inbox_btree)
DROP INDEX IF EXISTS idx_agent_messages_id_status;

-- Verify the remaining indexes match expected set
DO $$
DECLARE
    expected_indexes TEXT[] := ARRAY[
        'agent_messages_pkey',
        'idx_agent_messages_inbox_btree',
        'idx_agent_messages_outbox_btree'
    ];
    actual_index TEXT;
    missing_index BOOLEAN;
BEGIN
    missing_index := FALSE;
    FOREACH actual_index IN ARRAY expected_indexes LOOP
        IF NOT EXISTS (
            SELECT 1 FROM pg_indexes
            WHERE tablename = 'agent_messages' AND indexname = actual_index
        ) THEN
            RAISE WARNING 'Expected index % is missing!', actual_index;
            missing_index := TRUE;
        END IF;
    END LOOP;

    IF NOT missing_index THEN
        RAISE NOTICE 'Index cleanup complete. Expected 3 indexes present.';
    ELSE
        RAISE WARNING 'Some expected indexes are missing. Manual intervention required.';
    END IF;
END $$;

-- Update statistics for query planner
ANALYZE agent_messages;

COMMIT;

-- Verify final state (informational)
-- SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'agent_messages' ORDER BY indexname;
