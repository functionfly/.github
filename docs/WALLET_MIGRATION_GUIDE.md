# Unified Wallet System Migration Guide

This guide documents the migration from two separate wallet systems (user_wallets for registry fees and agent_billing_controls for execution credits) to a unified wallet system.

## Overview

### Before (Two Separate Systems)

1. **User Wallet (Registry Platform Fees)** - `user_wallets` table
   - Purpose: Pay registry fees (publish, version updates)
   - Location: `internal/storage/registry/platform_fees.go`
   - Top-up: Stripe Checkout with `purpose: "registry_wallet_credit"`

2. **Agent Wallet (Execution Credits)** - `agent_billing_controls` table
   - Purpose: Pay for agent function execution
   - Location: `internal/agent/billing/controls.go`
   - Top-up: Stripe Checkout with `purpose: "agent_execution_credits"`

### After (Unified System)

**Unified Wallet** - `wallets` table with polymorphic ownership
- Single table supporting both users and agents
- Unified transaction ledger: `wallet_transactions`
- Backward compatible interfaces
- Feature flag controlled rollout

## Migration Steps

### Step 1: Apply Schema Migration

```bash
# Apply the unified wallet schema
# This creates the new tables without affecting existing ones
psql $DATABASE_URL -f migrations/000200_unified_wallet_system.up.sql
```

Or via golang-migrate:
```bash
migrate -path migrations -database "$DATABASE_URL" up
```

### Step 2: Run Data Migration

```bash
# Dry run first to preview changes
go run cmd/migrate-wallets/migrate.go --dry-run --verbose

# Apply the migration
go run cmd/migrate-wallets/migrate.go

# With custom batch size
go run cmd/migrate-wallets/migrate.go --batch-size=500
```

### Step 3: Update Application Code

Update your server initialization to wire up the unified wallet service:

```go
// In your server initialization (e.g., cmd/orchestrator-api/main.go)

import (
    "github.com/functionfly/functionfly/internal/wallet"
    webhookshandlers "github.com/functionfly/functionfly/internal/api/handlers/webhooks"
)

// 1. Create wallet repository and service
walletRepo := wallet.NewRepository(db)
walletService := wallet.NewService(walletRepo, redisClient)

// 2. Create compatibility wrappers for gradual migration
platformFeesWrapper := wallet.NewPlatformFeeRepositoryWrapper(walletService)
billingCtrlWrapper := wallet.NewBillingControllerWrapper(walletService)

// 3. Update webhook handler with feature flags
stripeWebhookHandler := webhookshandlers.NewStripeWebhookHandlerV2(
    webhookshandlers.StripeWebhookHandlerV2Config{
        FinancialTxRepo: financialTxRepo,
        BillingCtrl:     billingCtrl,           // Legacy - can be removed after full migration
        UserRepo:        userRepo,
        PlatformFees:    platformFees,          // Legacy - can be removed after full migration
        SFAddons:        sfAddons,
        WalletService:   walletService,         // NEW
        NotificationSvc: notificationSvc,
        WebhookSecret:   os.Getenv("STRIPE_WEBHOOK_SECRET"),
        UseUnifiedWalletForUsers:  false,     // Start with false, enable after testing
        UseUnifiedWalletForAgents: false,     // Start with false, enable after testing
    },
)
```

### Step 4: Gradual Rollout

Enable the unified wallet gradually:

```go
// After testing, enable for users
stripeWebhookHandler.EnableUnifiedWalletForUsers()

// After more testing, enable for agents
stripeWebhookHandler.EnableUnifiedWalletForAgents()

// To disable (emergency rollback)
stripeWebhookHandler.DisableUnifiedWallet()
```

## Files Changed/Created

### New Files

1. **Schema Migration**
   - `migrations/000200_unified_wallet_system.up.sql` - Creates new tables
   - `migrations/000200_unified_wallet_system.down.sql` - Rollback

2. **Wallet Package**
   - `internal/wallet/models.go` - Data models
   - `internal/wallet/repository.go` - Database operations
   - `internal/wallet/service.go` - Business logic
   - `internal/wallet/compat_adapter.go` - Backward compatibility

3. **Data Migration**
   - `cmd/migrate-wallets/migrate.go` - Migration script

