-- Add missing tenant_id column to function_dna_analysis_queue
-- The column was in the original migration 20260502230000_function_dna.up.sql
-- but was missing when schema was applied manually

ALTER TABLE function_dna_analysis_queue ADD COLUMN IF NOT EXISTS tenant_id TEXT NOT NULL DEFAULT '';

-- Add missing columns that were also in original migration
ALTER TABLE function_dna_analysis_queue ADD COLUMN IF NOT EXISTS function_type TEXT NOT NULL DEFAULT 'full';
ALTER TABLE function_dna_analysis_queue ADD COLUMN IF NOT EXISTS last_error TEXT;
ALTER TABLE function_dna_analysis_queue ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();