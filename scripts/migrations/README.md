# Fix for Incorrect Tenant Assignment Bug

## Problem Summary

A bug in the signup flow caused new users to be incorrectly assigned to existing tenants (often enterprise tenants with the 'enterprise' plan) instead of creating a new individual tenant with the 'starter' plan for each user.

### Root Cause

The signup code in `internal/auth/user_auth.go` and `internal/auth/oauth.go` used this logic:

```go
// BUG: Listed ALL tenants and used the first one!
tenants, err := a.repo.ListTenants()
if len(tenants) > 0 {
    tenantID = tenants[0].ID  // Wrong! Reused existing tenant
}
```

### Impact

- New users were associated with existing tenants (potentially enterprise tenants)
- These users incorrectly saw enterprise features and SLA status cards
- Multiple users shared the same tenant instead of each having their own

## Fix Applied

### 1. Code Fix (Already Applied)

The signup code now always creates a new tenant for each user:

```go
// FIXED: Always create new tenant for each user
tenant, err := a.repo.CreateTenant(context.Background(), "Default Tenant")
tenantID := tenant.ID
```

### 2. Frontend Fallback (Already Applied)

Changed plan defaulting from `??` to `||` to handle empty strings:

```typescript
// Before: plan: userData.user.plan ?? 'starter',  // Only null/undefined
// After:
plan: userData.user.plan || 'starter',  // Also catches empty strings
```

## Data Migration for Affected Users (Neon)

### Prerequisites

Ensure you have the Neon connection string from your `.env` file:

```bash
export DATABASE_URL="postgresql://username:password@ep-xxx-xxx.us-east-1.aws.neon.tech/functionfly?sslmode=require"
```

### Step 1: Preview Affected Users

Run this to see which users need fixing (no changes made):

```bash
# Using psql with Neon (sslmode=require is required)
psql "$DATABASE_URL" -f scripts/migrations/preview_affected_users.sql
```

Or via Neon's SQL Editor:

1. Open [Neon Console](https://console.neon.tech)
2. Go to your project → SQL Editor
3. Paste contents of `preview_affected_users.sql`
4. Run (this is read-only, safe to run)

### Step 2: Run the Migration

**Important**: Neon supports transactions. In the SQL Editor, each statement runs in its own transaction by default. The migration script uses a `DO $$` block which runs as a single atomic operation.

#### Option A: Neon SQL Editor (Easiest)

1. Open [Neon Console](https://console.neon.tech) → SQL Editor
2. Paste contents of `fix_user_tenants.sql`
3. Review the preview output at the top
4. Check the NOTICE messages showing which users will be fixed
5. If correct, the DO block will auto-commit when complete

#### Option B: psql (Recommended for large datasets)

```bash
# Run the full migration
psql "$DATABASE_URL" -f scripts/migrations/fix_user_tenants.sql
```

#### Option C: Go Migration Script (With dry-run)

```bash
# Set your Neon connection string
cd /home/micro/projects/functionfly
export DATABASE_URL="postgresql://username:password@ep-xxx-xxx.us-east-1.aws.neon.tech/functionfly?sslmode=require"

# Preview only (dry run - no changes made)
go run scripts/migrations/fix_user_tenants.go

# Apply changes
export MIGRATION_AUTO_COMMIT=true
go run scripts/migrations/fix_user_tenants.go
```

### Step 3: Verify the Fix

```sql
-- Check tenant distribution after migration
SELECT
    t.plan,
    COUNT(DISTINCT u.tenant_id) as tenant_count,
    COUNT(u.id) as user_count
FROM users u
JOIN tenants t ON u.tenant_id = t.id
GROUP BY t.plan
ORDER BY t.plan;

-- All users should now have their own tenant
-- Expected: tenant_count ≈ user_count for 'starter' plan
```

## Migration Logic

The migration assumes:

1. **First user in a shared tenant** = Legitimate tenant owner (kept in place)
2. **All other users** = Incorrectly assigned (moved to new starter tenants)

This conservative approach ensures we don't accidentally move the original tenant owner.

## Safety Notes

- **Always run in a transaction** so you can rollback if needed
- **Preview first** using the preview SQL to understand the scope
- **Test on a copy** of production data before running on live database
- The migration creates new tenants with 'starter' plan and moves users to them
- Original tenants are preserved (not deleted)

## Neon-Specific Notes

### Connection String Format

Neon requires `sslmode=require` in the connection string:

```
postgresql://[user]:[password]@[neon-host]/[dbname]?sslmode=require
```

### Branch Considerations

If using Neon branches:

1. Run the migration on a **branch first** to test
2. Create branch: `neonctl branches create --name migration-test`
3. Get branch connection string: `neonctl connection-string --branch migration-test`
4. Run migration on branch, verify, then apply to main

### Connection Pooling

Neon uses connection pooling (PgBouncer). For scripts that need direct access:

```bash
# Use the direct connection string (port 5432 instead of pooled 5433)
# Found in Neon Console → Connection Details → "Connection without pooling"
export DATABASE_URL="postgresql://...neon.tech:5432/functionfly?sslmode=require"
```

## Post-Migration (CRITICAL: Fly.io Restart)

After running the SQL migration, you **MUST restart the Fly.io app** to clear:

- **Prepared statement caches** (may reference old tenant associations)
- **In-memory user caches** on running machines
- **Redis session data** (if applicable)

### Restart Fly.io App

```bash
# Option 1: Rolling restart (zero downtime, recommended)
fly apps restart functionfly-control

# Option 2: Full redeploy (if you also have code changes)
fly deploy --app functionfly-control
```

### Clear Redis Cache (if using Redis)

```bash
# Connect to Redis
fly redis connect --app functionfly-control

# Inside Redis CLI, clear user cache:
EVAL "return redis.call('del', unpack(redis.call('keys', 'user:*')))" 0
# Or clear all:
FLUSHDB
exit
```

### Verification Steps

After restart:

1. **Check app status**: `fly status --app functionfly-control`
2. **Check logs**: `fly logs --app functionfly-control -n 50`
3. **Test login** with affected user (e.g., `traseputallaz@gmail.com`)
4. **Verify plan**: After login, check the dashboard shows "Starter" not "Enterprise"

### What Users Need to Do

1. **Re-login** - Users will need to log out and back in (or clear browser cookies)
2. **No data migration** - Any existing data (functions, apps) stays with the original enterprise tenant
3. **Fresh start** - Users now have their own starter tenant

### Files

4. **Neon users**: The old tenant rows are preserved in your database ( Neon storage is not affected)

## Files

- `preview_affected_users.sql` - Preview only, no changes
- `fix_user_tenants.sql` - SQL migration with transaction support
- `fix_user_tenants.go` - Go script alternative with dry-run mode
