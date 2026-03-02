# Enterprise Features Implementation Plan

## Quick Implementation Path

This plan focuses on the highest-impact enterprise features that can be implemented quickly.

---

## 1. Core Utilities & Hooks

### Plan Utilities (`web/dashboard/src/lib/plan-utils.ts`)

```typescript
import { PLANS } from './constants';

export type PlanTier = 'free' | 'starter' | 'professional' | 'enterprise';

export const PLAN_HIERARCHY: Record<PlanTier, number> = {
  free: 0,
  starter: 1,
  professional: 2,
  enterprise: 3,
};

/**
 * Check if user is on enterprise plan
 */
export const isEnterprise = (plan?: string): boolean => 
  plan?.toLowerCase() === 'enterprise';

/**
 * Check if user has at least the specified plan tier
 */
export const hasMinPlan = (userPlan: string, minPlan: PlanTier): boolean => {
  const userTier = PLAN_HIERARCHY[userPlan.toLowerCase() as PlanTier] ?? 0;
  const requiredTier = PLAN_HIERARCHY[minPlan];
  return userTier >= requiredTier;
};

/**
 * Get plan display name with enterprise styling
 */
export const getPlanDisplayName = (plan?: string): string => {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  return PLANS[planKey]?.name || plan || 'Unknown';
};

/**
 * Get plan limits for the user's tier
 */
export const getPlanLimits = (plan?: string) => {
  const planKey = plan?.toUpperCase() as keyof typeof PLANS;
  return PLANS[planKey]?.limits;
};

/**
 * Check if a feature is available for the plan
 */
export const FEATURES = {
  ADVANCED_ANALYTICS: ['professional', 'enterprise'],
  CUSTOM_DASHBOARDS: ['enterprise'],
  AUDIT_LOGS: ['enterprise'],
  SLA_DASHBOARD: ['enterprise'],
  DEDICATED_SUPPORT: ['enterprise'],
  EXPORT_REPORTS: ['enterprise'],
  TEAM_MANAGEMENT: ['professional', 'enterprise'],
  API_ACCESS: ['starter', 'professional', 'enterprise'],
  WEBHOOKS: ['professional', 'enterprise'],
  CUSTOM_DOMAINS: ['starter', 'professional', 'enterprise'],
} as const;

export type FeatureKey = keyof typeof FEATURES;

export const hasFeature = (plan: string | undefined, feature: FeatureKey): boolean => {
  if (!plan) return false;
  const allowedPlans = FEATURES[feature];
  return allowedPlans.includes(plan.toLowerCase() as PlanTier);
};
```

### Plan Hook (`web/dashboard/src/hooks/usePlan.ts`)

```typescript
import { useAuthStore } from '@/stores/authStore';
import { 
  isEnterprise, 
  hasMinPlan, 
  getPlanLimits, 
  hasFeature,
  type PlanTier,
  type FeatureKey 
} from '@/lib/plan-utils';

interface UsePlanReturn {
  plan: string | undefined;
  isEnterprise: boolean;
  isPaid: boolean;
  limits: ReturnType<typeof getPlanLimits>;
  hasFeature: (feature: FeatureKey) => boolean;
  hasMinPlan: (minPlan: PlanTier) => boolean;
}

export const usePlan = (): UsePlanReturn => {
  const user = useAuthStore((state) => state.user);
  const plan = user?.plan;

  return {
    plan,
    isEnterprise: isEnterprise(plan),
    isPaid: hasMinPlan(plan ?? '', 'starter'),
    limits: getPlanLimits(plan),
    hasFeature: (feature: FeatureKey) => hasFeature(plan, feature),
    hasMinPlan: (minPlan: PlanTier) => hasMinPlan(plan ?? '', minPlan),
  };
};
```

---

## 2. Enterprise Badge Component

### EnterpriseBadge (`web/dashboard/src/components/enterprise/EnterpriseBadge.tsx`)

```typescript
import { Crown } from 'lucide-react';
import { motion } from 'framer-motion';
import { usePlan } from '@/hooks/usePlan';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

export function EnterpriseBadge() {
  const { isEnterprise } = usePlan();

  if (!isEnterprise) return null;

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <motion.div
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            className="flex items-center gap-1.5 px-2.5 py-1 rounded-full 
                       bg-gradient-to-r from-amber-500/20 to-yellow-500/20 
                       border border-amber-500/30 cursor-pointer
                       hover:border-amber-500/50 transition-colors"
          >
            <Crown className="w-3.5 h-3.5 text-amber-400" />
            <span className="text-xs font-medium text-amber-400">
              Enterprise
            </span>
          </motion.div>
        </TooltipTrigger>
        <TooltipContent 
          side="bottom" 
          className="bg-bg-tertiary border-white/10"
        >
          <div className="space-y-1">
            <p className="font-medium text-amber-400">Enterprise Plan</p>
            <p className="text-xs text-text-secondary">
              99.99% SLA • Dedicated Support • Unlimited Everything
            </p>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
```

