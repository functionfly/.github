-- Remove chain-of-custody columns
DROP INDEX IF EXISTS idx_delegator_fn;
DROP INDEX IF EXISTS idx_parent_attestation;
DROP INDEX IF EXISTS idx_delegation_chain;

ALTER TABLE trust_attestations DROP COLUMN IF EXISTS delegation_output_hash;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS delegation_input_hash;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS delegator_trust_score;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS delegator_agent_id;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS delegator_function_id;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS delegation_depth;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS parent_attestation_id;
ALTER TABLE trust_attestations DROP COLUMN IF EXISTS delegation_chain_id;
