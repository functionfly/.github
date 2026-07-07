-- Remove cold_start_ms column
ALTER TABLE function_dna_execution_metrics DROP COLUMN IF EXISTS cold_start_ms;
ALTER TABLE function_dna_execution_metrics_2026_05 DROP COLUMN IF EXISTS cold_start_ms;
ALTER TABLE function_dna_execution_metrics_2026_06 DROP COLUMN IF EXISTS cold_start_ms;
ALTER TABLE function_dna_execution_metrics_2026_07 DROP COLUMN IF EXISTS cold_start_ms;
ALTER TABLE function_dna_execution_metrics_2026_08 DROP COLUMN IF EXISTS cold_start_ms;

-- Drop index
DROP INDEX IF EXISTS idx_dna_metrics_cold_start_ms;
