# Marketing Site Launch Readiness Review

**Date:** 2026-03-26
**Scope:** Marketing site (`web/site/`) - platform pages, pricing, trust layer

---

## Executive Summary

The marketing site is **mostly production-ready** with solid content and proper SEO infrastructure. However, there are **3 critical issues** that need resolution before launch:

1. **[CRITICAL] SLA page is a placeholder** - Pricing advertises SLAs but SLA page says "Coming soon"
2. **[HIGH] Careers page is a placeholder** - Limited content
3. **[MEDIUM] "The Agora" competitor reference** - Needs verification if real competitor

---

## Page-by-Page Review

### ✅ Homepage ([`index.astro`](web/site/src/pages/index.astro))

| Aspect | Status | Notes |
|--------|--------|-------|
| Content | ✅ Complete | Hero, audience segments, features, code example, CTA |
| SEO | ✅ Complete | Title, description, OG tags, canonical URL |
| Links | ✅ Valid | Dashboard, docs links use config variables |
| CTA | ✅ Working | Points to `APP_DASHBOARD_ORIGIN` |

**Features claimed:**

- Verified Tool Publishing ✅ (Trust API routes exist)
- Verification Levels & Revocation ✅
- Execution Trace & Trust Scores ✅ (handler: `HandleGetTrustScore`)
- Agent-ready Tooling ✅
- Marketplace for Verified Tools ⚠️ (check registry implementation)
- Trust-powered Routing ✅

---

### ✅ Pricing ([`pricing.astro`](web/site/src/pages/pricing.astro:1))

| Plan Tier | Price | Status |
|-----------|-------|--------|
| **Platform - Free** | $0/mo | ✅ |
| **Platform - Starter** | $29/mo | ✅ |
| **Platform - Professional** | $99/mo | ✅ (99.9% SLA advertised) |
| **Platform - Enterprise** | Custom | ✅ (99.99% SLA advertised) |
| **State Fabric - Sandbox** | $0/mo | ✅ |
| **State Fabric - Starter** | $19/mo | ✅ |
| **State Fabric - Pro** | $99/mo | ✅ |
| **State Fabric - Business** | $499/mo | ✅ |
| **State Fabric - Enterprise** | Custom | ✅ |
| **Agent Starter** | $49/mo | ✅ |
| **Agent Scale** | $299/mo | ✅ |
| **Agent Pro** | $999/mo | ✅ |
| **Agent Enterprise** | Custom | ✅ |

**Add-ons listed:**

- Hot Cache Booster ($49/mo per 5GB) ✅
- Advanced Security Pack ($99/mo) ✅
- AI Memory Pack ($149/mo) ✅
- Advanced Insights ($79/mo) ✅

**Trust API billing:** Mentioned as usage-based, links to Trust page ✅

---

### ✅ For Agents ([`for-agents.astro`](web/site/src/pages/for-agents.astro:1))

| Section | Status | Implementation Match |
|---------|--------|---------------------|
| Discovery & manifests | ✅ | ✅ Handler: `discovery.go` |
| Trust API | ✅ | ✅ Routes in [`routes_trustapi.go`](internal/api/routes_trustapi.go:56) |
| Verification flows | ✅ | ✅ `HandleSubmitVerification` exists |
| Agent execution plans | ✅ | ✅ Table with 4 tiers |
| State Fabric for agents | ✅ | ✅ Links to pricing section |

**Trust API endpoints documented:**

- `GET /v1/trust/score/{function_id}` ✅
- `POST /v1/trust/batch` ✅
- `GET /v1/trust/history/{function_id}` ✅
- `POST /v1/trust/verify` ✅

---

### ✅ Trust Page ([`trust.astro`](web/site/src/pages/trust.astro:1))

| Feature | Status | Implementation |
|---------|--------|----------------|
| Verification levels (L1-L4) | ✅ | Trust score components in backend |
| Signing & attestations | ✅ | Attestation records in storage |
| Revocation | ✅ | Status tracking |
| Zero-knowledge vault | ✅ | Client-side encryption |
| Comparison table | ⚠️ | See note below |

**⚠️ Comparison Table Competitors:**

| Competitor | Status |
|-------------|--------|
| RapidAPI | Real company ✅ |
| Toolhouse | Real company ✅ |
| **The Agora** | **Unverified** - need to confirm this is a real competitor or placeholder |

---

### ✅ Layout ([`Layout.astro`](web/site/src/layouts/Layout.astro:1))

| Aspect | Status |
|--------|--------|
| SEO meta tags | ✅ Description, canonical |
| Open Graph | ✅ type, url, title, description, image, site_name |
| Twitter Cards | ✅ summary_large_image |
| Fonts | ✅ Inter, JetBrains Mono |
| Navigation | ✅ Sticky header, active states |
| Footer | ✅ Brand, links, copyright |

