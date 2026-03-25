# Account Sharing & Seat Management

> **Reference:** Full design spec in [`plans/ACCOUNT_SHARING.md`](../plans/ACCOUNT_SHARING.md)

## Quick Reference

### Seat Limits by Plan

| Plan | Max Users | Behavior at Limit |
|------|-----------|-------------------|
| Starter | 3 | Cannot invite/add beyond 3 active users |
| Pro | 10 | Cannot invite/add beyond 10 active users |
| Enterprise | Unlimited (-1) | No limit enforced |

### Key Thresholds

- **Warning threshold:** 80% of seat limit → tenant receives warning notification
- **Grace period:** 30 days when a plan downgrade causes the tenant to exceed the new seat limit

### User States

| State | Counts Against Seat Limit | Can Login | Restorable |
|-------|---------------------------|-----------|------------|
| Active | ✅ | ✅ | — |
| Deactivated | ❌ | ❌ | ✅ (via Reactivate) |
| Deleted | ❌ | ❌ | ❌ |

### Admin API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/admin/tenants/{tenantId}/seat-usage` | GET | Get seat usage for a tenant |
| `/v1/admin/users/{userId}/deactivate` | POST | Soft-delete (deactivate) a user |
| `/v1/admin/users/{userId}/reactivate` | POST | Reactivate a deactivated user |

### Grace Period Logic

When a tenant's plan is downgraded to a plan with fewer seats:

1. If current active users ≤ new seat limit → no grace period needed
2. If current active users > new seat limit → `SeatGracePeriodEnd` is set to 30 days from downgrade
3. During grace period: tenant can continue using all users but cannot add new users
4. After grace period: oldest deactivated users remain deactivated; if still over limit, oldest active users are auto-deactivated (up to the limit)

### Key Implementation Details

- **Soft delete only:** Users are never hard-deleted via the admin API. `DeactivatedAt` and `DeactivatedBy` are set; data is preserved.
- **Billing:** Billing is per-tenant, not per-seat. Sharing within seat limits has no extra cost.
- **DeactivatedBy:** Records which admin deactivated a user for audit purposes.
- **No automatic payment-impacting actions:** Auto-deactivation during grace period does not trigger refunds or plan downgrades.
