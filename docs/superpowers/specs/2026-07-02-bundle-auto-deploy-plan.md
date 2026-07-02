# Bundle Auto-Deploy: Implementation Plan

**Spec:** `docs/superpowers/specs/2026-07-02-bundle-auto-deploy-design.md`

---

## Phase 1: DB Schema + Template Migration (no behavior change)

### 1.1 Migration file

Create `migrations/20260702170000_bundle_auto_deploy.up.sql`:

```sql
-- Extend bundle_subscriptions with deployment tracking
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS deploy_status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS deploy_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS deploy_error TEXT;
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS deployed_at TIMESTAMPTZ;
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS provider_id UUID REFERENCES providers(id);
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS script_name TEXT;
ALTER TABLE bundle_subscriptions ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

-- Index for retry ticker
CREATE INDEX IF NOT EXISTS idx_bundle_subscriptions_deploy_status
    ON bundle_subscriptions(deploy_status) WHERE deploy_status IN ('failed', 'awaiting_provider');

-- Bundle function templates
CREATE TABLE IF NOT EXISTS bundle_function_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_slug TEXT NOT NULL,
    function_name TEXT NOT NULL,
    runtime TEXT NOT NULL DEFAULT 'js',
    code TEXT NOT NULL,
    route_path TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(bundle_slug, function_name, version)
);

CREATE INDEX IF NOT EXISTS idx_bundle_function_templates_slug
    ON bundle_function_templates(bundle_slug);

-- Seed templates (migrate from hardcoded Go)
INSERT INTO bundle_function_templates (bundle_slug, function_name, runtime, code, route_path, version)
VALUES
    ('saas-starter', 'stripe-webhook', 'js', $CODE$, '/stripe-webhook', 1),
    ('saas-starter', 'welcome-email', 'js', $CODE$, '/welcome-email', 1),
    ('marketplace', 'create-listing', 'js', $CODE$, '/create-listing', 1),
    ('marketplace', 'send-message', 'js', $CODE$, '/send-message', 1),
    ('ai-app', 'chat-completion', 'js', $CODE$, '/chat-completion', 1),
    ('ai-app', 'embed-and-store', 'js', $CODE$, '/embed-and-store', 1)
ON CONFLICT (bundle_slug, function_name, version) DO NOTHING;
```

### 1.2 Storage layer

Add to `internal/storage/bundle_repository.go`:

- `UpdateBundleSubscription(ctx, sub *BundleSubscription) error`
- `ListBundleTemplates(ctx, bundleSlug string) ([]*BundleFunctionTemplate, error)`
- `ListPendingDeployments(ctx) ([]*BundleSubscription, error)` — for retry ticker
- `ListAwaitingProvider(ctx, tenantID uuid.UUID) ([]*BundleSubscription, error)` — for provider-connect hook

### 1.3 Types

Add to `internal/types/types.go`:

```go
type BundleFunctionTemplate struct {
    ID           uuid.UUID `json:"id"`
    BundleSlug   string    `json:"bundle_slug"`
    FunctionName string    `json:"function_name"`
    Runtime      string    `json:"runtime"`
    Code         string    `json:"code"`
    RoutePath    string    `json:"route_path"`
    Version      int       `json:"version"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

Extend `BundleSubscription` with new fields.

**Verify:** `make build-fast` passes. No behavior change.

---

## Phase 2: CF Workers Deployment Engine

### 2.1 Template adapter — `internal/api/handlers/billing/bundles_templates.go`

**New file.** Functions:

- `generateWorkerScript(bundleSlug string, templates []*BundleFunctionTemplate) []byte` — combines templates into a single CF worker with URL router
- `adaptStateAPI(code string) string` — rewrites `state.set/get/push` to KV calls
- `adaptAIAPI(code string) string` — rewrites `ai.chat.completions.create` etc. to FunctionFly API proxy fetch calls
- `wrapWithRouter(templates []adaptedTemplate) string` — generates the `export default { fetch() { ... } }` dispatcher

Template adaptation uses string replacement (not AST parsing) for MVP. The templates are controlled by us, not user input.

### 2.2 Deploy orchestrator — `internal/api/handlers/billing/bundles_deploy.go`

**New file.** Core functions:

```go
// DeployBundle deploys bundle templates to the user's provider.
func (h *Handler) DeployBundle(ctx context.Context, sub *storage.BundleSubscription) error

// deployToCloudflare handles CF Workers deployment.
func (h *Handler) deployToCloudflare(ctx context.Context, sub *storage.BundleSubscription, provider *storage.Provider) error

// failDeploy records a deployment failure and schedules retry.
func (h *Handler) failDeploy(sub *storage.BundleSubscription, format string, args ...interface{}) error

// validateProviderAccess checks if the provider token has required permissions.
func (h *Handler) validateProviderAccess(ctx context.Context, provider *storage.Provider) error
```

`deployToCloudflare` flow:
1. Get account ID from CF API (`GET /accounts`)
2. Get workers.dev subdomain (`GET /accounts/{id}/workers/subdomain`)
3. Check idempotency (backend already exists?)
4. Create KV namespace `{script_name}-state`
5. Generate worker script with KV binding
6. Deploy via `CloudflareDeploymentClient.Deploy()`
7. Enable workers.dev via `EnableWorkersDev()`
8. On partial failure: cleanup (delete script + KV namespace)
9. Create backend row with worker URL
10. Update subscription: `deploy_status = "deployed"`

