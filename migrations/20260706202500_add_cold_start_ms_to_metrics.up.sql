-- Add cold_start_ms column to function_dna_execution_metrics for tracking cold start duration
ALTER TABLE function_dna_execution_metrics ADD COLUMN IF NOT EXISTS cold_start_ms INT DEFAULT 0;

-- Update partitions
ALTER TABLE function_dna_execution_metrics_2026_05 ADD COLUMN IF NOT EXISTS cold_start_ms INT DEFAULT 0;
ALTER TABLE function_dna_execution_metrics_2026_06 ADD COLUMN IF NOT EXISTS cold_start_ms INT DEFAULT 0;
ALTER TABLE function_dna_execution_metrics_2026_07 ADD COLUMN IF NOT EXISTS cold_start_ms INT DEFAULT 0;
ALTER TABLE function_dna_execution_metrics_2026_08 ADD COLUMN IF NOT EXISTS cold_start_ms INT DEFAULT 0;

-- Index for efficient cold start duration queries
CREATE INDEX IF NOT EXISTS idx_dna_metrics_cold_start_ms ON function_dna_execution_metrics(cold_start_ms) WHERE cold_start = true;
