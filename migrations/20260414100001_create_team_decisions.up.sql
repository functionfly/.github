-- Migration: Create team_decisions table for Decision Recorder feature
-- Created: 2026-04-14
-- Purpose: Store team decisions with rationale, outcome, and approval tracking

CREATE TABLE IF NOT EXISTS team_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    rationale TEXT,
    outcome TEXT,
    alternatives TEXT DEFAULT '[]'::text,

    -- Audit
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Approval workflow
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    approved_by UUID,
    approved_at TIMESTAMP WITH TIME ZONE,

    -- Source tracking
    source_type VARCHAR(50) DEFAULT 'manual',
    source_id VARCHAR(255),

    -- Link to TeamMemory for AI-extracted content
    team_memory_id UUID,

    -- Tags for categorization
    tags TEXT DEFAULT '[]'::text,

    -- Importance for sorting/filtering
    importance_score DOUBLE PRECISION DEFAULT 0.5,

    -- Constraints
    CONSTRAINT team_decisions_status_check CHECK (status IN ('pending', 'approved', 'superseded', 'deprecated')),
    CONSTRAINT team_decisions_source_type_check CHECK (source_type IN ('manual', 'ai_extracted', 'imported'))
);

-- Indexes for common queries
CREATE INDEX idx_team_decisions_team_id ON team_decisions(team_id);
CREATE INDEX idx_team_decisions_status ON team_decisions(status);
CREATE INDEX idx_team_decisions_created_by ON team_decisions(created_by);
CREATE INDEX idx_team_decisions_created_at ON team_decisions(created_at DESC);

-- GIN index for text search on title, description, rationale, outcome
CREATE INDEX idx_team_decisions_search ON team_decisions USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '') || ' ' || COALESCE(rationale, '') || ' ' || COALESCE(outcome, '')));

-- Comments
COMMENT ON TABLE team_decisions IS 'Team decision recorder - stores decisions with rationale, outcome, and approval tracking';
COMMENT ON COLUMN team_decisions.rationale IS 'Why this decision was made';
COMMENT ON COLUMN team_decisions.outcome IS 'What specifically was decided';
COMMENT ON COLUMN team_decisions.alternatives IS 'JSON array of alternative options that were considered';
COMMENT ON COLUMN team_decisions.source_type IS 'How the decision was created: manual, ai_extracted, or imported';
COMMENT ON COLUMN team_decisions.team_memory_id IS 'Reference to team_memory table for AI-extracted decisions';
COMMENT ON COLUMN team_decisions.importance_score IS 'Importance rating 0-1, used for sorting and filtering';