4. **Updated Handlers**
   - `internal/api/handlers/webhooks/stripe_wallet.go` - Webhook handler v2

### Modified Files (Backward Compatible)

The following files can continue using the old interfaces during migration:

- `internal/api/handlers/webhooks/stripe.go` - Uses wrapper interfaces
- `internal/storage/registry/platform_fees.go` - Legacy but functional
- `internal/agent/billing/controls.go` - Legacy but functional

## Schema Comparison

### Old: user_wallets

| Column | Type | Notes |
|--------|------|-------|
| user_id | UUID (PK) | Wallet owner |
| balance_usd | DECIMAL(14,4) | Current balance |
| lifetime_earnings_usd | DECIMAL(14,4) | Total credits received |
| lifetime_fees_usd | DECIMAL(14,4) | Total fees paid |
| created_at | TIMESTAMPTZ | Creation time |
| updated_at | TIMESTAMPTZ | Last update |

### Old: agent_billing_controls

| Column | Type | Notes |
|--------|------|-------|
| id | UUID (PK) | Unique ID |
| agent_id | TEXT (unique) | Wallet owner |
| spend_cap_monthly_usd | DECIMAL(10,2) | Monthly limit |
| spend_cap_daily_usd | DECIMAL(10,2) | Daily limit |
| credit_balance_usd | DECIMAL(10,2) | Current balance |
| billing_mode | TEXT | per_agent/per_tenant/per_team |
| team_id | UUID | Optional team grouping |
| alert_thresholds | DECIMAL[] | Low balance alert thresholds |
| created_at | TIMESTAMPTZ | Creation time |
| updated_at | TIMESTAMPTZ | Last update |

### New: wallets (Unified)

| Column | Type | Notes |
|--------|------|-------|
| id | UUID (PK) | Unique ID |
| owner_type | TEXT | 'user' or 'agent' |
| owner_id | TEXT | user_id or agent_id |
| user_id | UUID (FK) | Set when owner_type='user' |
| agent_id | TEXT (FK) | Set when owner_type='agent' |
| wallet_type | TEXT | 'unified', 'registry', or 'execution' |
| balance_usd | DECIMAL(14,4) | Current balance (from both old tables) |
| lifetime_earnings_usd | DECIMAL(14,4) | From user_wallets |
| lifetime_spent_usd | DECIMAL(14,4) | From user_wallets (renamed) |
| spend_cap_monthly_usd | DECIMAL(10,2) | From agent_billing_controls |
| spend_cap_daily_usd | DECIMAL(10,2) | From agent_billing_controls |
| alert_thresholds | DECIMAL[] | From agent_billing_controls |
| billing_mode | TEXT | From agent_billing_controls |
| team_id | UUID | From agent_billing_controls |
| status | TEXT | 'active', 'suspended', 'closed' |
| closed_at | TIMESTAMPTZ | When wallet was closed |
| closure_reason | TEXT | Why wallet was closed |
| created_at | TIMESTAMPTZ | Creation time |
| updated_at | TIMESTAMPTZ | Last update |

### New: wallet_transactions (Unified Ledger)

| Column | Type | Notes |
|--------|------|-------|
| id | UUID (PK) | Unique ID |
| wallet_id | UUID (FK) | Links to wallets |
| transaction_type | TEXT | credit, debit, fee_payment, execution_charge, etc. |
| amount_usd | DECIMAL(14,4) | Transaction amount |
| balance_before_usd | DECIMAL(14,4) | Balance before transaction |
| balance_after_usd | DECIMAL(14,4) | Balance after transaction |
| status | TEXT | pending, completed, failed, reversed |
| reference | TEXT | External reference (Stripe payment ID) |
| parent_transaction_id | UUID (FK) | For refunds/reversals |
| triggered_by_type | TEXT | user, agent, system, admin, webhook |
| triggered_by_id | TEXT | Who triggered it |
| execution_id | UUID | For execution charges |
| function_id | UUID | For execution charges |
| fee_type | TEXT | publish, version_update, commission |
| metadata | JSONB | Additional data |
| idempotency_key | TEXT | Prevent duplicates |
| created_at | TIMESTAMPTZ | Creation time |
| completed_at | TIMESTAMPTZ | When transaction completed |
| reversed_at | TIMESTAMPTZ | When transaction was reversed |

