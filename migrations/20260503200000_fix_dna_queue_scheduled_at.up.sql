-- Fix function_dna_analysis_queue schema mismatch
-- The code expects `scheduled_at` but the table was created with `queued_at`

ALTER TABLE function_dna_analysis_queue
ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ DEFAULT NOW();

UPDATE function_dna_analysis_queue
SET scheduled_at = queued_at
WHERE scheduled_at IS NULL;
