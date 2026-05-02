-- Migration: Drop team_memories and memory_extractions tables
-- Reverses: 20260411191000_team_memories

-- Drop triggers first (order matters for dependencies)
DROP TRIGGER IF EXISTS team_memories_updated_at_trigger ON team_memories;
DROP TRIGGER IF EXISTS team_memories_audit_trigger ON team_memories;
DROP TRIGGER IF EXISTS memory_extractions_audit_trigger ON memory_extractions;

-- Drop helper functions
DROP FUNCTION IF EXISTS mark_team_memory_accessed(UUID);
DROP FUNCTION IF EXISTS apply_team_memory_decay();
DROP FUNCTION IF EXISTS update_team_memories_updated_at();
DROP FUNCTION IF EXISTS team_memories_audit_trigger_function();

-- Drop view
DROP VIEW IF EXISTS active_team_memories;

-- Drop tables (memory_extractions depends on team_memories, so drop first)
DROP TABLE IF EXISTS memory_extractions;
DROP TABLE IF EXISTS team_memories;