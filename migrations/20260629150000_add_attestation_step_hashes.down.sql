-- Remove per-step hash columns from trust_attestations
DROP INDEX IF EXISTS idx_trust_attestations_code_hash;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS output_hash;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS input_hash;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS code_hash;
