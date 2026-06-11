-- DNA Tables: Remove foreign key constraints and idempotency index
-- Migration: 20260611000000_function_dna_constraints

-- Drop foreign key constraints
ALTER TABLE function_dna_mutations
    DROP CONSTRAINT IF EXISTS fk_dna_mutations_profile;

ALTER TABLE function_dna_execution_metrics
    DROP CONSTRAINT IF EXISTS fk_dna_metrics_profile;

ALTER TABLE function_dna_analysis_queue
    DROP CONSTRAINT IF EXISTS fk_dna_queue_profile;

-- Drop idempotency index
DROP INDEX IF EXISTS idx_dna_queue_function_pending;
