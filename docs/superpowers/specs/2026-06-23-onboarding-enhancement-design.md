# Onboarding Flow Enhancement Design

**Date:** 2026-06-23
**Status:** Draft
**Author:** Kilo

## Overview

This design enhances the FunctionFly onboarding flow with missing steps and features, implementing proper plan-based feature gating to ensure users only see features available to their plan tier.

---

## Current State

The current onboarding flow consists of 5 steps:
1. **Welcome** - Overview and feature introduction
2. **Connect Provider** - Cloud provider API token connection
3. **Deploy Function** - Sample function deployment
4. **Test Failover** - Failover simulation testing
5. **Team Setup** - Team member invitations (admin only)

---

## Proposed Enhanced Flow

### Step Order and Plan Gating

| # | Step | Free | Starter | Professional | Enterprise |
|---|------|------|---------|--------------|------------|
| 1 | Welcome | ✅ | ✅ | ✅ | ✅ |
| 2 | Plan Selection | ✅ | ✅ | ✅ | ✅ |
| 3 | Connect Provider | ✅ | ✅ | ✅ | ✅ |
| 4 | Environment Setup | ✅ | ✅ | ✅ | ✅ |
| 5 | Region Selection | ✅ | ✅ | ✅ | ✅ |
| 6 | Deploy Function | ✅ | ✅ | ✅ | ✅ |
| 7 | API Key Intro | ✅ | ✅ | ✅ | ✅ |
| 8 | Test Failover | ✅ | ✅ | ✅ | ✅ |
| 9 | Integrations | ✅ | ✅ | ✅ | ✅ |
| 10 | Custom Domain | ❌ | ✅ (1) | ✅ (5) | ✅ (∞) |
| 11 | Team Setup | ✅ | ✅ | ✅ | ✅ |

### Step Descriptions

#### 1. WelcomeStep (existing)
No changes. Welcome and feature overview.

#### 2. PlanSelectionStep (NEW)
**Purpose:** Allow users to understand and select their plan during onboarding.

**UI Elements:**
- Plan comparison cards (Free, Starter, Professional, Enterprise)
- Current plan highlighted with "Current Plan" badge
- Features list per plan
- Price display with annual/monthly toggle
- "Upgrade" button for users wanting higher tier
- "Continue with [Plan]" button

**Plan Gating:**
- All users see this step
- Free users see upgrade prompts for higher tiers
- Paid users see their current plan pre-selected

**Data Stored:**
```typescript
{
  selectedPlan: 'free' | 'starter' | 'professional' | 'enterprise',
  billingCycle: 'monthly' | 'annual',
  upgradeIntent: boolean
}
```

#### 3. ConnectProviderStep (existing)
No structural changes. Provider selection and API token validation.

#### 4. EnvironmentSetupStep (NEW)
**Purpose:** Create and configure development/staging/production environments.

**UI Elements:**
- Environment cards: Development, Staging, Production
- Environment description and use case
- Default environment selection
- "Create custom environment" option (plan-gated)

**Plan Gating:**
- All users see default 3 environments (dev/staging/prod)
- Custom environments: Starter+ (up to 3), Pro+ (up to 10), Enterprise (unlimited)

**Backend Integration:**
- Uses existing `environmentService` from `api/environment.ts`
- POST to `/v1/users/me/environment` to set active environment

**Data Stored:**
```typescript
{
  environments: ['development', 'staging', 'production'],
  activeEnvironment: 'development',
  customEnvironments: string[]
}
```

#### 5. RegionSelectionStep (NEW)
**Purpose:** Select preferred deployment regions for lower latency.

**UI Elements:**
- World map with region selection
- Region cards with latency indicators
- "Auto-select best regions" option
- Multi-select (up to plan limit)

**Plan Gating:**
- Free: 1 region
- Starter: 3 regions
- Professional: 5 regions
- Enterprise: unlimited

**Backend Integration:**
- Uses existing regions from `constants.ts` PROVIDERS
- POST to `/v1/users/me/regions` to save preferences

**Data Stored:**
```typescript
{
  selectedRegions: string[],
  autoSelect: boolean
}
```

#### 6. DeployFunctionStep (existing)
No structural changes. Sample/custom function deployment.

#### 7. APIKeyIntroStep (NEW)
**Purpose:** Introduce users to API key management and demonstrate usage.

**UI Elements:**
- "Create your first API key" button
- API key name input
- Environment selector for key scope
- Example cURL command display
- Copy-to-clipboard functionality
- Link to API keys management page

**Plan Gating:**
- All users can create API keys (limits per plan from `limits.go`)
- Free: 1 key, Starter: 5, Pro: 20, Enterprise: unlimited

**Backend Integration:**
- Uses existing `api-keys.ts` service
- POST to `/v1/api-keys` to create key

**Data Stored:**
```typescript
{
  createdApiKey: boolean,
  apiKeyName: string,
  apiKeyId: string
}
```

#### 8. TestFailoverStep (existing)
No changes. Failover simulation testing.

#### 9. IntegrationsStep (NEW)
**Purpose:** Connect notification integrations (Slack, Discord) during onboarding.

**UI Elements:**
- Integration cards: Slack, Discord, GitHub, Sentry
- OAuth connect buttons
- Post-install configuration hints
- Skip option

**Plan Gating:**
- All users see Slack, Discord, GitHub (basic)
- Professional+: Sentry, Datadog, New Relic
- Enterprise+: All integrations including custom webhooks

**Backend Integration:**
- OAuth flow for Slack/Discord
- Uses existing webhook infrastructure
- POST to `/v1/integrations` to save connections

**Data Stored:**
```typescript
{
  connectedIntegrations: Array<{
    id: string,
    type: 'slack' | 'discord' | 'github' | 'sentry' | 'datadog',
    connectedAt: string
  }>
}
```

