-- Merkle audit trail for attestation log integrity.
-- Provides append-only, tamper-evident proof that attestations have not been altered.

CREATE TABLE IF NOT EXISTS merkle_tree_heads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tree_size BIGINT NOT NULL,
    root_hash VARCHAR(64) NOT NULL,
    previous_hash VARCHAR(64),
    timestamp TIMESTAMPTZ NOT NULL,
    signature VARCHAR(512),
    public_key_id VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_merkle_heads_size ON merkle_tree_heads(tree_size);
CREATE INDEX IF NOT EXISTS idx_merkle_heads_ts ON merkle_tree_heads(timestamp DESC);

CREATE TABLE IF NOT EXISTS merkle_nodes (
    id BIGSERIAL PRIMARY KEY,
    level INT NOT NULL,
    index BIGINT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    leaf_id VARCHAR(32),
    UNIQUE (level, index)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_merkle_node_hash ON merkle_nodes(hash);
CREATE INDEX IF NOT EXISTS idx_merkle_node_leaf ON merkle_nodes(leaf_id) WHERE leaf_id IS NOT NULL;
