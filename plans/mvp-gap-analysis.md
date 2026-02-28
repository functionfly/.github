# FunctionFly MVP Gap Analysis
**Date:** February 28, 2026  
**Launch Target:** March 2, 2026  
**Scope:** Web Dashboard (`web/dashboard/src`)

---

## 🔴 Critical Blockers (Must Fix Before Launch)

### 1. Broken Navigation / Routing

| Issue | File | Line |
|-------|------|------|
| **"Forgot password?" is a dead `href="#"`** — no password reset route or page exists | `pages/AuthPage/LoginForm.tsx` | 233 |
| **Sidebar admin check is wrong** — uses `user.role === 'admin'` but `AdminRoute` checks for `super_admin`, `support`, `billing_admin`, `developer_admin` — admins can't see admin nav | `components/layout/Sidebar.tsx` | 78 |
| **`/state-fabric/new` route doesn't exist** — "Create State Fabric" button navigates there but it's not registered in `App.tsx` | `pages/StateFabricPage/index.tsx` | 108 |
| **"Functions" is missing from the sidebar** — `navigationSections` has Registry, Providers, State Fabric, Analytics, Settings — but no Functions link | `components/layout/Sidebar.tsx` | 41-68 |

### 2. Core Pages Showing Fake Data

| Issue | File |
|-------|------|
| **DashboardPage is entirely mock data** — stats, provider status, and recent activity are all hardcoded | `pages/DashboardPage/index.tsx` |
| **FunctionsPage is entirely mock data** — 4 hardcoded functions, no real API call | `pages/FunctionsPage/index.tsx` |

### 3. Missing Critical Pages

| Issue |
|-------|
| **No password reset page** — `PasswordResetFlow.tsx` component exists but no route is registered |
| **No 404 page** — all unmatched routes silently redirect to `/` |

### 4. Production Readiness

| Issue | File | Line |
|-------|------|------|
| **Excessive `console.log` in routing guards** — logs auth state on every render | `App.tsx` | 81-110 |

---

## 🟠 High Priority (Should Fix Before Launch)

### Mock/Hardcoded Data

| Issue | File | Line |
|-------|------|------|
| **AnalyticsPage uses simulated WebSocket** — fake `setInterval` generates random data | `pages/AnalyticsPage/index.tsx` | 83-124 |
| **SettingsPage profile form has hardcoded values** — "John Doe" / "john@example.com" | `pages/SettingsPage/index.tsx` | 55-65 |
| **SettingsPage billing shows hardcoded usage** — "8,234 / 10,000 requests" is static | `pages/SettingsPage/index.tsx` | 118 |
| **SettingsPage API Keys shows hardcoded key** — `ff_live_••••••••••••` is fake | `pages/SettingsPage/index.tsx` | 159 |
| **FunctionEditorPage uses hardcoded providers** — `mockProviders` instead of API | `pages/FunctionsPage/FunctionEditorPage.tsx` | 49-53 |
| **FunctionDetailPage redeploy is simulated** — `setTimeout` mock instead of real API | `pages/FunctionsPage/FunctionDetailPage.tsx` | 153-160 |

### Missing UI Elements on Public Pages

| Issue | File |
|-------|------|
| **FeedbackPage missing Navbar** — no top navigation bar | `pages/FeedbackPage/index.tsx` |
| **BlogPostPage missing Footer** — BlogPage has Footer but BlogPostPage does not | `pages/BlogPostPage/index.tsx` |

---

## 🟡 Medium Priority (Fix If Time Allows)

### UX Gaps

| Issue | File | Line |
|-------|------|------|
| **SettingsPage notifications use raw HTML checkbox** — should use `Switch` component | `pages/SettingsPage/index.tsx` | 198-201 |
| **SettingsPage "Upgrade to Pro" button has no action** — dead button | `pages/SettingsPage/index.tsx` | 136 |
| **FunctionEditorPage "Save Draft" does nothing** — only logs to console | `pages/FunctionsPage/FunctionEditorPage.tsx` | 266-268 |
| **RegistryDeployPage "Go to Dashboard" UX gap** — no guidance on creating an app | `pages/RegistryDeployPage/index.tsx` | 355 |
| **MFA setup not linked from Settings** — `MFASetup.tsx` exists but unreachable | `pages/SettingsPage/index.tsx` |
| **FeedbackPage API endpoint mismatch** — submits to `/api/feedback` vs `/v1/` prefix | `pages/FeedbackPage/index.tsx` | 231 |

### Missing Pages

| Issue |
|-------|
| **No Terms of Service page** — pricing/signup pages reference it but `/terms` route doesn't exist |

---

## 🟢 Low Priority / Post-Launch

| Issue | File |
|-------|------|
| **Sidebar "Recent" items are hardcoded** — always shows Registry + Dashboard | `components/layout/Sidebar.tsx` | 93-96 |
| **TeamPage uses placeholder images** — `/api/placeholder/200/200` for all team photos | `pages/TeamPage/index.tsx` | 12 |

---

## ✅ Items to Verify Before Launch

| Item | Component |
|------|-----------|
| OAuth callback flow works end-to-end | `pages/AuthPage/OAuthCallback.tsx` |
| Email verification flow works end-to-end | `pages/AuthPage/VerifyEmailPage.tsx` |
| `StateFabricDetailPage` is fully implemented | `pages/StateFabricPage/StateFabricDetailPage.tsx` |
| `AdminFunctionsPage` and `AdminRegistryPage` are implemented | `pages/AdminFunctionsPage/`, `pages/AdminRegistryPage/` |
| `DocsPage` content is real, not placeholder | `pages/DocsPage/` |
| `ChangelogPage` has real or meaningful data | `pages/ChangelogPage/` |

---

## Summary by Priority

```
🔴 Critical (4 routing bugs + 2 mock data pages + 2 missing pages + console.logs) = ~9 items
🟠 High (6 mock data + 2 missing UI elements) = ~8 items  
🟡 Medium (6 UX gaps + 1 missing page) = ~7 items
🟢 Low (2 cosmetic issues) = ~2 items
✅ Verify (6 items) = ~6 items
```

**Recommended focus for March 2nd launch:**
1. Fix all 🔴 Critical items first
2. Fix 🟠 High priority SettingsPage real data + public page Navbar/Footer
3. Verify the ✅ items work end-to-end
4. Defer 🟡 Medium and 🟢 Low to post-launch