### 2.3 CloudflareDeploymentClient additions

In `internal/adapters/cloudflare/deployment.go`, add:

- `CreateKVNamespace(ctx, name string) (namespaceID string, error)` — `POST /accounts/{id}/storage/kv/namespaces`
- `DeleteKVNamespace(ctx, namespaceID string) error` — `DELETE /accounts/{id}/storage/kv/namespaces/{id}`
- `GetAccountID(ctx) (string, error)` — `GET /accounts`
- `GetWorkersSubdomain(ctx) (string, error)` — `GET /accounts/{id}/workers/subdomain`

### 2.4 Webhook integration

In `internal/api/handlers/webhooks/stripe.go`, modify `handleBundleSubscriptionCheckout`:

```go
// After existing provisioning (app + functions):
sub, err := h.repo.GetBundleSubscription(ctx, tenantID, bundleSlug)

providerID := session.Metadata["provider_id"]
if providerID != "" {
    sub.ProviderID = &providerID
    sub.DeployStatus = "deploying"
    sub.ScriptName = fmt.Sprintf("%s-%s", tenantID.String()[:8], bundleSlug)
    h.repo.UpdateBundleSubscription(ctx, sub)

    go h.DeployBundle(context.Background(), sub)
} else {
    sub.DeployStatus = "awaiting_provider"
    h.repo.UpdateBundleSubscription(ctx, sub)
}
```

**Verify:** `make build-fast` passes. Manual test: insert a bundle subscription with `deploy_status = 'deploying'` and a real CF provider, call `DeployBundle` directly.

---

## Phase 3: Retry Ticker + Provider-Connect Hook

### 3.1 Retry ticker — `internal/api/handlers/billing/bundles_retry.go`

**New file.** Background goroutine started in `NewHandler()` or `Server.Start()`:

```go
func (h *Handler) StartDeployRetryTicker(ctx context.Context) {
    ticker := time.NewTicker(60 * time.Second)
    go func() {
        for {
            select {
            case <-ctx.Done():
                ticker.Stop()
                return
            case <-ticker.C:
                h.retryFailedDeployments(ctx)
            }
        }
    }()
}

func (h *Handler) retryFailedDeployments(ctx context.Context) {
    subs, _ := h.repo.ListPendingDeployments(ctx)
    for _, sub := range subs {
        if sub.NextRetryAt != nil && sub.NextRetryAt.After(time.Now()) {
            continue
        }
        h.DeployBundle(ctx, sub)
    }
}
```

Backoff calculation in `failDeploy`:
```go
backoff := []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute}
nextRetry := backoff[min(sub.DeployAttempts, len(backoff)-1)]
sub.NextRetryAt = timePtr(time.Now().Add(nextRetry))
```

### 3.2 Provider-connect hook

In `internal/api/handlers/providers/connect.go`, after successful provider connection:

```go
// Check for bundle subscriptions awaiting provider
awaiting, _ := h.repo.ListAwaitingProvider(ctx, tenantID)
for _, sub := range awaiting {
    sub.ProviderID = &provider.ID
    sub.DeployStatus = "deploying"
    h.repo.UpdateBundleSubscription(ctx, sub)
    go h.DeployBundle(context.Background(), sub)
}
```

**Verify:** `make build-fast` passes. Manual test: create subscription with `awaiting_provider`, connect a CF provider, verify deployment triggers.

---

## Phase 4: Dashboard UI

### 4.1 Provider selector on checkout

In `web/dashboard/src/components/bundle/`:

- `BundleCheckout.tsx` — adds provider radio selector before Stripe redirect
- Only shows providers with `status = 'active'` from `/v1/providers`
- If no provider connected: shows "Connect a provider" CTA instead of checkout button
- Selected `provider_id` passed to checkout session creation endpoint

### 4.2 Deployment status banner

In `web/dashboard/src/pages/apps/`:

- Shows deploy status for bundle subscriptions: `deploying` (spinner), `deployed` (green check), `failed` (red + error message + Retry button), `awaiting_provider` (yellow + connect CTA)
- Polls subscription status every 5s while `deploying`

### 4.3 Checkout session creation

In `internal/api/handlers/billing/bundles.go`, modify `HandleCreateCheckoutSession`:

```go
// Add to Stripe session metadata:
metadata["provider"] = req.Provider     // "workers", "vercel", etc.
metadata["provider_id"] = req.ProviderID
```

**Verify:** `cd web/dashboard && npx vitest run` passes. Manual test: checkout with CF provider selected, verify Stripe metadata includes provider.

---

## Execution Order

| Step | Files | Depends on |
|------|-------|-----------|
| 1. Migration | `migrations/20260702170000_*.sql` | — |
| 2. Types + storage | `types.go`, `bundle_repository.go` | Step 1 |
| 3. Template adapter | `bundles_templates.go` | Step 2 |
| 4. CF deployment additions | `deployment.go` | — |
| 5. Deploy orchestrator | `bundles_deploy.go` | Steps 2, 3, 4 |
| 6. Webhook integration | `stripe.go` | Step 5 |
| 7. Retry ticker | `bundles_retry.go` | Step 5 |
| 8. Provider-connect hook | `connect.go` | Step 5 |
| 9. Dashboard provider selector | `BundleCheckout.tsx` | Step 6 |
| 10. Dashboard status banner | `apps/` | Step 6 |

Steps 3, 4, 7, 8 can be parallelized. Steps 9, 10 can be parallelized.
