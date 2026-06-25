# PCI DSS v4.0 Full Production Audit Trail — Design

**Date:** 2026-06-25
**Status:** Implemented

---

## Overview

This design covers implementing a complete PCI DSS v4.0 compliant audit trail for all cardholder data environment (CDE) access across the FunctionFly platform. The existing `pci_audit_repository.go` and `pci_audit_events` table provide the foundation; this design addresses the **integration gap** — no handler was calling the PCI audit logging.

---

## Existing Infrastructure

| Component | Location | Status |
|-----------|----------|--------|
| `pci_audit_events` table | `migrations/20260419045430_create_pci_audit_tables.up.sql` | ✅ Exists |
| `PCIAuditEvent` model | `internal/types/models_migrated.go` | ✅ Exists |
| `PCIAuditRepository` | `internal/storage/pci_audit_repository.go` | ✅ Exists (append-only, chain-hashed) |
| PCI event types | `internal/types/models_migrated.go:3858-3880` | ✅ Payment/Key/Auth types exist |
| Severity constants | `PCISeverityInfo/Warning/Critical/Emergency` | ✅ Defined |

---

## Scope: What's Implemented

### 1. PCI Audit Helper (`internal/api/helpers/pci_audit.go`)

A shared helper that:
- Extracts actor context from `http.Request` (userID, email, role, IP, user agent, sessionID)
- Wraps `PCIAuditRepository.LogPCIAuditEvent` for async non-blocking calls
- Provides typed convenience methods for each event category

### 2. Admin PCI Event Types (`internal/types/models_migrated.go`)

New event types for admin actions on cardholder data:

| Event Type | Description |
|-----------|-------------|
| `admin_manual_refund` | Admin issues a manual refund |
| `admin_plan_change` | Admin changes a tenant's plan |
| `admin_invoice_adjustment` | Admin adjusts an invoice |
| `admin_payment_method_override` | Admin replaces payment method |
| `admin_subscription_cancel` | Admin cancels a subscription |
| `admin_discount_override` | Admin applies manual discount |

### 3. Admin Convenience Methods (`internal/storage/pci_audit_repository.go`)

New methods:
- `LogAdminAction()` — for admin billing operations on cardholder data
- `LogCardDataAccessAsync()` — async wrapper for cardholder data read/write

### 4. Billing Handler Integration (`internal/api/handlers/billing/handler.go`)

PCI audit logging added at these touchpoints:

| Handler | Event Type | Severity |
|---------|-----------|----------|
| `HandleGetSubscription` | `card_data_read` | info |
| `HandleListPaymentMethods` | `card_data_read` | info |
| `HandleCreateSetupIntent` | `card_data_tokenized` | info |
| `HandleSetDefaultPaymentMethod` | `card_data_write` | warning |
| `HandleDetachPaymentMethod` | `card_data_delete` | warning |
| `HandleCreateCheckoutSession` | `payment_initiated` | info |
| `HandleCreatePortalSession` | `card_data_read` | info |
| `HandleCancelSubscription` | `admin_subscription_cancel` | warning |

### 5. Stripe Webhook Integration (`internal/api/handlers/webhooks/stripe.go`)

PCI audit logging at these webhook events:

| Webhook Event | PCI Event Type | Severity |
|--------------|---------------|----------|
| `invoice.payment_succeeded` | `payment_processed` | info |
| `invoice.payment_failed` | `payment_failed` | warning |
| `charge.refunded` | `refund_processed` | info |
| `charge.dispute.created` | `chargeback_received` | critical |
| `charge.dispute.funds_withdrawn` | `chargeback_received` | critical |
| `checkout.session.completed` | `payment_initiated` | info |
| `payment_intent.payment_fail` | `payment_failed` | warning |

### 6. Wallet Handler Integration (`internal/api/handlers/billing/wallet.go`)

PCI audit logging for:
- Low balance threshold alerts (when wallet balance drops below `AGENT_WALLET_LOW_BALANCE_USD`)
- Agent wallet credit operations

---

## Severity Mapping

| Severity | When used |
|----------|-----------|
| `emergency` | Encryption key destroy, security incidents, failed tampering detection |
| `critical` | Chargebacks, failed access attempts, key rotation/retirement |
| `warning` | Admin actions, payment failures, card data write/delete |
| `info` | Card data reads, payment initiated/processed, successful transactions |

---

## Retention

| Event Severity | Retention Period |
|---------------|-----------------|
| `info`, `warning` | 1 year |
| `critical`, `emergency` | 3 years |

---

## Async Logging Pattern

All PCI audit logging uses async goroutines to avoid blocking payment flows:

```go
go func() {
    helper.LogPCIAuditEventAsync(ctx, params)
}()
```

Errors are logged but never fail the parent operation.

---

## Actor Context Extraction

The helper extracts from `http.Request`:

| Field | Source |
|-------|--------|
| `ActorUserID` | `middleware.GetUserFromContext(r).UserID` |
| `ActorEmail` | `middleware.GetUserFromContext(r).Email` |
| `ActorRole` | `middleware.GetUserFromContext(r).Role` |
| `ActorIP` | `r.RemoteAddr` (or `X-Forwarded-For`) |
| `ActorUserAgent` | `r.Header.Get("User-Agent")` |
| `SessionID` | `r.Header.Get("X-Session-ID")` |
| `RequestID` | `middleware.GetRequestID(r.Context())` |

---

## Chain Hash Integrity

Each PCI audit event's `chain_hash` is computed as:
```
SHA256(event_id + event_type + created_at + previous_chain_hash)
```

This creates an immutable, tamper-evident chain — any gap or modification is detectable via `VerifyAuditChain()`.

---

## Files Changed

| File | Change |
|------|--------|
| `internal/api/helpers/pci_audit.go` | **NEW** — actor context extraction + async logging |
| `internal/types/models_migrated.go` | **ADD** — admin PCI event type constants |
| `internal/storage/pci_audit_repository.go` | **ADD** — `LogAdminAction()` + async wrappers |
| `internal/api/handlers/billing/handler.go` | **ADD** — PCI audit calls at cardholder data touchpoints |
| `internal/api/handlers/webhooks/stripe.go` | **ADD** — PCI audit calls at payment event handlers |
| `internal/api/handlers/billing/wallet.go` | **ADD** — PCI audit for wallet balance alerts |
| `internal/api/routes.go` | **ADD** — inject `pciAuditRepo` into billing handler |
