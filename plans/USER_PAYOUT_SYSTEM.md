# User Payout System — Stripe Connect

## Context

The codebase has a mature payment inflow system (Stripe Checkout, subscriptions, wallet credits, agent billing) but **no payout outflow mechanism**. The `publisher_earnings` table tracks earnings with statuses `pending`, `available`, `withdrawn`, `withheld`, and has a `stripe_payout_id` column — but no code exists to actually pay users. This plan adds a production-ready payout system using **Stripe Connect Express** accounts and **Stripe Transfers**.

---

## Architecture Decision: Stripe Connect (not Plaid)

**Why Stripe Connect over Plaid:**
- Already using Stripe for all payments — single provider, unified ledger
- Stripe Connect Express handles KYC/identity verification, tax reporting (1099), and bank account management — Plaid would only link bank accounts and still require Stripe or another processor for the actual money movement
- Stripe Transfers are atomic, idempotent, and integrate with existing webhook infrastructure
- Express onboarding means FunctionFly never handles sensitive bank details (reduces PCI scope)
- Plaid would add a second provider dependency, second webhook system, and additional complexity with no advantage

**Connect account type:** Express — Stripe handles onboarding, identity verification, and payout scheduling. FunctionFly controls when and how much to transfer.

---

## Files to Create/Modify

### New Files

| File | Purpose |
|------|---------|
| `internal/payment/connect.go` | Stripe Connect: create account links, retrieve account status |
| `internal/payment/transfer.go` | Execute payouts via Stripe Transfers |
| `internal/storage/payout_repository.go` | Payout account CRUD, payout request ledger |
| `migrations/20260403000000_payout_system.up.sql` | `payout_accounts`, `payout_requests` tables |
| `migrations/20260403000000_payout_system.down.sql` | Drop tables |
| `internal/api/handlers/billing/payouts.go` | HTTP handlers for payout endpoints |
| `web/dashboard/src/pages/SettingsPage/components/PayoutSettingsTab.tsx` | Frontend: onboarding, balance, request payout |
| `web/dashboard/src/api/payouts.ts` | Frontend API client for payout endpoints |

### Modified Files

| File | Change |
|------|--------|
| `internal/api/handlers/billing/handler.go` | Add `payoutRepo` field to `Handler` struct |
| `internal/api/routes_auth.go` | Register payout routes |
| `internal/api/handlers/webhooks/stripe.go` | Handle `account.updated`, `payout.paid`, `payout.failed` webhook events |
| `internal/storage/interfaces.go` | Add payout repository methods to `Repository` interface |
| `internal/storage/repositories_billing.go` | Wire payout repository methods through `PostgresDB` |
| `web/dashboard/src/pages/SettingsPage/SettingsPage.tsx` | Add Payout Settings tab |
| `.env.example` | Add `STRIPE_CONNECT_CLIENT_ID` |

---

## Database Schema

### `payout_accounts` — Stripe Connect account per user

```sql
CREATE TABLE IF NOT EXISTS payout_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    stripe_connect_account_id TEXT NOT NULL UNIQUE,          -- acct_xxx
    account_status VARCHAR(20) NOT NULL DEFAULT 'pending',   -- pending, restricted, active, disabled
    payouts_enabled BOOLEAN NOT NULL DEFAULT false,
    charges_enabled BOOLEAN NOT NULL DEFAULT false,
    requirements_due TEXT[],                                   -- Stripe requirements currently_due
    country CHAR(2) NOT NULL DEFAULT 'US',
    currency CHAR(3) NOT NULL DEFAULT 'usd',
    business_type VARCHAR(20),                                -- individual, company
    onboarding_completed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payout_accounts_user_id ON payout_accounts(user_id);
CREATE INDEX idx_payout_accounts_stripe_id ON payout_accounts(stripe_connect_account_id);
```

### `payout_requests` — audit trail for every payout

```sql
CREATE TABLE IF NOT EXISTS payout_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    payout_account_id UUID NOT NULL REFERENCES payout_accounts(id),
    amount_cents INT NOT NULL CHECK (amount_cents > 0),
    currency CHAR(3) NOT NULL DEFAULT 'usd',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',            -- pending, processing, completed, failed, cancelled
    stripe_transfer_id TEXT UNIQUE,                           -- tr_xxx (set on success)
    stripe_payout_id TEXT,                                    -- po_xxx (from connected account's payout)
    failure_code TEXT,
    failure_message TEXT,
    idempotency_key TEXT NOT NULL UNIQUE,                     -- prevents duplicate payouts
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payout_requests_user_id ON payout_requests(user_id);
CREATE INDEX idx_payout_requests_status ON payout_requests(status);
CREATE INDEX idx_payout_requests_stripe_transfer ON payout_requests(stripe_transfer_id);
```

