-- Migration: Create team_memories table for Shared Brain / Team Memory Engine
-- Created: 2026-04-11
-- Features: Structured memory types, vector embeddings, client-side encryption toggle, RLS policies

-- Ensure pgvector extension is available (idempotent)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
    CREATE EXTENSION IF NOT EXISTS vector;
  END IF;
EXCEPTION
  WHEN OTHERS THEN NULL;
END
$$;

-- Team Memory table: Central "Shared Brain" for team knowledge
CREATE TABLE IF NOT EXISTS team_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,

    -- Memory categorization
    memory_type VARCHAR(50) NOT NULL, -- 'decision', 'preference', 'process', 'client_context'
    category VARCHAR(100), -- e.g., 'client:acme-corp', 'process:onboarding', 'decision:architecture'

    -- Content (structured JSON) - NULL when encrypted (encrypted_content used instead)
    content JSONB,
    summary TEXT, -- Human-readable summary (always plaintext for search/listing)

    -- Source tracking
    source_conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    source_event_id UUID,
    created_by UUID NOT NULL REFERENCES users(id),

    -- Vector embedding for semantic search (generated from plaintext before encryption)
    embedding vector(1536),

    -- Confidence & validation
    confidence_score FLOAT DEFAULT 0.9, -- 0.0-1.0
    is_validated BOOLEAN DEFAULT false,
    validated_by UUID REFERENCES users(id),
    validated_at TIMESTAMP WITH TIME ZONE,

    -- Retention & priority (MVP defaults: decisions=2yr, preferences=1yr, process=forever, client=until inactive)
    importance_score FLOAT DEFAULT 0.5, -- 0.0-1.0, decays with disuse
    ttl_days INT DEFAULT 0, -- 0 = never expire (override per type defaults in app layer)
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Access tracking
    access_count INT DEFAULT 0,
    last_accessed_at TIMESTAMP WITH TIME ZONE,

    -- Auto-update tracking
    auto_update_enabled BOOLEAN DEFAULT true,
    last_auto_updated_at TIMESTAMP WITH TIME ZONE,
    extraction_confidence FLOAT, -- confidence from AI extraction (if auto-created)

    -- Client-side encryption toggle (user-controlled zero-knowledge option)
    is_encrypted BOOLEAN DEFAULT false,
    encrypted_content BYTEA, -- AES-256-GCM ciphertext (NULL if not encrypted)
    encryption_iv BYTEA,     -- 12-byte IV for GCM
    encryption_tag BYTEA,    -- 16-byte auth tag

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Constraints
ALTER TABLE team_memories
    ADD CONSTRAINT valid_memory_type CHECK (memory_type IN ('decision', 'preference', 'process', 'client_context')),
    ADD CONSTRAINT valid_confidence CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    ADD CONSTRAINT valid_importance CHECK (importance_score >= 0.0 AND importance_score <= 1.0),
    ADD CONSTRAINT content_or_encrypted CHECK (
        (is_encrypted = false AND content IS NOT NULL) OR
        (is_encrypted = true AND encrypted_content IS NOT NULL)
    );

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_team_memories_team_id ON team_memories(team_id);
CREATE INDEX IF NOT EXISTS idx_team_memories_tenant_id ON team_memories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_team_memories_team_type ON team_memories(team_id, memory_type);
CREATE INDEX IF NOT EXISTS idx_team_memories_category ON team_memories(category) WHERE category IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_team_memories_expires ON team_memories(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_team_memories_created_by ON team_memories(created_by);
CREATE INDEX IF NOT EXISTS idx_team_memories_validated ON team_memories(is_validated, team_id);
CREATE INDEX IF NOT EXISTS idx_team_memories_auto_update ON team_memories(auto_update_enabled, last_auto_updated_at) WHERE auto_update_enabled = true;
CREATE INDEX IF NOT EXISTS idx_team_memories_encrypted ON team_memories(is_encrypted) WHERE is_encrypted = true;

-- HNSW vector index for semantic search (cosine similarity)
-- Only create if pgvector extension is available
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
    CREATE INDEX IF NOT EXISTS idx_team_memories_embedding_hnsw
    ON team_memories USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
  END IF;
EXCEPTION
  WHEN OTHERS THEN NULL;
END
$$;

-- GIN index for JSON content search (partial, only unencrypted content)
CREATE INDEX IF NOT EXISTS idx_team_memories_content_gin ON team_memories USING GIN(content) WHERE is_encrypted = false;

-- Composite index for common filtered queries
CREATE INDEX IF NOT EXISTS idx_team_memories_active ON team_memories(team_id, memory_type, is_validated, created_at DESC)
WHERE (expires_at IS NULL OR expires_at > NOW());

-- Row-Level Security (RLS) policies for multi-tenancy
ALTER TABLE team_memories ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy (uses existing tenant context pattern)
CREATE POLICY team_memories_tenant_isolation ON team_memories
    FOR ALL
    USING (tenant_id = current_tenant_id());

-- Team member access policy (allows access if user is team member)
CREATE POLICY team_memories_team_access ON team_memories
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM team_memberships tm
            WHERE tm.team_id = team_memories.team_id
            AND tm.user_id = current_user_id()
        )
    );

