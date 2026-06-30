-- Add chain-of-custody columns for multi-agent delegation tracking.
-- These fields record which agent delegated to which function, at what depth,
-- forming a verifiable chain of custody.

ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS delegation_chain_id VARCHAR(32);
ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS parent_attestation_id VARCHAR(32);
ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS delegation_depth INT DEFAULT 0;
ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS delegator_function_id UUID;
ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS delegator_agent_id VARCHAR(255);
ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS delegator_trust_score DOUBLE PRECISION;
ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS delegation_input_hash VARCHAR(64);
ALTER TABLE trust_attestations ADD COLUMN IF NOT EXISTS delegation_output_hash VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_delegation_chain ON trust_attestations(delegation_chain_id) WHERE delegation_chain_id IS NOT NULL AND delegation_chain_id != '';
CREATE INDEX IF NOT EXISTS idx_parent_attestation ON trust_attestations(parent_attestation_id) WHERE parent_attestation_id IS NOT NULL AND parent_attestation_id != '';
CREATE INDEX IF NOT EXISTS idx_delegator_fn ON trust_attestations(delegator_function_id) WHERE delegator_function_id IS NOT NULL;
