-- Migration: Create trust infrastructure tables
-- Description: Adds trust revocation, attestation, and policy tables for complete trust infrastructure

-- ============================================
-- Trust Revocations Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_revocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_author VARCHAR(255),
    function_name VARCHAR(255),
    revocation_id VARCHAR(32) NOT NULL UNIQUE,
    reason VARCHAR(50) NOT NULL,
    reason_details TEXT,
    severity VARCHAR(20) NOT NULL DEFAULT 'high',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    revocation_type VARCHAR(30) NOT NULL DEFAULT 'full',
    impact_description TEXT,
    revoked_by UUID NOT NULL,
    revoked_by_type VARCHAR(20) NOT NULL DEFAULT 'admin',
    revoked_by_partner UUID,
    report_id UUID,
    original_trust_score FLOAT DEFAULT 0,
    original_trust_tier VARCHAR(20),
    original_is_verified BOOLEAN DEFAULT FALSE,
    evidence_urls JSONB DEFAULT '[]'::jsonb,
    documentation_url VARCHAR(500),
    appeal_status VARCHAR(20),
    appeal_submitted_at TIMESTAMP WITH TIME ZONE,
    appeal_reason TEXT,
    revoked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    lifted_at TIMESTAMP WITH TIME ZONE,
    lifted_by UUID,
    lift_reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    notified_users BOOLEAN DEFAULT FALSE,
    search_index_updated BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for fast lookup by function_id
CREATE INDEX IF NOT EXISTS idx_trust_revocations_function_id ON trust_revocations(function_id);
CREATE INDEX IF NOT EXISTS idx_trust_revocations_status ON trust_revocations(status);
CREATE INDEX IF NOT EXISTS idx_trust_revocations_revocation_id ON trust_revocations(revocation_id);
CREATE INDEX IF NOT EXISTS idx_trust_revocations_revoked_at ON trust_revocations(revoked_at DESC);

-- Composite index for finding active revocations
CREATE INDEX IF NOT EXISTS idx_trust_revocations_active ON trust_revocations(function_id, status) 
WHERE status = 'active';

-- ============================================
-- Trust Attestations Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_attestations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attestation_id VARCHAR(32) NOT NULL UNIQUE,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_version VARCHAR(50),
    function_author VARCHAR(255),
    function_name VARCHAR(255),
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'valid',
    title VARCHAR(255) NOT NULL,
    description TEXT,
    results JSONB DEFAULT '{}'::jsonb,
    attester_id UUID NOT NULL,
    attester_type VARCHAR(20) NOT NULL DEFAULT 'system',
    attester_name VARCHAR(255),
    attester_partner_id UUID,
    verification_level VARCHAR(30),
    proof_hash VARCHAR(64) NOT NULL,
    previous_hash VARCHAR(64),
    signature VARCHAR(512),
    public_key_id VARCHAR(100),
    source_data_hash VARCHAR(64),
    attested_at TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_by UUID,
    revoke_reason TEXT,
    revocation_id UUID,
    is_immutable BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for attestations
CREATE INDEX IF NOT EXISTS idx_trust_attestations_function_id ON trust_attestations(function_id);
CREATE INDEX IF NOT EXISTS idx_trust_attestations_attestation_id ON trust_attestations(attestation_id);
CREATE INDEX IF NOT EXISTS idx_trust_attestations_type ON trust_attestations(type);
CREATE INDEX IF NOT EXISTS idx_trust_attestations_status ON trust_attestations(status);
CREATE INDEX IF NOT EXISTS idx_trust_attestations_attested_at ON trust_attestations(attested_at DESC);

-- Composite index for valid attestations lookup
CREATE INDEX IF NOT EXISTS idx_trust_attestations_valid ON trust_attestations(function_id, type, status) 
WHERE status = 'valid';

-- ============================================
-- Trust Policies Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    version INTEGER DEFAULT 1,
    owner_id UUID NOT NULL,
    owner_type VARCHAR(20) NOT NULL DEFAULT 'user',
    owner_partner_id UUID,
    rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_action VARCHAR(20) NOT NULL DEFAULT 'deny',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_default BOOLEAN DEFAULT FALSE,
    use_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE,
    deprecated_at TIMESTAMP WITH TIME ZONE,
    deprecated_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for policies
