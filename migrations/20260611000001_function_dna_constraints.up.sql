-- DNA Tables: Add foreign key constraints and idempotency index
-- Migration: 20260611000000_function_dna_constraints

-- ============================================================================
-- Add unique index for analysis queue idempotency
-- This enables ON CONFLICT DO NOTHING in EnqueueAnalysis
-- ============================================================================
CREATE UNIQUE INDEX IF NOT EXISTS idx_dna_queue_function_pending
    ON function_dna_analysis_queue(function_id)
    WHERE status IN ('pending', 'processing');

-- ============================================================================
-- Add foreign key constraints to DNA tables
-- These ensure referential integrity and enable CASCADE deletes
-- ============================================================================

-- FK: mutations reference profiles (function_id, function_type)
ALTER TABLE function_dna_mutations
    ADD CONSTRAINT fk_dna_mutations_profile
    FOREIGN KEY (function_id, function_type)
    REFERENCES function_dna_profiles(function_id, function_type)
    ON DELETE CASCADE;

-- FK: execution metrics reference profiles
-- Note: partitioned tables inherit FK constraints from parent
ALTER TABLE function_dna_execution_metrics
    ADD CONSTRAINT fk_dna_metrics_profile
    FOREIGN KEY (function_id)
    REFERENCES function_dna_profiles(function_id)
    ON DELETE CASCADE;

-- FK: analysis queue reference profiles
ALTER TABLE function_dna_analysis_queue
    ADD CONSTRAINT fk_dna_queue_profile
    FOREIGN KEY (function_id)
    REFERENCES function_dna_profiles(function_id)
    ON DELETE CASCADE;