---

### ⚠️ SLA Page ([`sla.astro`](web/site/src/pages/sla.astro:1))

**Status: ❌ PLACEHOLDER - NOT PRODUCTION READY**

```
Availability commitments: Coming soon.
Service credits and remedies: Coming soon.
```

**Issue:** Pricing page advertises:

- Professional: 99.9% SLA
- Enterprise: 99.99% SLA

But the SLA page has no actual commitments. This is a **legal and compliance risk** for enterprise customers.

---

### ⚠️ Careers Page ([`careers.astro`](web/site/src/pages/careers.astro:1))

**Status: ⚠️ MINIMAL PLACEHOLDER**

Content appears to be a single line or very minimal. If you're actively hiring, this needs expansion.

---

### ✅ Privacy Policy ([`privacy.astro`](web/site/src/pages/privacy.astro:1))

Status: ✅ Appears complete (21,407 chars). Covers GDPR, CCPA, data handling.

---

### ✅ Terms of Service ([`terms.astro`](web/site/src/pages/terms.astro:1))

Status: ✅ Appears complete (20,156 chars). Standard terms.

---

### ✅ Contact Page ([`contact.astro`](web/site/src/pages/contact.astro:1))

Status: ✅ Basic contact form/page.

---

### ✅ About Page ([`about.astro`](web/site/src/pages/about.astro:1))

Status: ✅ Basic about page.

---

### ✅ Changelog ([`changelog.astro`](web/site/src/pages/changelog.astro:1))

Status: ✅ Present, links from footer.

---

### ✅ Blog Index ([`blog/index.astro`](web/site/src/pages/blog/index.astro))

Status: ✅ Fetches from blog API.

---

### ✅ Blog Post ([`blog/[slug].astro`](web/site/src/pages/blog/[slug].astro))

Status: ✅ Dynamic slug routing.

---

## SEO & Technical Review

| Item | Status | Notes |
|------|--------|-------|
| Sitemap | ✅ | Generated via `@astrojs/sitemap` |
| robots.txt | ✅ | Allows all, references sitemap |
| Canonical URLs | ✅ | Dynamic via `SITE_ORIGIN` config |
| OG Images | ✅ | `/og-default.svg` exists |
| Favicon | ✅ | SVG favicon present |
| Apple Touch Icon | ✅ | Present |

---

## Platform Implementation Verification

The following Trust API endpoints mentioned in marketing materials **exist in backend**:

| Endpoint (Marketing) | Handler | Status |
|---------------------|---------|--------|
| `GET /v1/trust/score/{function_id}` | `HandleGetTrustScore` | ✅ [`trust.go:18`](internal/api/handlers/trustapi/trust.go:18) |
| `POST /v1/trust/batch` | `HandleBatchTrustScore` | ✅ [`trust.go:78`](internal/api/handlers/trustapi/trust.go:78) |
| `GET /v1/trust/history/{function_id}` | `HandleGetTrustHistory` | ✅ [`trust.go:157`](internal/api/handlers/trustapi/trust.go:157) |
| `POST /v1/trust/verify` | `HandleSubmitVerification` | ✅ [`trust.go:218`](internal/api/handlers/trustapi/trust.go:218) |
| Trust Score Scheduler | `TrustScoreScheduler` | ✅ [`routes.go:306`](internal/api/routes.go:306) |

---

## Critical Issues for Launch

### 1. [CRITICAL] SLA Page is Placeholder

**File:** [`web/site/src/pages/sla.astro`](web/site/src/pages/sla.astro:28)

**Problem:**

- Pricing advertises 99.9% SLA (Professional) and 99.99% SLA (Enterprise)
- SLA page says "Coming soon" for actual commitments

**Required Action:**

- Either remove SLA claims from pricing page until SLA is finalized
- OR complete the SLA page with actual terms before launch
- Consider: credit structure, exclusions, remedy process

---

### 2. [HIGH] Careers Page Minimal Content

**File:** [`web/site/src/pages/careers.astro`](web/site/src/pages/careers.astro:1)

**Problem:** Page appears to be placeholder with minimal content.

**Required Action:**

- Expand with actual job listings if hiring
- Or remove careers link from footer if not actively hiring

---

### 3. [MEDIUM] Unverified Competitor Reference

**File:** [`web/site/src/pages/trust.astro`](web/site/src/pages/trust.astro:220)

**Problem:** "The Agora" in comparison table - is this a real competitor?

**Required Action:**

- Verify "The Agora" exists and claims are accurate
- If not a real competitor, remove or replace with actual market competitor

---

## Recommendations

### Must Fix Before Launch

1. Complete or remove SLA claims from pricing
2. Expand or remove careers page

### Should Fix Before Launch

