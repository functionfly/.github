# Stripe Signup & Upgrade Flow - Phase 2 Implementation Plan

## Overview

**Phase 1 (Completed):** Fixes #3 (Error Recovery), #4 (URL Validation), #6 (Subscription Refresh), #9 (Invoice Links)

**Phase 2 (This Plan):** Fixes #1, #2, #7, #11 - Loading States, Trial UI, Plan Display, Usage Visualization

---

## Fix #1: Loading State During Stripe Customer Creation (HIGH PRIORITY)

### Problem
When `CreateOrGetStripeCustomer` is called during billing portal/checkout:
- Takes several seconds if Stripe API is slow
- No loading indicator on PricingPage checkout
- User might click multiple times causing potential duplicate requests
- BillingSettingsTab already has `billingPortalLoading` state working correctly

### Files to Modify
1. **`web/dashboard/src/pages/PricingPage/index.tsx`** - Add checkout initiation loading state
2. **`web/dashboard/src/pages/PricingPage/components/FunctionPlanCard.tsx`** - Accept and display loading prop

### Implementation Details

#### In `PricingPage/index.tsx`:

Add state near existing state declarations (~line 37):
```typescript
const [checkoutInitiating, setCheckoutInitiating] = useState(false);
```

Wrap `handlePlanSelect` with loading state:
```typescript
const handlePlanSelect = async (planId: string, priceId?: string) => {
  // ... existing validation for enterprise, free plans ...
  
  setCheckoutInitiating(true);  // NEW
  setShowConfetti(true);
  
  try {
    const { url } = await createCheckoutSession(priceId, successUrl, cancelUrl);
    window.location.href = url;
  } catch (error) {
    // ... existing error dialog logic ...
  } finally {
    setCheckoutInitiating(false);  // NEW - only reached if error
  }
};
```

Pass `checkoutInitiating` and selected plan to FunctionPlanCards:
```tsx
<FunctionPlanCard
  key={plan.id}
  plan={plan}
  index={index}
  onPlanSelect={handlePlanSelect}
  disabled={checkoutInitiating}  // NEW prop
  isLoading={checkoutInitiating}  // NEW prop
/>
```

#### In `FunctionPlanCard.tsx`:
Add `disabled` and `isLoading` props, show spinner when initiating:
```tsx
<Button
  onClick={() => !disabled && !isLoading && onPlanSelect(plan.id, plan.priceId)}
  disabled={disabled || isLoading}
>
  {isLoading ? (
    <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Starting Checkout...</>
  ) : (
    buttonText
  )}
</Button>
```

---

## Fix #2: Trial Period UI Elements (MEDIUM PRIORITY)

### Problem
Subscription model has `TrialEnd *time.Time` field (`models_billing.go:31`) but no UI shows trial status.

### Files to Modify
1. **`internal/api/handlers/billing/handler.go`** - Add trial fields to SubscriptionResponse
2. **`web/dashboard/src/api/billing.ts`** - Add trial fields to Subscription interface
3. **`web/dashboard/src/pages/SettingsPage/components/BillingSettingsTab.tsx`** - Display trial info

### Implementation Details

#### Backend (`handler.go`):

Add to SubscriptionResponse struct:
```go
type SubscriptionResponse struct {
  // ... existing fields ...
  TrialEnd             *time.Time `json:"trial_end,omitempty"`        // NEW
  IsTrialing           bool       `json:"is_trialing"`               // NEW  
  TrialDaysRemaining  int        `json:"trial_days_remaining"`      // NEW
}
```

Add computation after building response object:
```go
var isTrialing bool = false
var daysRemaining int = 0

if subscription.TrialEnd != nil {
    now := time.Now()
    isTrialing = now.Before(*subscription.TrialEnd)
    diff := subscription.TrialEnd.Sub(now)
    daysRemaining = int(diff.Hours() / 24)
    if daysRemaining < 0 {
        daysRemaining = 0
    }
}

response.TrialEnd = subscription.TrialEnd
response.IsTrialing = isTrialing
response.TrialDaysRemaining = daysRemaining
```

#### Frontend API (`billing.ts`):

Add to Subscription interface:
```typescript
export interface Subscription {
  // ... existing fields ...
  trial_end: string | null;        // NEW
  is_trialing: boolean;            // NEW
  trial_days_remaining: number;   // NEW
}
```

#### BillingSettingsTab UI:

Add after status badges section (~line 297). Import Clock from lucide-react:
```tsx
{subscription.is_trialing && (
  <div className={`p-4 rounded-lg border ${
    subscription.trial_days_remaining <= 3 
      ? 'bg-amber-500/10 border-amber-500/20' 
      : 'bg-blue-500/10 border-blue-500/20'
  }`}>
    <div className="flex items-center gap-2">
      <Clock className="w-5 h-5 text-text-muted" />
      <span className="text-sm font-medium">Trial Period</span>
    </div>
    <p className="text-sm mt-1">
      <span className={
        subscription.trial_days_remaining <= 3 ? 'text-amber-400' : 'text-blue-400'
      }>
        {subscription.trial_days_remaining} days remaining
      </span>
      {' · '}Ends {formatDate(subscription.trial_end)}
    </p>
    {subscription.trial_days_remaining <= 3 && (
      <p className="text-xs mt-2 text-amber-400">
        Your trial ends soon. Choose a plan to continue using premium features.
      </p>
    )}
  </div>
)}
```

---

## Fix #7: Show Actual Plan for Non-Subscribed Users (MEDIUM PRIORITY)

### Problem
Non-subscribed users see minimal `{displayPlan}` text with no context about what their plan includes.

### Files to Modify
1. **`web/dashboard/src/pages/SettingsPage/components/BillingSettingsTab.tsx`** - Enhance free-tier display section

### Implementation Details

Replace the simple non-subscription display block (~lines 327-336) with enhanced version:

```tsx
{!subscription && (
  <div className="space-y-4">
    {/* Enhanced Plan Card */}
    <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-[#6366f1]/10 to-[#8b5cf6]/10 border border-border-default">
      <div>
        <h3 className="font-semibold text-text-primary capitalize">{displayPlan} Plan</h3>
        <p className="text-sm text-text-secondary mt-1">
          {displayPlan === 'free' || displayPlan.toLowerCase() === 'free' ? (
            <span><Badge variant="secondary" className="mr-2">Free Forever</Badge>Basic features included</span>
          ) : (
            <span><Badge variant="secondary" className="mr-2">{displayPlan}</Badge>Active</span>
          )}
        </p>
      </div>
      <Badge>Current</Badge>
    </div>

    {/* Free tier features list */}
    {(displayPlan === 'free' || displayPlan.toLowerCase() === 'free') && (
      <>
        <div className="p-4 rounded-lg bg-bg-secondary border border-border-default">
          <p className="font-medium text-text-primary mb-2">Your Free Plan includes:</p>
          <ul className="space-y-1 text-sm text-text-secondary">
            <li className="flex items-center gap-2"><Check className="w-4 h-4 text-green-500" />Basic function deployment</li>
            <li className="flex items-center gap-2"><Check className="w-4 h-4 text-green-500" />Community support</li>
            <li className="flex items-center gap-2"><Check className="w-4 h-4 text-green-500" />Registry access</li>
            <li className="flex items-center gap-2"><Check className="w-4 h-4 text-green-500" />Up to 5 functions</li>
          </ul>
        </div>

        {/* Upgrade prompt */}
        <div className="p-4 rounded-lg bg-gradient-to-r from-indigo-500/10 to-purple-500/10 border border-indigo-500/20">
          <p className="text-sm font-medium text-text-primary mb-2">Ready to unlock more?</p>
          <ul className="space-y-1 text-xs text-text-secondary mb-3">
            <li>- Unlimited executions</li>
            <li>- Priority support</li>
            <li>- Advanced analytics</li>
          </ul>
          <Button size="sm" onClick={() => openPortal(`${window.location.origin}/pricing`)}>
            View Plans & Pricing
          </Button>
        </div>
      </>
    )}
  </div>
)}
```

Ensure Check icon is imported from lucide-react (may already exist).

---

## Fix #11: Usage Visualization (MEDIUM PRIORITY)

### Problem
`GET /v1/billing/usage` endpoint exists but has no UI. Users can't see their consumption.

### Files to Modify
1. **`web/dashboard/src/api/billing.ts`** - Add proper UsageSummary type
2. **`web/dashboard/src/pages/SettingsPage/components/BillingSettingsTab.tsx`** - Add Usage card section

### Implementation Details

#### API Types (`billing.ts`):

Add after existing types:
```typescript
export interface UsageDataPoint {
  event_type: string;
  quantity: number;
  unit_price_cents: number;
  total_cost_cents: number;
  timestamp: string;
}

export interface UsageSummary {
  start: string;
  end: string;
  total_events: number;
  total_cost_usd: number;
  events: UsageDataPoint[];
}
```

