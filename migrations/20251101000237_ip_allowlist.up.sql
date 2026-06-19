-- IP Allowlist table for admin dashboard security
-- Supports CIDR notation for IPv4 and IPv6 ranges
CREATE TABLE IF NOT EXISTS ip_allowlist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    cidr VARCHAR(45) NOT NULL, -- Supports IPv4 and IPv6 CIDR notation
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_whitelist BOOLEAN DEFAULT TRUE, -- TRUE=allow, FALSE=block
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient lookups
CREATE INDEX idx_ip_allowlist_cidr ON ip_allowlist(cidr);
CREATE INDEX idx_ip_allowlist_active ON ip_allowlist(is_active);
CREATE INDEX idx_ip_allowlist_created_by ON ip_allowlist(created_by);

-- Comment for documentation
COMMENT ON TABLE ip_allowlist IS 'Admin dashboard IP allowlist supporting CIDR notation for IPv4/IPv6';
COMMENT ON COLUMN ip_allowlist.is_whitelist IS 'When TRUE, matching IPs are allowed; when FALSE, matching IPs are blocked';