---

## Backend Implementation

### 1. `internal/payment/connect.go` — Stripe Connect Operations

- `CreateConnectedAccount(ctx, email, country, businessType) (accountID string, error)` — Creates a Stripe Connect Express account with `controller.requirements_collection: "eventually"`, returns `acct_xxx`
- `CreateAccountLink(ctx, accountID, refreshURL, returnURL) (onboardingURL string, error)` — Generates a Stripe Account Link for Express onboarding redirect
- `GetConnectedAccount(ctx, accountID) (*stripe.Account, error)` — Retrieves account status, requirements, payouts_enabled
- `CreateLoginLink(ctx, accountID) (dashboardURL string, error)` — Generates Stripe Express Dashboard login link for the connected user to manage their account

### 2. `internal/payment/transfer.go` — Payout Execution

- `CreateTransfer(ctx, amountCents int, currency, destinationAccountID, idempotencyKey string, metadata map[string]string) (transferID string, error)` — Creates a Stripe Transfer from platform account to connected account. Uses `stripe.StringKey("Idempotency-Key", idempotencyKey)` for deduplication.
- Validates: amount > 0, amount <= available balance, destination is an active connected account with payouts_enabled

### 3. `internal/storage/payout_repository.go` — Database Layer

```go
type PayoutAccount struct { ... }       // maps to payout_accounts table
type PayoutRequest struct { ... }       // maps to payout_requests table

type PayoutRepository struct { db *sql.DB }

// Payout Accounts
func (r *PayoutRepository) CreatePayoutAccount(ctx, account *PayoutAccount) error
func (r *PayoutRepository) GetPayoutAccountByUserID(ctx, userID) (*PayoutAccount, error)
func (r *PayoutRepository) GetPayoutAccountByStripeID(ctx, stripeAccountID) (*PayoutAccount, error)
func (r *PayoutRepository) UpdatePayoutAccountStatus(ctx, accountID, status string, payoutsEnabled bool, requirements []string) error

// Payout Requests
func (r *PayoutRepository) CreatePayoutRequest(ctx, req *PayoutRequest) error
func (r *PayoutRepository) GetPayoutRequestByID(ctx, id) (*PayoutRequest, error)
func (r *PayoutRequest) GetPayoutRequestByIdempotencyKey(ctx, key) (*PayoutRequest, error)
func (r *PayoutRepository) UpdatePayoutRequestStatus(ctx, id, status string, stripeTransferID, stripePayoutID *string) error
func (r *PayoutRepository) GetPayoutRequestsByUser(ctx, userID, limit, offset) ([]*PayoutRequest, int, error)
func (r *PayoutRepository) GetAvailableBalance(ctx, userID) (int, error)  // SUM(available publisher_earnings) - SUM(completed payouts)
```

### 4. `internal/api/handlers/billing/payouts.go` — HTTP Handlers

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | `/v1/billing/payouts/onboard` | `HandleStartPayoutOnboarding` | Creates Connect account (if needed) + returns Account Link URL |
| GET | `/v1/billing/payouts/status` | `HandleGetPayoutStatus` | Returns account status, available balance, payouts_enabled |
| POST | `/v1/billing/payouts/request` | `HandleRequestPayout` | Validates balance, creates Transfer, records payout_request |
| GET | `/v1/billing/payouts/history` | `HandleListPayouts` | Paginated payout request history |
| POST | `/v1/billing/payouts/dashboard` | `HandleGetExpressDashboard` | Returns Express Dashboard login link |

**`HandleStartPayoutOnboarding`** flow:
1. Auth check (requires logged-in user)
2. Check if `payout_accounts` row exists for user
3. If not, call `payment.CreateConnectedAccount()` → insert `payout_accounts` row
4. Call `payment.CreateAccountLink()` with `refresh_url` and `return_url`
5. Return `{ "onboarding_url": "https://connect.stripe.com/..." }`

**`HandleRequestPayout`** flow:
1. Auth check
2. Get payout account — must be `account_status = 'active'` and `payouts_enabled = true`
3. Call `payoutRepo.GetAvailableBalance(userID)` — calculates available earnings minus already-processed payouts
4. Validate `amount_cents > 0` and `amount_cents <= available_balance` and `amount_cents >= MIN_PAYOUT_CENTS` ($10.00)
5. Generate `idempotency_key = fmt.Sprintf("payout_%s_%d", userID, time.Now().UnixNano())`
6. Insert `payout_requests` with status `processing`
7. Call `payment.CreateTransfer()` with idempotency key
8. On success: update `payout_requests` status to `completed`, set `stripe_transfer_id`, update `publisher_earnings` status to `withdrawn`
9. On failure: update status to `failed`, set `failure_code`/`failure_message`
10. Return result

