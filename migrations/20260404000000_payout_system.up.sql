-- Payout system: Stripe Connect accounts, payout requests, and audit trail.
-- All monetary amounts stored as integer cents unless otherwise noted.

BEGIN;

-- ──────────────────────────────────────────────────────────────────────────────
-- stripe_connect_accounts — tracks Stripe Express/Standard connected accounts
-- per user so we can transfer funds and trigger payouts.
-- ──────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS stripe_connect_accounts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    stripe_account_id   TEXT NOT NULL UNIQUE,           -- Stripe connected account ID (acct_xxx)
    account_status      TEXT NOT NULL DEFAULT 'pending', -- pending, onboarding, active, restricted, disabled
    payouts_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    details_submitted   BOOLEAN NOT NULL DEFAULT FALSE,
    charges_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    country             TEXT,                           -- ISO 3166-1 alpha-2
    currency            TEXT NOT NULL DEFAULT 'usd',
    -- Masked bank info for display only (never store full account numbers)
    bank_last4          TEXT,
    bank_name           TEXT,
    onboarding_url      TEXT,                           -- Stripe Express onboarding link (refreshed on access)
    onboarding_url_expires_at TIMESTAMPTZ,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stripe_connect_accounts_user_id ON stripe_connect_accounts(user_id);
CREATE INDEX idx_stripe_connect_accounts_status ON stripe_connect_accounts(account_status);

-- ──────────────────────────────────────────────────────────────────────────────
-- payout_requests — user-initiated withdrawal requests.  Each request is
-- verified server-side against the user's available earnings, then a Stripe
-- Transfer is created to the connected account.  The actual payout to the
-- bank is handled separately by Stripe (payout.created / payout.paid events).
-- ──────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payout_requests (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    connect_account_id      UUID NOT NULL REFERENCES stripe_connect_accounts(id),
    -- Amount in USD cents
    amount_cents            INT NOT NULL CHECK (amount_cents > 0),
    currency                TEXT NOT NULL DEFAULT 'usd',
    status                  TEXT NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed, cancelled
    -- Stripe references
    stripe_transfer_id      TEXT,                           -- Stripe Transfer ID (tr_xxx)
    stripe_payout_id        TEXT,                           -- Stripe Payout ID  (po_xxx) — filled when payout.paid fires
    -- Idempotency
    idempotency_key         TEXT UNIQUE NOT NULL,           -- client-generated key to prevent duplicate requests
    -- Admin / audit
    failure_reason          TEXT,
    reviewed_by             UUID REFERENCES users(id),
    reviewed_at             TIMESTAMPTZ,
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payout_requests_user_id ON payout_requests(user_id);
CREATE INDEX idx_payout_requests_status ON payout_requests(status);
CREATE INDEX idx_payout_requests_connect_account ON payout_requests(connect_account_id);

-- ──────────────────────────────────────────────────────────────────────────────
-- payout_ledger — immutable audit trail.  Every movement of funds related to
-- payouts is recorded here (earnings credited, payout debited, reversal, etc.).
-- ──────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payout_ledger (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entry_type      TEXT NOT NULL,       -- earning_credit, payout_debit, payout_reversal, adjustment
    amount_cents    INT NOT NULL,        -- positive for credits, negative for debits
    currency        TEXT NOT NULL DEFAULT 'usd',
    reference_type  TEXT,                -- publisher_earning, payout_request, admin_adjustment
    reference_id    UUID,                -- FK to the source record
    balance_after_cents INT NOT NULL DEFAULT 0, -- running balance snapshot after this entry
    description     TEXT,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payout_ledger_user_id ON payout_ledger(user_id);
CREATE INDEX idx_payout_ledger_reference ON payout_ledger(reference_type, reference_id);
CREATE INDEX idx_payout_ledger_created_at ON payout_ledger(created_at);

-- ──────────────────────────────────────────────────────────────────────────────
-- Trigger: auto-update updated_at on stripe_connect_accounts and payout_requests
-- ──────────────────────────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION update_payout_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_stripe_connect_accounts_updated_at
    BEFORE UPDATE ON stripe_connect_accounts
    FOR EACH ROW EXECUTE FUNCTION update_payout_updated_at();

CREATE TRIGGER trg_payout_requests_updated_at
    BEFORE UPDATE ON payout_requests
    FOR EACH ROW EXECUTE FUNCTION update_payout_updated_at();

COMMIT;