---

## 3. Feature Gate Component

### EnterpriseFeature (`web/dashboard/src/components/enterprise/EnterpriseFeature.tsx`)

```typescript
import { ReactNode } from 'react';
import { Lock } from 'lucide-react';
import { usePlan, type FeatureKey } from '@/hooks/usePlan';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';

interface EnterpriseFeatureProps {
  children: ReactNode;
  feature: FeatureKey;
  fallback?: 'hide' | 'upgrade' | 'blur';
  upgradeMessage?: string;
}

export function EnterpriseFeature({
  children,
  feature,
  fallback = 'upgrade',
  upgradeMessage = 'This feature is available on higher plans',
}: EnterpriseFeatureProps) {
  const { hasFeature } = usePlan();
  const navigate = useNavigate();
  const hasAccess = hasFeature(feature);

  if (hasAccess) return <>{children}</>;

  if (fallback === 'hide') return null;

  if (fallback === 'blur') {
    return (
      <div className="relative">
        <div className="blur-sm pointer-events-none select-none">
          {children}
        </div>
        <div className="absolute inset-0 flex items-center justify-center">
          <UpgradePrompt message={upgradeMessage} />
        </div>
      </div>
    );
  }

  return <UpgradePrompt message={upgradeMessage} />;
}

function UpgradePrompt({ message }: { message: string }) {
  const navigate = useNavigate();

  return (
    <Card className="border-dashed border-white/20 bg-bg-secondary/50">
      <CardContent className="flex flex-col items-center justify-center py-8 text-center">
        <div className="w-12 h-12 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
          <Lock className="w-6 h-6 text-amber-400" />
        </div>
        <h3 className="font-medium text-white mb-1">Enterprise Feature</h3>
        <p className="text-sm text-text-secondary mb-4 max-w-xs">
          {message}
        </p>
        <Button 
          variant="outline" 
          onClick={() => navigate('/pricing')}
          className="border-amber-500/30 hover:border-amber-500/50"
        >
          View Plans
        </Button>
      </CardContent>
    </Card>
  );
}
```

---

## 4. Enhanced Settings Page

### EnterpriseSettingsSection (`web/dashboard/src/components/enterprise/EnterpriseSettingsSection.tsx`)

```typescript
import { Crown, Shield, Headphones, FileText, TrendingUp } from 'lucide-react';
import { usePlan } from '@/hooks/usePlan';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useNavigate } from 'react-router-dom';

export function EnterpriseSettingsSection() {
  const { isEnterprise, plan } = usePlan();
  const navigate = useNavigate();

  if (!isEnterprise) return null;

  return (
    <Card className="border-amber-500/20 bg-gradient-to-br from-amber-500/5 to-transparent">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-amber-500 to-yellow-500 
                            flex items-center justify-center">
              <Crown className="w-5 h-5 text-white" />
            </div>
            <div>
              <CardTitle className="text-white">Enterprise Plan</CardTitle>
              <p className="text-sm text-text-secondary">
                Active since March 2024
              </p>
            </div>
          </div>
          <Badge className="bg-amber-500/20 text-amber-400 border-amber-500/30">
            Active
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Feature Grid */}
        <div className="grid grid-cols-3 gap-4">
          <FeatureCard
            icon={TrendingUp}
            value="99.99%"
            label="SLA Uptime"
          />
          <FeatureCard
            icon={Shield}
            value="∞"
            label="Unlimited"
          />
          <FeatureCard
            icon={Headphones}
            value="24/7"
            label="Support"
          />
        </div>

        {/* Quick Actions */}
        <div className="flex flex-wrap gap-3">
          <Button 
            variant="outline" 
            onClick={() => navigate('/enterprise/sla')}
            className="border-amber-500/30 hover:bg-amber-500/10"
          >
            <TrendingUp className="w-4 h-4 mr-2" />
            SLA Dashboard
          </Button>
          <Button 
            variant="outline"
            onClick={() => navigate('/enterprise/audit')}
            className="border-amber-500/30 hover:bg-amber-500/10"
          >
            <FileText className="w-4 h-4 mr-2" />
            Audit Logs
          </Button>
          <Button 
            variant="outline"
            onClick={() => navigate('/enterprise/support')}
            className="border-amber-500/30 hover:bg-amber-500/10"
          >
            <Headphones className="w-4 h-4 mr-2" />
            Contact Support
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function FeatureCard({ 
  icon: Icon, 
  value, 
  label 
}: { 
  icon: typeof Crown; 
  value: string; 
  label: string;
}) {
  return (
    <div className="p-4 rounded-lg bg-bg-secondary border border-white/8 text-center">
      <Icon className="w-5 h-5 text-amber-400 mx-auto mb-2" />
      <p className="text-lg font-semibold text-white">{value}</p>
      <p className="text-xs text-text-secondary">{label}</p>
    </div>
  );
}
```

---

## 5. Dashboard Enterprise Card

### EnterpriseStatusCard (`web/dashboard/src/components/dashboard/EnterpriseStatusCard.tsx`)

