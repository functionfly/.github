-- Remove tenant_id and other columns added to fix schema mismatch
ALTER TABLE function_dna_analysis_queue DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE function_dna_analysis_queue DROP COLUMN IF EXISTS function_type;
ALTER TABLE function_dna_analysis_queue DROP COLUMN IF EXISTS last_error;
ALTER TABLE function_dna_analysis_queue DROP COLUMN IF EXISTS created_at;