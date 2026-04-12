-- Migration: Add payout approval workflow for large payouts
-- Implements multi-sig/approval workflow for security

-- Add approval fields to payout_requests table
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS requires_approval BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS approval_threshold_usd DECIMAL(10,2);
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS approval_notes TEXT;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS second_approval_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS second_approval_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS rejection_reason TEXT;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS rejected_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMP WITH TIME ZONE;

-- Create payout approval rules table
CREATE TABLE IF NOT EXISTS payout_approval_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    min_amount_usd DECIMAL(10,2) NOT NULL,
    max_amount_usd DECIMAL(10,2),
    required_approvals INT NOT NULL DEFAULT 1, -- Number of approvals needed (1 or 2)
    approver_roles VARCHAR[] NOT NULL DEFAULT '{admin}', -- Array of roles that can approve
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INT NOT NULL DEFAULT 0, -- Higher priority rules evaluated first
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create payout approval audit log
CREATE TABLE IF NOT EXISTS payout_approval_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payout_request_id UUID NOT NULL REFERENCES payout_requests(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL, -- 'submitted', 'approved', 'rejected', 'second_approved', 'cancelled'
    performed_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    performed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    previous_status VARCHAR(20),
    new_status VARCHAR(20),
    notes TEXT,
    ip_address INET,
    user_agent TEXT
);

-- Insert default approval rules
INSERT INTO payout_approval_rules (name, description, min_amount_usd, max_amount_usd, required_approvals, approver_roles, priority)
VALUES
    ('Large Payout Rule', 'Requires single admin approval for payouts >= $1000', 1000.00, 9999.99, 1, '{admin,billing_manager}', 10),
    ('Very Large Payout Rule', 'Requires two admin approvals for payouts >= $10000', 10000.00, NULL, 2, '{admin}', 20)
ON CONFLICT DO NOTHING;

-- Indexes
CREATE INDEX idx_payout_requests_requires_approval ON payout_requests(requires_approval, status) WHERE requires_approval = true;
CREATE INDEX idx_payout_requests_approved_by ON payout_requests(approved_by);
CREATE INDEX idx_payout_approval_audit_payout_id ON payout_approval_audit(payout_request_id);
CREATE INDEX idx_payout_approval_audit_performed_at ON payout_approval_audit(performed_at DESC);
CREATE INDEX idx_payout_approval_rules_active_priority ON payout_approval_rules(is_active, priority DESC);

-- Comments
COMMENT ON TABLE payout_approval_rules IS 'Rules governing approval requirements for payout requests';
COMMENT ON TABLE payout_approval_audit IS 'Audit trail for all payout approval actions';