1. Verify "The Agora" competitor reference

### Nice to Have

- Add structured data (JSON-LD) for Organization schema
- Add breadcrumb navigation for better SEO
- Consider adding customer logos/testimonials if available

---

---

# Dashboard Launch Readiness Review

**Date:** 2026-03-26
**Scope:** Dashboard SPA (`web/dashboard/src/`)

## Executive Summary

The dashboard is **~80% launch-ready** for authenticated users. Core functionality is implemented including:

- Main dashboard with metrics
- Functions management
- Agents management
- Pricing page with Stripe integration
- Enterprise SLA page (fully functional API)

**Critical Gaps:**

1. Enterprise Support page has no backend - buttons don't do anything
2. Enterprise Audit page has placeholder data only
3. SLA page on marketing site is placeholder (conflicts with dashboard SLA claims)

---

## Dashboard Page Review

### ✅ Landing Page (Unauthenticated)

**File:** [`web/dashboard/src/pages/LandingPage/index.tsx`](web/dashboard/src/pages/LandingPage/index.tsx:1)

**Status:** Legacy page - redirects logged-out users to `web/site` Astro marketing site.

**Note:** The dashboard's landing page is not actually used in production since unauthenticated users are redirected to the marketing Astro site.

---

### ✅ Pricing Page

**File:** [`web/dashboard/src/pages/PricingPage/index.tsx`](web/dashboard/src/pages/PricingPage/index.tsx:1)

| Component | Status |
|-----------|--------|
| Platform Plans (Free/Starter/Pro/Enterprise) | ✅ Matches marketing site |
| State Fabric Plans (Sandbox/Starter/Pro/Business/Enterprise) | ✅ Matches marketing site |
| Agent Plans (Starter/Scale/Pro/Enterprise) | ✅ Matches marketing site |
| Stripe Checkout Integration | ✅ `createCheckoutSession` |
| Structured Data | ✅ `PricingPageStructuredData` |
| SEO Meta Tags | ✅ `MetaTags` component |

**Plan Constants Verified:** [`lib/constants.ts`](web/dashboard/src/lib/constants.ts:248)

- All plan prices match marketing site
- Stripe price IDs use environment variables with fallback placeholders

---

### ✅ Main Dashboard (Authenticated)

**File:** [`web/dashboard/src/pages/DashboardPage/index.tsx`](web/dashboard/src/pages/DashboardPage/index.tsx:1)

| Widget | Data Source | Status |
|--------|-------------|--------|
| Functions list | `functionsApi.list()` | ✅ |
| Providers status | `providersApi.getConnectedProviders()` | ✅ |
| Apps list | `appsApi.list()` | ✅ |
| Usage graph | `dashboardApi.getUsage(14)` | ✅ |
| Execution rate | `dashboardApi.getExecutionRate(24)` | ✅ |
| Activity feed | `dashboardApi.getActivity(20)` | ✅ |
| Memory usage | `dashboardApi.getMemoryUsage()` | ✅ |
| Metrics | `dashboardApi.getMetrics()` | ✅ |
| Health status | `dashboardApi.getHealthStatus()` | ✅ |
| Onboarding resume | `useOnboardingStore()` | ✅ |
| Enterprise status | `EnterpriseStatusCard` | ✅ |
| Trust score badge | `TrustScoreBadge` | ✅ |

---

### ✅ Enterprise SLA Page

**File:** [`web/dashboard/src/pages/EnterpriseSLAPage/index.tsx`](web/dashboard/src/pages/EnterpriseSLAPage/index.tsx:1)

| Feature | Implementation |
|---------|----------------|
| SLA Overview Cards | ✅ API: `enterpriseSlaApi.getOverview()` |
| Uptime History Chart | ✅ API: `enterpriseSlaApi.getUptimeHistory()` |
| Incident History | ✅ API: `enterpriseSlaApi.getIncidents()` |
| Enterprise Gate | ✅ Redirects non-Enterprise users |

**API Verified:** [`enterprise.ts`](web/dashboard/src/api/enterprise.ts:37)

---

### ⚠️ Enterprise Support Page

**File:** [`web/dashboard/src/pages/EnterpriseSupportPage/index.tsx`](web/dashboard/src/pages/EnterpriseSupportPage/index.tsx:1)

**Status: ⚠️ UI PRESENT - NO BACKEND**

| Component | Issue |
|-----------|-------|
| Live Chat button | ❌ No API, does nothing |
| Email Support button | ❌ No API, does nothing |
| Phone Support button | ❌ No API, does nothing |
| Account Manager card | ❌ UI only, no integration |
| Contact Support button | ❌ Does nothing |
| Schedule Call button | ❌ Does nothing |
| Support SLA grid | ❌ Hardcoded values, UI only |