## API Compatibility

### Old PlatformFeeRepository Interface

All methods continue to work via the `PlatformFeeRepositoryWrapper`:

```go
GetWallet(ctx, userID) -> Returns legacy UserWallet struct
GetOrCreateWallet(ctx, userID) -> Creates unified wallet, returns legacy struct
GetWalletBalance(ctx, userID) -> Returns balance from unified wallet
CreditWallet(ctx, userID, amount, stripeRef) -> Credits unified wallet
DebitWallet(ctx, userID, amount, description) -> Debits unified wallet
HasWalletCreditReference(ctx, ref) -> Checks unified transaction ledger
```

### Old billing.Controller Interface

All methods continue to work via the `BillingControllerWrapper`:

```go
GetOrCreateControls(ctx, agentID) -> Returns legacy AgentBillingControls
CheckSpendCap(ctx, agentID, cost) -> Checks unified wallet spend caps
ConsumeCredits(ctx, agentID, amount) -> Debits unified wallet
AddCredits(ctx, agentID, amount) -> Credits unified wallet
GetAgentSpend(ctx, agentID, period) -> Returns spend from unified ledger
UpdateSpendCap(ctx, agentID, daily, monthly) -> Updates unified wallet
```

## Rollback Procedure

If issues are detected:

1. **Disable unified wallet (immediate)**
   ```go
   stripeWebhookHandler.DisableUnifiedWallet()
   ```

2. **Verify legacy systems are working**
   - Check that new credits go to old tables
   - Check that debits work from old tables

3. **Investigate issues**
   - Check logs for errors
   - Verify data consistency

4. **Fix issues and re-enable**
   ```go
   stripeWebhookHandler.EnableUnifiedWalletForUsers()
   stripeWebhookHandler.EnableUnifiedWalletForAgents()
   ```

5. **Full rollback (if needed)**
   ```bash
   # Restore data from backup (if migration was destructive)
   # Or simply continue with dual-write mode
   ```

## Data Consistency Verification

After migration, verify data consistency:

```sql
-- Check that all user wallets were migrated
SELECT 
    COUNT(*) as unified_wallets,
    (SELECT COUNT(*) FROM user_wallets) as legacy_wallets
FROM wallets 
WHERE owner_type = 'user';

-- Check that all agent wallets were migrated
SELECT 
    COUNT(*) as unified_wallets,
    (SELECT COUNT(*) FROM agent_billing_controls) as legacy_wallets
FROM wallets 
WHERE owner_type = 'agent';

-- Verify balance totals match
SELECT 
    'user' as wallet_type,
    SUM(balance_usd) as unified_total,
    (SELECT SUM(balance_usd) FROM user_wallets) as legacy_total
FROM wallets 
WHERE owner_type = 'user'
UNION ALL
SELECT 
    'agent' as wallet_type,
    SUM(balance_usd) as unified_total,
    (SELECT SUM(credit_balance_usd) FROM agent_billing_controls) as legacy_total
FROM wallets 
WHERE owner_type = 'agent';
```

## Post-Migration Cleanup

After full migration is complete and verified:

1. Remove legacy repository code
2. Remove compatibility wrappers
3. Update all code to use unified wallet interfaces directly
4. Archive or drop old tables (after backup)
5. Remove feature flags

## Monitoring

During migration, monitor:

1. **Wallet balance drift**
   - Compare unified vs legacy balances periodically
   - Alert if differences exceed threshold

2. **Transaction counts**
   - Ensure all transactions are recorded in unified ledger
   - Monitor for duplicate transactions

3. **Error rates**
   - Monitor unified wallet operation errors
   - Track fallback to legacy system frequency

4. **Performance**
   - Unified wallet adds polymorphic queries
   - Monitor query performance and add indexes if needed

## Support

For issues during migration:

1. Check this guide for common procedures
2. Review logs for specific error messages
3. Verify database connectivity and permissions
4. Test in staging environment before production
5. Contact the platform team for assistance

## Related Documents

- `internal/wallet/README.md` - Wallet package documentation
- `migrations/000200_unified_wallet_system.up.sql` - Schema details
- `cmd/migrate-wallets/migrate.go` - Migration tool source