-- Audit trigger for tracking changes
CREATE OR REPLACE FUNCTION team_memories_audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit_log (table_name, record_id, operation, old_data, changed_at, changed_by)
        VALUES ('team_memories', OLD.id, 'DELETE', row_to_json(OLD), NOW(), current_user_id());
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO audit_log (table_name, record_id, operation, old_data, new_data, changed_at, changed_by)
        VALUES ('team_memories', NEW.id, 'UPDATE', row_to_json(OLD), row_to_json(NEW), NOW(), current_user_id());
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit_log (table_name, record_id, operation, new_data, changed_at, changed_by)
        VALUES ('team_memories', NEW.id, 'INSERT', row_to_json(NEW), NOW(), current_user_id());
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS team_memories_audit_trigger ON team_memories;
CREATE TRIGGER team_memories_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON team_memories
    FOR EACH ROW EXECUTE FUNCTION team_memories_audit_trigger_function();

-- Auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION update_team_memories_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS team_memories_updated_at_trigger ON team_memories;
CREATE TRIGGER team_memories_updated_at_trigger
    BEFORE UPDATE ON team_memories
    FOR EACH ROW EXECUTE FUNCTION update_team_memories_updated_at();

-- Helper function: Mark memory as accessed (increments counter, updates timestamp)
CREATE OR REPLACE FUNCTION mark_team_memory_accessed(memory_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE team_memories
    SET access_count = access_count + 1,
        last_accessed_at = NOW()
    WHERE id = memory_id;
END;
$$ LANGUAGE plpgsql;

-- Helper function: Apply importance decay (called by scheduled job)
CREATE OR REPLACE FUNCTION apply_team_memory_decay()
RETURNS INTEGER AS $$
DECLARE
    affected_count INTEGER;
BEGIN
    UPDATE team_memories
    SET importance_score = GREATEST(importance_score * 0.9, 0.1)
    WHERE last_accessed_at < NOW() - INTERVAL '90 days'
      AND importance_score > 0.1
      AND is_encrypted = false; -- Only decay unencrypted memories (encrypted handled client-side)
    
    GET DIAGNOSTICS affected_count = ROW_COUNT;
    RETURN affected_count;
END;
$$ LANGUAGE plpgsql;

-- View: Active team memories (non-expired, unencrypted summary for dashboard)
CREATE OR REPLACE VIEW active_team_memories AS
SELECT
    id,
    tenant_id,
    team_id,
    memory_type,
    category,
    summary,
    confidence_score,
    is_validated,
    importance_score,
    access_count,
    last_accessed_at,
    is_encrypted,
    created_at,
    updated_at
FROM team_memories
WHERE (expires_at IS NULL OR expires_at > NOW());

-- Memory Extractions table (pending validation queue)
CREATE TABLE IF NOT EXISTS memory_extractions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,

    -- Extracted data
    memory_type VARCHAR(50) NOT NULL,
    category VARCHAR(100),
    content JSONB NOT NULL,
    summary TEXT NOT NULL,
    confidence FLOAT NOT NULL,
    rationale TEXT,

    -- Review status
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'approved', 'rejected', 'auto_applied'
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,

    -- Auto-apply threshold (MVP default: 0.9)
    auto_apply_threshold FLOAT DEFAULT 0.9,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_memory_extractions_team ON memory_extractions(team_id);
CREATE INDEX IF NOT EXISTS idx_memory_extractions_conv ON memory_extractions(conversation_id);
CREATE INDEX IF NOT EXISTS idx_memory_extractions_status ON memory_extractions(status);
CREATE INDEX IF NOT EXISTS idx_memory_extractions_created ON memory_extractions(created_at) WHERE status = 'pending';

-- RLS policies for memory_extractions
ALTER TABLE memory_extractions ENABLE ROW LEVEL SECURITY;

CREATE POLICY memory_extractions_team_isolation ON memory_extractions
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM team_memberships tm
            WHERE tm.team_id = memory_extractions.team_id
            AND tm.user_id = current_user_id()
        )
    );

-- Comments
COMMENT ON TABLE team_memories IS 'Team Memory Engine - Shared Brain for team knowledge, decisions, preferences, processes, and client context. Supports client-side encryption toggle per memory.';
COMMENT ON COLUMN team_memories.is_encrypted IS 'User-controlled toggle: when true, content is encrypted client-side and server only stores ciphertext';
COMMENT ON COLUMN team_memories.embedding IS 'Vector embedding for semantic search, generated from plaintext before encryption';
COMMENT ON TABLE memory_extractions IS 'Queue for AI-extracted memories pending human validation (MVP: auto-apply >= 0.9 confidence)';
