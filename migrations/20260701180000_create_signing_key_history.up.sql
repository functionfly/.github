-- Signing key history: stores all attestation signing keys (current and historical)
-- Enables key rotation by preserving old public keys for verification of past attestations.

CREATE TABLE IF NOT EXISTS signing_key_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id VARCHAR(64) NOT NULL UNIQUE,
    public_key_hex TEXT NOT NULL,
    algorithm VARCHAR(50) NOT NULL,
    backend VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    fingerprint VARCHAR(64) NOT NULL,
    activated_at TIMESTAMPTZ NOT NULL,
    deactivated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_signing_key_history_fingerprint ON signing_key_history(fingerprint);
CREATE INDEX IF NOT EXISTS idx_signing_key_history_is_active ON signing_key_history(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_signing_key_history_activated_at ON signing_key_history(activated_at DESC);