### 5. Webhook Handlers (modify `internal/api/handlers/webhooks/stripe.go`)

New event handlers in `handleEvent()`:

- **`account.updated`** — Updates `payout_accounts` status when Stripe updates the connected account (e.g., after onboarding completion, or if requirements change). Extracts `account_id`, checks `payouts_enabled`, `requirements.currently_due`, updates DB.
- **`transfer.paid`** — Confirms transfer succeeded, updates `payout_requests` status to `completed`.
- **`transfer.failed`** — Marks payout as failed, records failure reason.

---

## Frontend Implementation

### `web/dashboard/src/api/payouts.ts`

```typescript
export interface PayoutStatus {
  has_connected_account: boolean;
  account_status: string;        // pending, restricted, active, disabled
  payouts_enabled: boolean;
  available_balance_cents: number;
  available_balance_usd: number;
  total_withdrawn_cents: number;
  requirements_due: string[];
  country: string;
}

export interface PayoutRequest {
  id: string;
  amount_cents: number;
  amount_usd: number;
  status: string;
  requested_at: string;
  completed_at: string | null;
}

export interface PayoutHistoryResponse {
  payouts: PayoutRequest[];
  limit: number;
  offset: number;
  total: number;
}

export async function startPayoutOnboarding(): Promise<{ onboarding_url: string }>
export async function getPayoutStatus(): Promise<PayoutStatus>
export async function requestPayout(amountCents: number): Promise<PayoutRequest>
export async function listPayouts(limit?: number, offset?: number): Promise<PayoutHistoryResponse>
export async function getExpressDashboard(): Promise<{ dashboard_url: string }>
```

### `PayoutSettingsTab.tsx`

- **Not onboarded state:** "Set Up Payouts" button → calls `startPayoutOnboarding()` → redirects to Stripe Connect onboarding
- **Onboarding pending state:** Shows requirements due, "Continue Onboarding" button
- **Active state:**
  - Available balance display
  - Withdraw form (amount input, min $10, max = available balance)
  - "Manage Stripe Account" button → Express Dashboard link
  - Payout history table (date, amount, status)
- **Restricted state:** Shows requirements that need attention

---

## Security Controls

1. **Server-side balance validation** — `HandleRequestPayout` recalculates available balance from DB at request time (not trusting client)
2. **Idempotency keys** — Every transfer uses a unique idempotency key stored in `payout_requests.idempotency_key`, preventing duplicate payouts on retry
3. **Minimum payout threshold** — $10.00 USD minimum to avoid micro-transfer fees eating into payouts
4. **Account verification** — Only accounts with `payouts_enabled = true` and `account_status = 'active'` can request payouts
5. **Webhook signature verification** — Existing `STRIPE_WEBHOOK_SECRET` validation applies to new event types
6. **Amount caps** — Server validates `amount_cents <= available_balance` and `amount_cents > 0`
7. **Row-level locking** — `GetAvailableBalance` uses `SELECT ... FOR UPDATE` within a transaction to prevent race conditions
8. **No bank details stored** — Stripe handles all sensitive banking information via Express onboarding; FunctionFly never sees or stores bank account numbers

---

## Environment Variables

Add to `.env.example`:
```
STRIPE_CONNECT_CLIENT_ID=ca_xxx          # Stripe Connect platform client ID
MIN_PAYOUT_USD=10.00                     # Minimum payout amount (default $10)
```

---

## Implementation Order

1. **Migration** — Create `payout_accounts` and `payout_requests` tables
2. **Storage layer** — `payout_repository.go` with all CRUD + balance calculation
3. **Payment layer** — `connect.go` and `transfer.go`
4. **Wire into interfaces** — Add methods to `Repository` interface, wire through `PostgresDB`
5. **Handlers** — `payouts.go` with all 5 endpoints
6. **Routes** — Register in `routes_auth.go`
7. **Webhooks** — Add `account.updated`, `transfer.paid`, `transfer.failed` to `stripe.go`
8. **Frontend API** — `payouts.ts`
9. **Frontend UI** — `PayoutSettingsTab.tsx` + wire into Settings page
10. **Verify** — Run `go build`, lint, test

---

## Verification Steps

1. `go build -o bin/orchestrator-api ./cmd/orchestrator-api` — Confirm compilation
2. `golangci-lint run` — No lint errors
3. `go test ./internal/...` — Storage tests pass
4. `cd web/dashboard && npx tsc --noEmit` — TypeScript compiles
5. Manual test: start API + dashboard, verify payout onboarding flow returns Stripe URL
