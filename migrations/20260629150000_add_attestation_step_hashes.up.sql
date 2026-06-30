-- Add per-step hash columns to trust_attestations
-- These track the code, input, and output hashes for execution attestations,
-- enabling per-step cryptographic verification of what was attested.

ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS code_hash VARCHAR(64);
ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS input_hash VARCHAR(64);
ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS output_hash VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_trust_attestations_code_hash ON trust_attestations(code_hash) WHERE code_hash IS NOT NULL AND code_hash != '';
