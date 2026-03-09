-- Admin roles and audit system
-- Add role column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50);

-- Create audit_events table for comprehensive logging
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES users(id),
    actor_email VARCHAR(255),
    tenant_id UUID REFERENCES tenants(id),
    action VARCHAR(100) NOT NULL, -- e.g., 'tenant.suspend', 'user.create', 'billing.plan.change'
    resource_type VARCHAR(50) NOT NULL, -- e.g., 'tenant', 'user', 'app', 'billing'
    resource_id UUID, -- ID of the affected resource
    request_id VARCHAR(255), -- For correlating with API requests
    before_state JSONB, -- State before the action (redacted for secrets)
    after_state JSONB, -- State after the action (redacted for secrets)
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    success BOOLEAN DEFAULT true
);

-- Basic billing tables (expanded in Phase 3)
CREATE TABLE IF NOT EXISTS pricing_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price_cents INTEGER NOT NULL, -- Price in cents (e.g., 9900 for $99)
    currency VARCHAR(3) DEFAULT 'USD',
    features JSONB, -- Feature flags/limits as JSON
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_audit_events_actor_user_id ON audit_events(actor_user_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_id ON audit_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_resource_type_id ON audit_events(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp ON audit_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_events_action ON audit_events(action);