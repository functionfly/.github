-- Payout system enhancements: schedule preferences, velocity tracking,
-- fee configuration, and admin management support.

BEGIN;

-- ──────────────────────────────────────────────────────────────────────────────
-- payout_schedule_preferences — per-user auto-payout schedule settings
-- ──────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payout_schedule_preferences (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    schedule_enabled    BOOLEAN NOT NULL DEFAULT FALSE,
    frequency           TEXT NOT NULL DEFAULT 'weekly',  -- weekly, biweekly, monthly
    minimum_amount_cents INT NOT NULL DEFAULT 5000,      -- $50.00 minimum to trigger auto-payout
    day_of_week         INT,                             -- 0=Sunday, 6=Saturday (for weekly/biweekly)
    day_of_month        INT,                             -- 1-28 (for monthly)
    currency            TEXT NOT NULL DEFAULT 'usd',
    last_auto_payout_at TIMESTAMPTZ,
    next_scheduled_at   TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payout_schedule_preferences_next ON payout_schedule_preferences(next_scheduled_at)
    WHERE schedule_enabled = TRUE;

-- ──────────────────────────────────────────────────────────────────────────────
-- payout_velocity_tracking — fraud prevention: track payout frequency/amounts
-- ──────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payout_velocity_tracking (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    window_start        TIMESTAMPTZ NOT NULL,
    window_end          TIMESTAMPTZ NOT NULL,
    payout_count        INT NOT NULL DEFAULT 0,
    total_amount_cents  INT NOT NULL DEFAULT 0,
    last_payout_at      TIMESTAMPTZ,
    flags               JSONB DEFAULT '{}',  -- e.g. {"rapid_succession": true, "amount_spike": true}
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payout_velocity_user_window ON payout_velocity_tracking(user_id, window_start, window_end);

-- ──────────────────────────────────────────────────────────────────────────────
-- payout_fee_config — platform fee configuration for payouts
-- ──────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payout_fee_config (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                TEXT NOT NULL UNIQUE,
    description         TEXT,
    fee_type            TEXT NOT NULL DEFAULT 'percentage',  -- percentage, flat, tiered
    fee_percent         DECIMAL(6,4) NOT NULL DEFAULT 0,    -- e.g. 2.5000 = 2.5%
    flat_fee_cents      INT NOT NULL DEFAULT 0,             -- flat fee in cents
    minimum_fee_cents   INT NOT NULL DEFAULT 0,             -- minimum fee floor
    maximum_fee_cents   INT,                                -- fee cap (NULL = no cap)
    is_active           BOOLEAN NOT NULL DEFAULT true,
    applies_to          TEXT NOT NULL DEFAULT 'all',        -- all, international, domestic
    effective_from       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_until     TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default fee configuration
INSERT INTO payout_fee_config (name, description, fee_type, fee_percent, flat_fee_cents, minimum_fee_cents, maximum_fee_cents)
VALUES
    ('standard', 'Standard payout processing fee', 'percentage', 0.0000, 0, 0, NULL),
    ('instant', 'Instant payout surcharge', 'percentage', 1.5000, 0, 50, 5000)
ON CONFLICT (name) DO NOTHING;

-- ──────────────────────────────────────────────────────────────────────────────
-- payout_fee_deductions — audit trail for fees deducted from each payout
-- ──────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payout_fee_deductions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payout_request_id   UUID NOT NULL REFERENCES payout_requests(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    gross_amount_cents  INT NOT NULL,         -- original requested amount
    fee_amount_cents    INT NOT NULL,         -- fee deducted
    net_amount_cents    INT NOT NULL,         -- amount actually transferred
    fee_config_id       UUID REFERENCES payout_fee_config(id),
    fee_type            TEXT NOT NULL,         -- percentage, flat, tiered
    fee_rate            DECIMAL(10,6),        -- rate applied (for percentage)
    currency            TEXT NOT NULL DEFAULT 'usd',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payout_fee_deductions_payout ON payout_fee_deductions(payout_request_id);
CREATE INDEX idx_payout_fee_deductions_user ON payout_fee_deductions(user_id);

-- ──────────────────────────────────────────────────────────────────────────────
-- Add approval workflow columns to payout_requests if missing
-- (migration 20260412000003 may have added some; use ADD IF NOT EXISTS)
-- ──────────────────────────────────────────────────────────────────────────────
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS requires_approval BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS approval_threshold_usd DECIMAL(14,4);
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id);
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS approval_notes TEXT;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS second_approval_by UUID REFERENCES users(id);
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS second_approval_at TIMESTAMPTZ;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS rejected_by UUID REFERENCES users(id);
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS rejection_reason TEXT;
ALTER TABLE payout_requests ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';

-- ──────────────────────────────────────────────────────────────────────────────
-- Trigger: auto-update updated_at on new tables
-- ──────────────────────────────────────────────────────────────────────────────
CREATE TRIGGER trg_payout_schedule_preferences_updated_at
    BEFORE UPDATE ON payout_schedule_preferences
    FOR EACH ROW EXECUTE FUNCTION update_payout_updated_at();

CREATE TRIGGER trg_payout_velocity_tracking_updated_at
    BEFORE UPDATE ON payout_velocity_tracking
    FOR EACH ROW EXECUTE FUNCTION update_payout_updated_at();

CREATE TRIGGER trg_payout_fee_config_updated_at
    BEFORE UPDATE ON payout_fee_config
    FOR EACH ROW EXECUTE FUNCTION update_payout_updated_at();

-- ──────────────────────────────────────────────────────────────────────────────
-- Indexes for common admin queries
-- ──────────────────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_payout_requests_created_at ON payout_requests(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payout_requests_user_status ON payout_requests(user_id, status);
CREATE INDEX IF NOT EXISTS idx_payout_requests_amount ON payout_requests(amount_cents);

COMMENT ON TABLE payout_schedule_preferences IS 'Per-user auto-payout schedule configuration';
COMMENT ON TABLE payout_velocity_tracking IS 'Tracks payout frequency and amounts for fraud prevention';
COMMENT ON TABLE payout_fee_config IS 'Platform fee configuration for different payout types';
COMMENT ON TABLE payout_fee_deductions IS 'Immutable audit trail of fees deducted per payout';

COMMIT;
