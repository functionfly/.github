-- Add tenant status for suspend/unsuspend functionality
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active';

-- Add plan field if not exists (for billing tiers)
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS plan VARCHAR(100);

-- Create index for status queries
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);