CREATE INDEX IF NOT EXISTS idx_trust_policies_owner ON trust_policies(owner_id, owner_type);
CREATE INDEX IF NOT EXISTS idx_trust_policies_policy_id ON trust_policies(policy_id);
CREATE INDEX IF NOT EXISTS idx_trust_policies_status ON trust_policies(status);
CREATE INDEX IF NOT EXISTS idx_trust_policies_is_default ON trust_policies(owner_id, is_default) 
WHERE is_default = TRUE;

-- ============================================
-- Trust Policy Evaluations Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_policy_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    evaluation_id VARCHAR(32) NOT NULL UNIQUE,
    policy_id UUID NOT NULL REFERENCES trust_policies(id) ON DELETE CASCADE,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_author VARCHAR(255),
    function_name VARCHAR(255),
    trust_score FLOAT DEFAULT 0,
    trust_tier VARCHAR(20),
    is_verified BOOLEAN DEFAULT FALSE,
    verification_level VARCHAR(30),
    is_revoked BOOLEAN DEFAULT FALSE,
    revocation_status VARCHAR(20),
    result VARCHAR(20) NOT NULL,
    decision VARCHAR(50) NOT NULL,
    reason TEXT,
    rule_results JSONB DEFAULT '[]'::jsonb,
    evaluated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    evaluated_by UUID NOT NULL,
    evaluated_by_type VARCHAR(20) DEFAULT 'api',
    cache_valid_until TIMESTAMP WITH TIME ZONE,
    is_cached BOOLEAN DEFAULT FALSE,
    request_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for evaluations
CREATE INDEX IF NOT EXISTS idx_trust_evaluations_policy ON trust_policy_evaluations(policy_id);
CREATE INDEX IF NOT EXISTS idx_trust_evaluations_function ON trust_policy_evaluations(function_id);
CREATE INDEX IF NOT EXISTS idx_trust_evaluations_result ON trust_policy_evaluations(result);
CREATE INDEX IF NOT EXISTS idx_trust_evaluations_evaluated_at ON trust_policy_evaluations(evaluated_at DESC);

-- Composite index for cached evaluations lookup
CREATE INDEX IF NOT EXISTS idx_trust_evaluations_cached ON trust_policy_evaluations(policy_id, function_id, is_cached, cache_valid_until) 
WHERE is_cached = TRUE;

-- ============================================
-- Trigger for updated_at timestamps
-- ============================================
CREATE OR REPLACE FUNCTION update_trust_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DROP TRIGGER IF EXISTS update_trust_revocations_updated_at ON trust_revocations;
CREATE TRIGGER update_trust_revocations_updated_at 
    BEFORE UPDATE ON trust_revocations 
    FOR EACH ROW 
    EXECUTE FUNCTION update_trust_updated_at_column();

DROP TRIGGER IF EXISTS update_trust_attestations_updated_at ON trust_attestations;
CREATE TRIGGER update_trust_attestations_updated_at 
    BEFORE UPDATE ON trust_attestations 
    FOR EACH ROW 
    EXECUTE FUNCTION update_trust_updated_at_column();

DROP TRIGGER IF EXISTS update_trust_policies_updated_at ON trust_policies;
CREATE TRIGGER update_trust_policies_updated_at 
    BEFORE UPDATE ON trust_policies 
    FOR EACH ROW 
    EXECUTE FUNCTION update_trust_updated_at_column();

-- ============================================
-- Comments for documentation
-- ============================================
COMMENT ON TABLE trust_revocations IS 'Tracks trust revocations for functions - marks functions as untrusted/downgraded';
COMMENT ON TABLE trust_attestations IS 'Immutable cryptographic attestations of function trustworthiness';
COMMENT ON TABLE trust_policies IS 'User-defined policies for evaluating function trustworthiness';
COMMENT ON TABLE trust_policy_evaluations IS 'Cached results of policy evaluations against functions';