#### BillingSettingsTab:

Add imports (BarChart3, TrendingUp from lucide-react):

Add usage query after wallet query:
```typescript
const { data: usageData, isLoading: usageLoading } = useQuery({
  queryKey: ['billing', 'usage'],
  queryFn: () => getUsage(),
  enabled: !!user && !!subscription && subscription.status === 'active',
  staleTime: 60_000,
  retry: false,
});
```

Add Usage Card between Wallet card and Invoice card:
```tsx
{subscription && subscription.status === 'active' && (
  <Card>
    <CardHeader>
      <CardTitle className="flex items-center gap-2">
        <BarChart3 className="h-5 w-5 text-[#6366f1]" />
        Usage (Last 30 Days)
      </CardTitle>
      <CardDescription>Your recent activity and resource consumption</CardDescription>
    </CardHeader>
    <CardContent>
      {usageLoading ? (
        <div className="flex items-center justify-center gap-2 p-4 text-text-muted">
          <Loader2 className="h-5 w-5 animate-spin" /><span>Loading usage...</span>
        </div>
      ) : usageData ? (
        <div className="space-y-4">
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 rounded-lg bg-bg-secondary border border-border-default p-4">
            <div>
              <p className="text-xs text-text-muted uppercase">Events</p>
              <p className="text-lg font-semibold tabular-nums">{(usageData.total_events ?? 0).toLocaleString()}</p>
            </div>
            <div>
              <p className="text-xs text-text-muted uppercase">Cost</p>
              <p className="text-lg font-medium tabular-nums">${((usageData.total_cost_usd ?? 0).toFixed(2)}</p>
            </div>
            <div>
              <p className="text-xs text-text-muted uppercase">Period</p>
              <p className="text-sm font-medium">{new Date(usageData.start).toLocaleDateString()} - {new Date(usageData.end).toLocaleDateString()}</p>
            </div>
          </div>
          
          {usageData.events?.length > 0 && (
            <div className="space-y-1">
              <p className="text-sm text-text-muted font-medium mb-2">Breakdown</p>
              {usageData.events.slice(0, 5).map((evt, i) => (
                <div key={i} className="flex justify-between py-2 px-3 rounded bg-bg-tertiary">
                  <span className="text-sm capitalize">{evt.event_type.replace(/_/g, ' ')}</span>
                  <span className="text-sm tabular-nums">{evt.quantity.toLocaleString()} × ${((evt.unit_price_cents ?? 0) / 100)}</span>
                </div>
              ))}
            </div>
          )}

          {!usageData.events?.length && <p className="text-center text-sm py-4 text-text-muted">No usage this period.</p>}
          
          {(displayPlan === 'free' || displayPlan.toLowerCase() === 'free') && (usageData.total_events ?? 0) > 50 && (
            <div className="mt-4 p-4 rounded-lg bg-gradient-to-r from-green-500/10 to-emerald-500/10 border border-green-500/20">
              <div className="flex items-start gap-2">
                <TrendingUp className="w-5 h-5 text-green-500 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-green-400">You're growing!</p>
                  <p className="text-xs text-text-muted mt-1">
                    Your usage suggests a paid plan could save you money.
                    <Button variant="link" size="sm" className="px-0 text-green-400 ml-1" onClick={() => window.location.href = '/pricing'}>Explore plans</Button>
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      ) : (
        <p className="text-center text-sm py-4 text-text-muted">No usage data available</p>
      )}
    </CardContent>
  </Card>
)}
```

---

## Execution Order

| Order | Fix | Effort | Dependencies |
|-------|-----|--------|-------------|
| 1 | #1 Loading States | ~30 min | None |
| 2 | #7 Plan Display | ~30 min | None |
| 3 | #11 Usage Viz | ~45 min | None |
| 4 | #2 Trial UI | ~45 min | Backend + Frontend |

**Total Estimated Time:** ~2.5 hours

## Verification Checklist

After all fixes are implemented:

- [ ] `go build ./cmd/orchestrator-api` passes without errors
- [ ] No TypeScript errors in dashboard build
- [ ] Pricing page shows spinner when clicking paid plan button
- [ ] Multiple clicks during checkout prevented by disabled state
- [ ] Free users in billing settings see features list and upgrade prompt
- [ ] Subscribed users see usage summary with events/cost breakdown
- [ ] Trial users (if applicable) see countdown badge with warning at 3 days
- [ ] Console clean of billing-related errors