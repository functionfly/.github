# Bundle Provisioning: Full Checkout Flow E2E Tests

**Date:** 2026-06-25
**Status:** Approved
**Scope:** Integration and E2E tests for the bundle checkout → webhook → provisioning pipeline with full tenant isolation verification.

---

## Problem

The bundle provisioning system has unit-level tests for individual components (provisioner, payment requirements, handler auth) but no tests covering the complete checkout flow end-to-end. Critical paths are untested:

- Checkout session creation through the billing handler
- Stripe webhook processing that triggers provisioning
- Multi-tenant isolation during concurrent checkouts
- Failure propagation (one tenant's failure blocking another)
- Webhook idempotency and replay safety
- Founder mode deferred billing lifecycle

## Approach

Two-layer test suite in `internal/e2e/`:

| Layer | File | Dependencies | Runs in CI |
|-------|------|-------------|------------|
| **Layer 1: Handler integration** | `bundle_checkout_test.go` | Mocks only (repo, provisioner, Stripe) | Yes (`go test ./internal/e2e/...`) |
| **Layer 2: Full E2E** | `bundle_checkout_e2e_test.go` | Real Stripe test keys, real Postgres | No (build tag `e2e`, manual trigger) |

Shared helpers in `testutil_test.go` and mock implementations in `mocks_test.go`.

---

## File Structure

```
internal/e2e/
├── bundle_checkout_test.go          # Layer 1: 10 handler-level integration tests
├── bundle_checkout_e2e_test.go      # Layer 2: 5 full E2E tests (//go:build e2e)
├── testutil_test.go                 # Shared helpers, fixture builders
└── mocks_test.go                    # Mock implementations
```

Package: `package e2e_test` (black-box testing).

---

## Mock Architecture

### MockRepo

Implements `storage.Repository` subset needed by billing handlers:

- `ListPricingBundles(ctx, active) ([]PricingBundle, error)` — returns seeded bundles
- `GetPricingBundleBySlug(ctx, slug) (*PricingBundle, error)`
- `GetPricingBundleByID(ctx, id) (*PricingBundle, error)`
- `GetUserByID(ctx, userID) (*User, error)` — returns seeded users
- `CreateBundleSubscription(ctx, sub) error` — stores in-memory, tracks calls
- `GetBundleSubscriptionByTenant(ctx, tenantID) (*BundleSubscription, error)`
- `CreateFounderModeRegistration(ctx, reg) error`
- `GetActiveFounderMode(ctx, tenantID, bundleID) (*FounderModeRegistration, error)`
- `ListFounderModesByTenant(ctx, tenantID) ([]FounderModeRegistration, error)`
- `ListActiveFounderModesByTenant(ctx, tenantID) ([]FounderModeRegistration, error)`
- `CountActiveFounderModeRegistrations(ctx) (int, error)`
- `CountRecentSuccessfulDeployments(ctx) (int, error)`
- `GetStripeSyncEventByEventID(ctx, eventID) (*StripeSyncEvent, error)` — for idempotency
- `CreateStripeSyncEvent(ctx, event) error`

### MockProvisioner

Function type matching `provisionBundleFn`: `func(ctx, tenantID, bundleSlug) (string, int, error)`.

Tracks:
- `CallLog []ProvisionCall` — records every invocation with tenantID + bundleSlug
- `CallCount() int`
- `LastError` — configurable per-tenant failure injection
- `TenantErrors map[uuid.UUID]error` — inject failures for specific tenants

### Stripe Mocking Strategy

`payment.CreateBundleCheckoutSession` calls `session.New(params)` directly — no injectable client. For Layer 1, mock Stripe at the HTTP level:

```go
func newMockStripeServer(t *testing.T) *httptest.Server {
    mux := http.NewServeMux()
    mux.HandleFunc("/v1/checkout/sessions", func(w http.ResponseWriter, r *http.Request) {
        // Parse params, return canned session
        json.NewEncoder(w).Encode(map[string]interface{}{
            "id":  "cs_test_" + uuid.New().String()[:8],
            "url": "https://checkout.stripe.com/test",
            "object": "checkout.session",
        })
    })
    mux.HandleFunc("/v1/customers", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "id":     "cus_test_" + uuid.New().String()[:8],
            "object": "customer",
        })
    })
    return httptest.NewServer(mux)
}
```

Set `stripe.Key = "sk_test_fake"` and override the Stripe backend base URL to the mock server. The `stripe-go` SDK allows setting `stripe.Backend` to a custom `*httptest.Server` URL via `stripe.SetBackend(stripe.APIBackend, mockBackend)`.

For Layer 2, use real Stripe test keys — no mocking needed.

Metadata verification happens at the mock HTTP level: the mock server parses the request body and asserts `metadata[bundle_slug]`, `metadata[tenant_id]`, `metadata[purpose]` are present and correct.

### MockEmailService

Implements `email.Service`. Captures sent emails for assertion.

---

## Test Scenarios

### Layer 1: Handler Integration Tests (`bundle_checkout_test.go`)

| # | Test Name | Scenario | Assertions |
|---|-----------|----------|------------|
| 1 | `TestBundleCheckout_HappyPath` | `POST /v1/billing/bundles/saas-starter/checkout` with valid auth | 200, session URL returned, Stripe metadata has `bundle_slug=saas-starter` + `tenant_id` |
| 2 | `TestBundleCheckout_FounderMode` | `POST /v1/billing/bundles/saas-starter/founder` with valid auth | 201, deferred subscription created, provisioner called async |
| 3 | `TestBundleCheckout_WebhookTriggersProvisioning` | POST simulated `checkout.session.completed` event with `purpose=bundle_subscription` | Subscription created with `active` status, provisioner called with correct tenantID + slug |
| 4 | `TestBundleCheckout_TenantIsolation` | Two tenants checkout simultaneously (goroutines) | Each gets separate subscription records, provisioner called with distinct tenantIDs, no cross-contamination |
| 5 | `TestBundleCheckout_ProvisioningFailureIsolation` | Tenant A provisioner fails, Tenant B succeeds | A's subscription marked `failed`, B's subscription `active`, B not blocked |
| 6 | `TestBundleCheckout_WebhookIdempotency` | Same `checkout.session.completed` event sent twice | Provisioner called exactly once, second webhook returns 200 without re-processing |
| 7 | `TestBundleCheckout_InvalidSlug` | `POST /v1/billing/bundles/invalid-slug/checkout` | 400, "Invalid bundle slug" |
| 8 | `TestBundleCheckout_Unauthenticated` | No auth context on checkout request | 401 |
| 9 | `TestBundleCheckout_BillingNotConfigured` | `payment.IsConfigured()=false` | 503, "Billing is not configured" |
| 10 | `TestBundleCheckout_ChangeBundle` | `POST /v1/billing/bundle/change` with `new_bundle_slug=marketplace` | 200, new checkout session created with marketplace price |

### Layer 2: Full E2E Tests (`bundle_checkout_e2e_test.go`)

Guarded by `//go:build e2e`. Skipped unless `TEST_STRIPE_KEY` and `TEST_DB_URL` are set.

| # | Test Name | Scenario | Assertions |
|---|-----------|----------|------------|
| 1 | `TestE2E_BundleCheckout_FullFlow` | Real Stripe checkout session → signed webhook → DB provisioning | Subscription in Postgres, `tenant_bundle_state` row exists, all components `active` |
| 2 | `TestE2E_MultiTenantDBIsolation` | Two tenants provision via real DB | Separate `tenant_bundle_state` rows, separate component `resource_id` values |
| 3 | `TestE2E_WebhookSignatureVerification` | Tampered webhook signature | Rejected with 401 |
| 4 | `TestE2E_FounderModeLifecycle` | Register founder → trigger threshold → convert to paid → provision | Full state machine: `deferred` → `active` |
| 5 | `TestE2E_ProvisioningIdempotency` | Provision same tenant twice via real DB | Same resource IDs returned, no duplicate rows |

---

## Helper Functions (`testutil_test.go`)

| Function | Purpose |
|----------|---------|
| `newTestRouter(handler, webhookHandler) *mux.Router` | Registers billing + webhook routes |
| `buildBundleCheckoutPayload(slug, successURL, cancelURL) io.Reader` | JSON request body |
| `buildStripeCheckoutCompletedEvent(sessionID, tenantID, bundleSlug, purpose) []byte` | Stripe webhook event JSON |
| `buildStripeCheckoutCompletedEventRaw(sessionID, tenantID, bundleSlug, purpose) json.RawMessage` | Event.Data.Raw for unverified webhooks |
| `seedTestBundle(repo, slug, priceID) uuid.UUID` | Inserts pricing bundle into mock repo |
| `seedTestUser(repo, tenantID, email) uuid.UUID` | Inserts test user into mock repo |
| `waitForAsyncProvisioning(timeout time.Duration, check func() bool)` | Polls for async provisioning completion (100ms intervals, configurable timeout) |
| `setAuthContext(r, userID, tenantID) *http.Request` | Injects auth claims into request context |
| `newTestHandler(repo, provisionerFn) *Handler` | Creates billing handler with injected mocks |
| `newTestWebhookHandler(repo) *StripeWebhookHandler` | Creates webhook handler with mock repo |

---

## Wiring Details

### Layer 1 (Mock-based)

```
Setup:
  mockRepo := NewMockRepo()
  mockProvisioner := NewMockProvisioner()
  handler := billing.NewHandler(mockRepo, nil, nil, nil)
  handler.SetBundleProvisioner(mockProvisioner.Call)
  webhookHandler := NewMockWebhookHandler(mockRepo)
  router := newTestRouter(handler, webhookHandler)
  server := httptest.NewServer(router)

Webhook simulation:
  os.Setenv("ALLOW_UNVERIFIED_WEBHOOKS", "true")
  os.Setenv("DEVELOPMENT", "true")
  payload := buildStripeCheckoutCompletedEvent(...)
  resp, _ := http.Post(server.URL+"/webhooks/stripe", "application/json", bytes.NewReader(payload))

Assertions:
  assert.Equal(t, 1, mockProvisioner.CallCount())
  assert.Equal(t, tenantID, mockProvisioner.CallLog[0].TenantID)
```

### Layer 2 (Real deps)

```
Setup:
  db, _ := sql.Open("postgres", os.Getenv("TEST_DB_URL"))
  stripe.Key = os.Getenv("TEST_STRIPE_KEY")
  webhookSecret := os.Getenv("TEST_WEBHOOK_SECRET")

Cleanup:
  t.Cleanup(func() {
      db.Exec("DELETE FROM tenant_bundle_state WHERE tenant_id = $1", tenantID)
      db.Exec("DELETE FROM bundle_subscriptions WHERE tenant_id = $1", tenantID)
      // Delete Stripe test customer
  })

Webhook simulation:
  payload := buildStripeCheckoutCompletedEvent(...)
  sig := webhook.GenerateTestSignedPayload(webhookSecret, payload, time.Now())
  req, _ := http.NewRequest("POST", server.URL+"/webhooks/stripe", bytes.NewReader(payload))
  req.Header.Set("Stripe-Signature", sig)
```

---

## Skip Conditions

### Layer 1
No skip conditions — runs with standard `go test`.

### Layer 2
```go
func requireE2EEnv(t *testing.T) {
    t.Helper()
    if os.Getenv("TEST_STRIPE_KEY") == "" {
        t.Skip("E2E: TEST_STRIPE_KEY not set")
    }
    if os.Getenv("TEST_DB_URL") == "" {
        t.Skip("E2E: TEST_DB_URL not set")
    }
}
```

---

## CI Integration

**Layer 1** runs on every push via existing `go test ./internal/...` (included in `make test-short`).

**Layer 2** runs manually or in deploy pipeline:
```bash
TEST_STRIPE_KEY=sk_test_... TEST_DB_URL="postgres://..." go test -tags=e2e ./internal/e2e/...
```

---

## Dependencies

No new external dependencies. Uses existing:
- `github.com/stretchr/testify` (assert/require)
- `github.com/google/uuid`
- `github.com/gorilla/mux`
- `github.com/stripe/stripe-go/v83` (Layer 2 only)
- `net/http/httptest` (stdlib)

---

## Success Criteria

1. All 10 Layer 1 tests pass with `go test ./internal/e2e/...`
2. All 5 Layer 2 tests pass with `go test -tags=e2e ./internal/e2e/...` (when env vars set)
3. No test leaks state (cleanup verified)
4. Concurrent tenant tests demonstrate no cross-contamination
5. Idempotency tests verify no duplicate provisioning on replayed webhooks