**Required:** Implement support ticket system or integrate with helpdesk (Zendesk, Intercom, etc.)

---

### ⚠️ Enterprise Audit Page

**File:** [`web/dashboard/src/pages/EnterpriseAuditPage/index.tsx`](web/dashboard/src/pages/EnterpriseAuditPage/index.tsx:1)

**Status: ⚠️ PLACEHOLDER DATA - NO BACKEND**

| Component | Issue |
|-----------|-------|
| Search input | ❌ UI only, no functionality |
| Filter button | ❌ Does nothing |
| Export button | ❌ Does nothing |
| Audit log table | ❌ Hardcoded placeholder data |
| Empty state | ⚠️ Shows "Audit log integration coming soon" |

**Required:** Implement audit logging backend or integrate with logging service (Datadog, Splunk, etc.)

---

## Cross-Cutting Concerns

### ✅ Stripe Integration

- Pricing page uses `createCheckoutSession` for paid plans
- Success/cancel URLs properly configured
- Environment variable fallbacks for Stripe price IDs

### ✅ SEO

- MetaTags component used on Landing and Pricing pages
- StructuredData components for rich snippets

### ⚠️ Environment Variables

```typescript
// Placeholder fallbacks in constants.ts
VITE_STRIPE_PRICE_STARTER || 'price_starter_placeholder'
VITE_STRIPE_PRICE_PROFESSIONAL || 'price_professional_placeholder'
VITE_STRIPE_PRICE_AGENT_STARTER || 'price_agent_starter_placeholder'
// etc.
```

**Action Required:** Set actual Stripe price IDs in production environment

---

## Consolidated Launch Readiness Report

### Marketing Site (`web/site/`) - ~85% Ready

| Page | Status | Notes |
|------|--------|-------|
| Homepage | ✅ | Complete |
| Pricing | ✅ | Matches dashboard constants |
| For Agents | ✅ | Trust API documented |
| Trust | ✅ | Implementation verified |
| SLA | ❌ | **Placeholder "Coming soon"** |
| Careers | ⚠️ | Minimal content |
| Privacy | ✅ | Complete |
| Terms | ✅ | Complete |

### Dashboard (`web/dashboard/`) - ~80% Ready

| Page | Status | Notes |
|------|--------|-------|
| Landing (unauthenticated) | ✅ | Redirects to Astro site |
| Pricing | ✅ | Full implementation + Stripe |
| Dashboard (authenticated) | ✅ | All widgets with API |
| Functions | ✅ | Full CRUD operations |
| Agents | ✅ | Agent SDK integrations |
| Enterprise SLA | ✅ | **Fully functional API** |
| Enterprise Support | ⚠️ | UI only, no backend |
| Enterprise Audit | ⚠️ | Placeholder data only |

---

## Priority Issues for Launch

### P0 - Must Fix Before Launch

| Issue | Location | Fix Required |
|-------|----------|--------------|
| Marketing SLA page placeholder | `web/site/src/pages/sla.astro` | Either complete SLA terms OR remove SLA claims from pricing |
| Enterprise Support no backend | `web/dashboard/src/pages/EnterpriseSupportPage/` | Implement ticket system or remove CTAs |
| Enterprise Audit no backend | `web/dashboard/src/pages/EnterpriseAuditPage/` | Implement audit logging or show "Coming soon" properly |

### P1 - Should Fix Before Launch

| Issue | Location | Fix Required |
|-------|----------|--------------|
| Stripe price ID placeholders | `web/dashboard/src/lib/constants.ts` | Set real `VITE_STRIPE_PRICE_*` values |
| Careers page minimal | `web/site/src/pages/careers.astro` | Expand or remove link |
| "The Agora" competitor | `web/site/src/pages/trust.astro` | Verify this is a real competitor |

### P2 - Nice to Have

| Issue | Location | Fix Required |
|-------|----------|--------------|
| Hotjar analytics | `web/dashboard/src/pages/LandingPage/index.tsx:75` | Remove console.log or implement properly |
| Missing structured data | Dashboard pages | Add Schema.org markup |

---

## Conclusion

**Overall: ~75-80% Launch Ready**

The platform is substantially complete with working Trust API, pricing, dashboard, and core functionality. The main concerns are:

1. **Marketing-Implementation Mismatch**: SLA page says "Coming soon" on marketing site while dashboard SLA page is fully functional
2. **Enterprise Features**: Support and Audit pages have UI but no backend
3. **Placeholder Content**: Careers page, SLA page need attention

**Recommended Launch Gate:**

1. Resolve SLA page conflict (marketing vs dashboard)
2. Implement or disable Enterprise Support CTAs
3. Set real Stripe price IDs in environment

**Estimated completion for launch:** 2-3 days of focused work on the P0 issues.