#### 10. CustomDomainStep (NEW - PLAN GATED)
**Purpose:** Add custom domain configuration during onboarding.

**UI Elements:**
- Domain input field
- DNS configuration instructions
- SSL certificate auto-provision info
- Domain validation status

**Plan Gating:**
- **Free:** NOT SHOWN (0 custom domains)
- **Starter:** SHOWN with limit of 1 domain
- **Professional:** SHOWN with limit of 5 domains
- **Enterprise:** SHOWN with unlimited domains

**Backend Integration:**
- Uses existing domain validation API
- POST to `/v1/domains` to register domain

**Data Stored:**
```typescript
{
  customDomain: string | null,
  domainVerified: boolean
}
```

**Upgrade Prompt for Free Users:**
- Show card explaining custom domains feature
- "Upgrade to Starter" button linking to plan selection
- "Learn more" link to documentation

#### 11. TeamSetupStep (existing)
No structural changes. Team invitations for admins.

---

## Implementation Plan

### 1. Update OnboardingStore

```typescript
// New step type union
export type OnboardingStep =
  | "welcome"
  | "plan-selection"
  | "connect-provider"
  | "environment-setup"
  | "region-selection"
  | "deploy-function"
  | "api-key-intro"
  | "test-failover"
  | "integrations"
  | "custom-domain"
  | "team-setup";

// New store state
interface OnboardingState {
  // ... existing fields
  selectedPlan?: PlanTier;
  environments?: Environment[];
  selectedRegions?: string[];
  createdApiKey?: boolean;
  connectedIntegrations?: IntegrationConfig[];
  customDomain?: string;
}
```

### 2. Plan Gating Utility

```typescript
// lib/plan-gating.ts
export function canShowStep(step: OnboardingStep, plan: PlanTier): boolean {
  switch (step) {
    case 'custom-domain':
      return plan !== 'free'; // Custom domains start at Starter
    case 'region-selection':
      return true; // All users can select regions
    // ... other steps
  }
}

export function getStepLimit(step: OnboardingStep, plan: PlanTier): number | Infinity {
  switch (step) {
    case 'region-selection':
      return { free: 1, starter: 3, professional: 5, enterprise: Infinity }[plan];
    case 'custom-domain':
      return { free: 0, starter: 1, professional: 5, enterprise: Infinity }[plan];
    // ...
  }
}
```

### 3. Component Structure

```
web/dashboard/src/pages/OnboardingPage/
├── index.tsx                    # Main onboarding page (updated)
├── WelcomeStep.tsx              # Existing
├── PlanSelectionStep.tsx        # NEW
├── ConnectProviderStep.tsx     # Existing
├── EnvironmentSetupStep.tsx    # NEW
├── RegionSelectionStep.tsx     # NEW
├── DeployFunctionStep.tsx       # Existing
├── APIKeyIntroStep.tsx          # NEW
├── TestFailoverStep.tsx         # Existing
├── IntegrationsStep.tsx         # NEW
├── CustomDomainStep.tsx         # NEW (plan-gated)
├── TeamSetupStep.tsx           # Existing
├── components/
│   ├── PlanCard.tsx
│   ├── EnvironmentCard.tsx
│   ├── RegionSelector.tsx
│   ├── IntegrationCard.tsx
│   └── DomainInput.tsx
└── hooks/
    ├── usePlanGating.ts
    └── useOnboardingSteps.ts
```

### 4. Step Visibility Logic

```typescript
// In OnboardingPage/index.tsx
const getVisibleSteps = (userPlan: PlanTier, isAdmin: boolean): OnboardingStep[] => {
  const allSteps: OnboardingStep[] = [
    'welcome',
    'plan-selection',
    'connect-provider',
    'environment-setup',
    'region-selection',
    'deploy-function',
    'api-key-intro',
    'test-failover',
    'integrations',
    'custom-domain',
    'team-setup'
  ];

  return allSteps.filter(step => canShowStep(step, userPlan));
};
```

---

## Feature Limits Reference

### From `internal/plans/limits.go`

| Feature | Free | Starter | Professional | Enterprise |
|---------|------|---------|--------------|------------|
| Custom Domains | 0 | 1 | 5 | ∞ |
| Regions | 1 | 3 | 5 | ∞ |
| API Keys | 1 | 5 | 20 | ∞ |
| Environments | 3 | 3 + 3 custom | 3 + 10 custom | ∞ |
| Team Members | 1 | 3 | 10 | ∞ |
| Secrets | 25 | 500 | 5000 | 1M |
| Functions | 3 | 5 | 25 | ∞ |

### From `web/dashboard/src/lib/constants.ts`

| Plan | Custom Domains | Regions (providers) |
|------|---------------|---------------------|
| Free | 0 | 2 providers |
| Starter | 1 | 3 providers |
| Professional | 5 | 5 providers |
| Enterprise | ∞ | ∞ providers |

---

## Migration Considerations

1. **Backward Compatibility:** Existing users who completed onboarding should not be affected
2. **Store Version:** Increment onboarding store version for migration
3. **Step Data:** New step data fields should be optional

---

## Testing Requirements

1. Verify Free users do NOT see Custom Domain step
2. Verify Starter users see Custom Domain step with limit of 1
3. Verify Professional users see Custom Domain step with limit of 5
4. Verify Enterprise users see Custom Domain step with unlimited
5. Verify all users see Plan Selection step
6. Verify region limits are enforced per plan
7. Verify environment creation limits per plan

---

## Success Metrics

- Onboarding completion rate improved by 20%
- Plan upgrade conversions from onboarding increased by 15%
- Time-to-first-function decreased by 30%
- Support tickets for "how to add custom domain" decreased by 50%
