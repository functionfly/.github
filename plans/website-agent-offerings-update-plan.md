# FunctionFly Website Update Plan: Advertising AI Agent Offerings

## Executive Summary

This plan outlines updates needed to FunctionFly's website to properly advertise revenue-generating AI agent offerings that currently aren't being highlighted.

---

## Revenue-Generating Offerings Identified

### 1. Agent Execution Plans (AEP) - Primary AI Agent Revenue Stream

| Plan | Price | Calls/Month | Concurrency | Burst | Daily Spend Cap |
|------|-------|-------------|-------------|-------|-----------------|
| Agent Starter | $49/mo | 500K | 10 | 50 | $5 |
| Agent Scale | $299/mo | 5M | 100 | 500 | $30 |
| Agent Pro | $999/mo | 25M | 500 | 2000 | $100 |
| Agent Enterprise | $2500+/mo | Unlimited | Unlimited | Unlimited | Custom |

**Key Agent Features:**
- Per-agent cost attribution and billing
- Spend caps (daily/monthly)
- Burst concurrency handling
- Multi-agent coordination
- State writes per hour (1K - 50K+)
- Memory storage (10GB - 500GB+)
- Log retention (30 - 365 days)

### 2. StateFabric - Stateful Layer for AI Agents

| Plan | Price | State Objects | Ops/Month | Features |
|------|-------|---------------|------------|----------|
| Sandbox | Free | 1 | 10K | Basic |
| Starter | $19/mo | 5 | 100K | Dev state |
| Pro | $99/mo | 50 | 1M | Hot cache, replay |
| Business | $499/mo | 500 | 10M | Multi-region |
| Enterprise | $1999/mo | Unlimited | Unlimited | Full features |

**Add-ons:**
- AI Memory Pack: $149/mo (vector index, embeddings)
- Hot Cache Booster: $49/mo
- Advanced Security Pack: $99/mo
- Advanced Insights: $79/mo

### 3. Function Registry/Marketplace

- Functions can be priced per call
- Developers can monetize their functions
- Categories: API Tools, Auth, Database, ML, Payment, etc.

### 4. Regular Function Hosting

- Free (100K requests), Starter ($29), Professional ($99), Enterprise (Custom)

---

## Website Gaps Identified

1. **Landing Pages** focus on "multi-cloud failover" but don't mention AI agents
2. **No Agent Execution Plans section** in pricing or main marketing
3. **Function registry** doesn't highlight monetization opportunities
4. **StateFabric** section exists in dashboard pricing but not emphasized for AI agents

---

## Files to Update

### Priority 1: Add Agent Pricing to Dashboard

1. **`web/dashboard/src/lib/constants.ts`**
   - Add AGENT_PLANS constant with all 4 tiers
   - Include: name, price, calls_per_month, concurrency, burst, daily_spend_cap, features

2. **`web/dashboard/src/pages/PricingPage/index.tsx`**
   - Import AGENT_PLANS
   - Add "Agent Execution Plans" section with pricing cards
   - Display alongside existing function hosting plans

### Priority 2: Add Agent Section to Landing Page

3. **`web/dashboard/src/pages/LandingPage/components/PricingSection.tsx`**
   - Update to show Agent plans alongside function plans
   - Or create separate AgentPricingSection component

4. **`web/dashboard/src/pages/LandingPage/components/FeaturesSection.tsx`**
   - Add "Built for AI Agents" feature highlights:
     - Per-agent billing & cost tracking
     - Budget enforcement
     - Burst handling
     - Multi-agent coordination
     - State management for agents

### Priority 3: Update Main Site Pages

5. **`web/site/src/pages/index.astro`**
   - Add section highlighting AI agent capabilities
   - Add Agent Execution Plans to pricing CTA

6. **`web/site/src/pages/pricing.astro`**
   - Add Agent plan tiers

### Priority 4: Function Registry Monetization

7. **`web/dashboard/src/pages/BrowseFunctionsPage/index.tsx`**
   - Add "Monetize Your Functions" banner/CTA
   - Highlight pricing per call feature
   - Add earnings potential indicators

---

## Implementation Order

```
Phase 1: Agent Pricing (High Impact)
├── 1.1 Add AGENT_PLANS to constants.ts
├── 1.2 Add Agent section to PricingPage
└── 1.3 Add Agent features to LandingPage

Phase 2: Site Updates (Medium Impact)
├── 2.1 Update index.astro with agent content
└── 2.2 Update pricing.astro

Phase 3: Registry Monetization (Lower Priority)
└── 3.1 Add monetization CTA to BrowseFunctionsPage
```

---

## Notes

- Agent pricing data is defined in `internal/plans/limits.go` (AgentStarterPriceCents, etc.)
- StateFabric is already in dashboard pricing but needs AI agent messaging
- The function registry already supports `price_per_call` in the API
