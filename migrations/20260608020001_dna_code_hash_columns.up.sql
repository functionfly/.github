-- DNA Mutation Code Hash Columns
-- Migration: 20260608020001_dna_code_hash_columns
-- Purpose: Store only code hashes instead of full code for security
-- The original_code and mutated_code columns remain nullable for backward
-- compatibility but should not be populated in normal operation.

ALTER TABLE function_dna_mutations
ADD COLUMN IF NOT EXISTS code_hash_algo TEXT NOT NULL DEFAULT 'sha256';

ALTER TABLE function_dna_mutations
ADD COLUMN IF NOT EXISTS original_code_hash TEXT;

ALTER TABLE function_dna_mutations
ADD COLUMN IF NOT EXISTS mutated_code_hash TEXT;

ALTER TABLE function_dna_mutations
ADD COLUMN IF NOT EXISTS code_size_bytes INT;

ALTER TABLE function_dna_mutations
ADD COLUMN IF NOT EXISTS line_count INT;

-- Index for looking up mutations by code hash (useful for deduplication)
CREATE INDEX IF NOT EXISTS idx_dna_mutations_original_hash
ON function_dna_mutations(original_code_hash)
WHERE original_code_hash IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_dna_mutations_mutated_hash
ON function_dna_mutations(mutated_code_hash)
WHERE mutated_code_hash IS NOT NULL;