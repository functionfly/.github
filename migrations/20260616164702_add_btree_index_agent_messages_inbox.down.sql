-- Rollback: Drop the B-tree indexes added for inbox/outbox queries

DROP INDEX IF EXISTS idx_agent_messages_inbox_btree;
DROP INDEX IF EXISTS idx_agent_messages_outbox_btree;
