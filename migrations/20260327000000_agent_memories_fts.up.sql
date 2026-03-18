-- Full-text search on agent_memories.content (works for TEXT or JSONB).
-- Best cheap 2026 option: Postgres FTS on existing instance, no extra infra.
CREATE INDEX IF NOT EXISTS idx_agent_memories_content_fts
ON agent_memories USING GIN (to_tsvector('english', coalesce(content::text, '')));
