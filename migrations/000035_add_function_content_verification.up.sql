-- Add function content verification tables

-- RegistryFunctionSignature table for digital signatures
CREATE TABLE registry_function_signatures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,
    algorithm VARCHAR(50) NOT NULL,
    key_id VARCHAR(255) NOT NULL,
    signature TEXT NOT NULL,
    signed_by VARCHAR(255) NOT NULL,
    signed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,
    verification_error TEXT,
    is_valid BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- RegistryFunctionMalwareScan table for malware scan results
CREATE TABLE registry_function_malware_scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,
    scan_engine VARCHAR(100) NOT NULL,
    scan_version VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    threats_found JSONB,
    risk_score FLOAT DEFAULT 0,
    scan_metadata JSONB,
    scanned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    scan_duration_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- RegistryFunctionApproval table for approval workflows
CREATE TABLE registry_function_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_version_id UUID NOT NULL REFERENCES registry_function_versions(id) ON DELETE CASCADE,
    approval_type VARCHAR(50) NOT NULL,
    requested_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL,
    trust_level VARCHAR(20) NOT NULL,
    review_deadline TIMESTAMP WITH TIME ZONE,
    assigned_to UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    comments TEXT,
    required_actions JSONB,
    completed_actions JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- RegistryFunctionApprovalComment table for approval comments
CREATE TABLE registry_function_approval_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_id UUID NOT NULL REFERENCES registry_function_approvals(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- RegistryFunctionVerificationStatus table for overall verification status
CREATE TABLE registry_function_verification_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_version_id UUID NOT NULL UNIQUE REFERENCES registry_function_versions(id) ON DELETE CASCADE,
    content_hash_verified BOOLEAN DEFAULT FALSE,
    signature_verified BOOLEAN DEFAULT FALSE,
    malware_scanned BOOLEAN DEFAULT FALSE,
    malware_status VARCHAR(50),
    malware_risk_score FLOAT DEFAULT 0,
    approval_required BOOLEAN DEFAULT FALSE,
    approval_status VARCHAR(50),
    approved_at TIMESTAMP WITH TIME ZONE,
    overall_status VARCHAR(50) NOT NULL,
    last_verified_at TIMESTAMP WITH TIME ZONE,
    next_verification_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_registry_function_signatures_function_version ON registry_function_signatures(function_version_id);
CREATE INDEX idx_registry_function_signatures_key_id ON registry_function_signatures(key_id);
CREATE INDEX idx_registry_function_signatures_is_valid ON registry_function_signatures(is_valid);

CREATE INDEX idx_registry_function_malware_scans_function_version ON registry_function_malware_scans(function_version_id);
CREATE INDEX idx_registry_function_malware_scans_status ON registry_function_malware_scans(status);
CREATE INDEX idx_registry_function_malware_scans_scanned_at ON registry_function_malware_scans(scanned_at);

CREATE INDEX idx_registry_function_approvals_function_version ON registry_function_approvals(function_version_id);
CREATE INDEX idx_registry_function_approvals_status ON registry_function_approvals(status);
CREATE INDEX idx_registry_function_approvals_assigned_to ON registry_function_approvals(assigned_to);
CREATE INDEX idx_registry_function_approvals_requested_by ON registry_function_approvals(requested_by);
CREATE INDEX idx_registry_function_approvals_review_deadline ON registry_function_approvals(review_deadline);

CREATE INDEX idx_registry_function_approval_comments_approval ON registry_function_approval_comments(approval_id);
CREATE INDEX idx_registry_function_approval_comments_user ON registry_function_approval_comments(user_id);

CREATE UNIQUE INDEX idx_registry_function_verification_status_version ON registry_function_verification_status(function_version_id);
CREATE INDEX idx_registry_function_verification_status_overall ON registry_function_verification_status(overall_status);
CREATE INDEX idx_registry_function_verification_status_last_verified ON registry_function_verification_status(last_verified_at);

-- Update triggers for updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_registry_function_signatures_updated_at
    BEFORE UPDATE ON registry_function_signatures
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_registry_function_malware_scans_updated_at
    BEFORE UPDATE ON registry_function_malware_scans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_registry_function_approvals_updated_at
    BEFORE UPDATE ON registry_function_approvals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_registry_function_approval_comments_updated_at
    BEFORE UPDATE ON registry_function_approval_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_registry_function_verification_status_updated_at
    BEFORE UPDATE ON registry_function_verification_status
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();