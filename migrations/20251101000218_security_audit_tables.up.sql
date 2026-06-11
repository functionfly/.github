-- Create security_scans table
CREATE TABLE IF NOT EXISTS security_scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    user_id UUID,
    scan_type VARCHAR(50) NOT NULL CHECK (scan_type IN ('penetration_test', 'vulnerability_scan', 'compliance_check')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    target TEXT NOT NULL,
    config JSONB,
    summary JSONB,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create vulnerabilities table
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    cvss_score DECIMAL(3,1),
    cve VARCHAR(50),
    category VARCHAR(50) NOT NULL CHECK (category IN ('injection', 'auth', 'crypto', 'config', 'network')),
    component VARCHAR(255) NOT NULL,
    location TEXT,
    status VARCHAR(20) NOT NULL CHECK (status IN ('open', 'fixed', 'accepted', 'false_positive')) DEFAULT 'open',
    remediation TEXT,
    reference_urls TEXT[],
    metadata JSONB,
    discovered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_security_scans_tenant_id ON security_scans(tenant_id);
CREATE INDEX IF NOT EXISTS idx_security_scans_user_id ON security_scans(user_id);
CREATE INDEX IF NOT EXISTS idx_security_scans_status ON security_scans(status);
CREATE INDEX IF NOT EXISTS idx_security_scans_started_at ON security_scans(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_scan_id ON vulnerabilities(scan_id);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_severity ON vulnerabilities(severity);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_status ON vulnerabilities(status);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_category ON vulnerabilities(category);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_discovered_at ON vulnerabilities(discovered_at DESC);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_security_scans_updated_at ON security_scans;
CREATE TRIGGER update_security_scans_updated_at BEFORE UPDATE ON security_scans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_vulnerabilities_updated_at ON vulnerabilities;
CREATE TRIGGER update_vulnerabilities_updated_at BEFORE UPDATE ON vulnerabilities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();