-- Create IP allowlist tables for tenant-based IP access control
-- Migration: 20260307000002_create_ip_allowlist_tables.up.sql

-- Main IP allowlist table
CREATE TABLE IF NOT EXISTS ip_allowlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    default_policy VARCHAR(20) DEFAULT 'deny',  -- 'allow' or 'deny'
    mfa_required_for_unknown_ip BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- IP allowlist entries table
CREATE TABLE IF NOT EXISTS ip_allowlist_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    allowlist_id UUID NOT NULL REFERENCES ip_allowlists(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,  -- 'ip' for single IP, 'cidr' for CIDR range
    value VARCHAR(100) NOT NULL,  -- '192.168.1.1' or '10.0.0.0/8'
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for efficient lookups
CREATE INDEX idx_ip_allowlist_tenant ON ip_allowlists(tenant_id);
CREATE INDEX idx_ip_allowlist_entries_allowlist ON ip_allowlist_entries(allowlist_id);

-- Add comments for documentation
COMMENT ON TABLE ip_allowlists IS 'Stores IP allowlist configurations for tenant-based access control';
COMMENT ON TABLE ip_allowlist_entries IS 'Contains individual IP addresses or CIDR ranges for each allowlist';
COMMENT ON COLUMN ip_allowlists.default_policy IS 'Default policy when client IP does not match any entry: allow or deny';
COMMENT ON COLUMN ip_allowlists.mfa_required_for_unknown_ip IS 'Whether MFA-verified users can bypass unknown IP restrictions';
COMMENT ON COLUMN ip_allowlist_entries.type IS 'Type of entry: ip (single address) or cidr (CIDR range)';
