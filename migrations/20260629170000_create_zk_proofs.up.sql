-- Zero-knowledge proofs for attestation verification.
-- Stores generated ZK proofs (existence, inclusion, range) for audit and replay.

CREATE TABLE IF NOT EXISTS zk_proofs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    proof_id VARCHAR(32) NOT NULL UNIQUE,
    type VARCHAR(30) NOT NULL,
    proof_data JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zk_proofs_type ON zk_proofs(type);
CREATE INDEX IF NOT EXISTS idx_zk_proofs_created ON zk_proofs(created_at DESC);