```typescript
import { Crown, TrendingUp, Shield, Clock } from 'lucide-react';
import { motion } from 'framer-motion';
import { usePlan } from '@/hooks/usePlan';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';

export function EnterpriseStatusCard() {
  const { isEnterprise } = usePlan();
  const navigate = useNavigate();

  if (!isEnterprise) return null;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <Card className="border-amber-500/20 overflow-hidden relative">
        {/* Background gradient */}
        <div className="absolute inset-0 bg-gradient-to-br from-amber-500/5 via-transparent to-yellow-500/5" />
        
        <CardHeader className="relative">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-amber-500 to-yellow-500 
                              flex items-center justify-center">
                <Crown className="w-4 h-4 text-white" />
              </div>
              <CardTitle className="text-white text-base">Enterprise Status</CardTitle>
            </div>
            <div className="flex items-center gap-2">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full 
                                 rounded-full bg-green-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500" />
              </span>
              <span className="text-xs text-green-400 font-medium">Active</span>
            </div>
          </div>
        </CardHeader>
        
        <CardContent className="relative space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <StatusItem
              icon={TrendingUp}
              label="SLA Status"
              value="99.99%"
              subtext="Last 30 days"
            />
            <StatusItem
              icon={Shield}
              label="Security"
              value="Compliant"
              subtext="SOC2 Type II"
            />
          </div>
          
          <div className="flex gap-2 pt-2">
            <Button 
              size="sm" 
              variant="outline"
              onClick={() => navigate('/enterprise/sla')}
              className="flex-1 border-amber-500/30 hover:bg-amber-500/10"
            >
              View SLA
            </Button>
            <Button 
              size="sm"
              onClick={() => navigate('/enterprise/support')}
              className="flex-1 bg-gradient-to-r from-amber-500 to-yellow-500 
                         hover:from-amber-600 hover:to-yellow-600"
            >
              Get Support
            </Button>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

function StatusItem({ 
  icon: Icon, 
  label, 
  value, 
  subtext 
}: { 
  icon: typeof Crown; 
  label: string; 
  value: string; 
  subtext: string;
}) {
  return (
    <div className="flex items-start gap-3">
      <div className="w-8 h-8 rounded-lg bg-amber-500/10 flex items-center justify-center flex-shrink-0">
        <Icon className="w-4 h-4 text-amber-400" />
      </div>
      <div>
        <p className="text-sm text-text-secondary">{label}</p>
        <p className="text-white font-medium">{value}</p>
        <p className="text-xs text-text-muted">{subtext}</p>
      </div>
    </div>
  );
}
```

---

## 6. Integration Points

### Update Navbar (`web/dashboard/src/components/layout/Navbar.tsx`)

Add `<EnterpriseBadge />` next to the user avatar:

```typescript
// In the navbar user section:
<div className="flex items-center gap-3">
  <EnterpriseBadge />
  <UserMenu />
</div>
```

### Update Dashboard Page

Add `<EnterpriseStatusCard />` to the top of the dashboard grid:

```typescript
// In DashboardPage:
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
  <EnterpriseStatusCard />
  {/* ... other metric cards */}
</div>
```

### Update Settings Page

Replace the generic billing card with `<EnterpriseSettingsSection />` for enterprise users.

---

## 7. New Routes to Add

Update `ROUTES` in constants:

```typescript
export const ROUTES = {
  // ... existing routes
  
  // Enterprise routes
  ENTERPRISE: '/enterprise',
  ENTERPRISE_SLA: '/enterprise/sla',
  ENTERPRISE_AUDIT: '/enterprise/audit',
  ENTERPRISE_SECURITY: '/enterprise/security',
  ENTERPRISE_SUPPORT: '/enterprise/support',
  ENTERPRISE_COMPLIANCE: '/enterprise/compliance',
} as const;
```

---

## Implementation Checklist

### Phase 1: Core Infrastructure
- [ ] Create `lib/plan-utils.ts` with utility functions
- [ ] Create `hooks/usePlan.ts` hook
- [ ] Add enterprise color tokens to CSS/design system

### Phase 2: Visual Components
- [ ] Create `EnterpriseBadge` component
- [ ] Add badge to Navbar
- [ ] Create `EnterpriseFeature` gate component

### Phase 3: Dashboard Integration
- [ ] Create `EnterpriseStatusCard`
- [ ] Add to Dashboard page
- [ ] Create `EnterpriseSettingsSection`
- [ ] Update Settings page

### Phase 4: New Pages (Optional)
- [ ] Create `EnterpriseSLAPage`
- [ ] Create `EnterpriseAuditPage`
- [ ] Create `EnterpriseSupportPage`

---

## Testing Strategy

1. **Unit Tests**: Test plan utility functions
2. **Component Tests**: Test EnterpriseFeature gate with different plans
3. **Integration Tests**: Test full user flows for enterprise users
4. **Visual Regression**: Screenshots of enterprise UI states
