import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { Subscription } from '@/api/billing';
import { PLANS } from '@/lib/constants';
import { usePlan } from '@/hooks/usePlan';
import { Check, Zap, ArrowUp, ArrowDown } from 'lucide-react';

interface PlansTabProps {
  subscription: Subscription | null;
  onOpenPortal: () => void;
}

const PLAN_ORDER = ['starter', 'professional', 'enterprise'] as const;
type PlanId = (typeof PLAN_ORDER)[number];

export function PlansTab({ subscription, onOpenPortal }: PlansTabProps) {
  const { plan: currentPlan, displayName } = usePlan();
  const userPlanKey = currentPlan?.toUpperCase() as keyof typeof PLANS;

  const currentPlanData = PLANS[userPlanKey];
  const currentTier = PLAN_ORDER.indexOf(currentPlan as PlanId) ?? 0;

  return (
    <div className="space-y-6">
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <Zap className="h-5 w-5 text-brand-500" />
            Your Plan
          </CardTitle>
          <CardDescription>Manage your subscription plan</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between p-4 rounded-lg bg-gradient-to-r from-brand-500/10 to-brand-600/10 border border-border-default">
            <div>
              <h3 className="font-semibold font-display text-text-primary capitalize">
                {displayName} Plan
              </h3>
              {currentPlanData && (
                <p className="text-sm text-text-secondary mt-1">
                  ${currentPlanData.price}/month
                  {currentPlanData.annualDiscount > 0 && (
                    <span className="ml-2 text-green-400">
                      {Math.round(currentPlanData.annualDiscount * 100)}% annual discount
                    </span>
                  )}
                </p>
              )}
            </div>
            {subscription?.status === 'active' && (
              <Badge variant="success" className="ff-badge-success">
                Active
              </Badge>
            )}
          </div>

          <div className="mt-4">
            <h4 className="text-sm font-medium text-text-muted mb-3">Plan Limits</h4>
            <div className="grid grid-cols-2 gap-3">
              {currentPlanData?.limits && Object.entries(currentPlanData.limits).map(([key, value]) => {
                if (typeof value !== 'number') return null;
                return (
                  <div
                    key={key}
                    className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary border border-border-default"
                  >
                    <span className="text-xs text-text-muted capitalize">
                      {key.replace(/([A-Z])/g, ' $1').trim()}
                    </span>
                    <span className="font-mono text-sm font-medium">
                      {value === Infinity ? 'Unlimited' : value.toLocaleString()}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display">Available Plans</CardTitle>
          <CardDescription>Compare plans and upgrade or downgrade</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            {PLAN_ORDER.map((planId) => {
              const planData = PLANS[planId.toUpperCase() as keyof typeof PLANS];
              if (!planData) return null;

              const planTier = PLAN_ORDER.indexOf(planId);
              const isCurrent = currentPlan?.toLowerCase() === planId;
              const isUpgrade = planTier > currentTier;
              const isDowngrade = planTier < currentTier && planTier > 0;

              return (
                <div
                  key={planId}
                  className={`p-4 rounded-lg border transition-colors ${
                    isCurrent
                      ? 'border-brand-500 bg-brand-500/10'
                      : 'border-border-default bg-bg-secondary hover:border-border-strong'
                  }`}
                >
                  <div className="flex items-center justify-between mb-3">
                    <h4 className="font-semibold capitalize">{planData.name}</h4>
                    {isCurrent && (
                      <Badge variant="success" className="ff-badge-success">
                        Current
                      </Badge>
                    )}
                  </div>

                  <div className="mb-4">
                    <span className="text-2xl font-bold">${planData.price}</span>
                    <span className="text-text-muted">/month</span>
                  </div>

                  <div className="space-y-2 mb-4">
                    {planData.limits && Object.entries(planData.limits).slice(0, 4).map(([key, value]) => {
                      if (typeof value !== 'number') return null;
                      return (
                        <div key={key} className="flex items-center gap-2 text-xs text-text-secondary">
                          <Check className="h-3 w-3 text-green-400" />
                          <span className="capitalize">
                            {key.replace(/([A-Z])/g, ' $1').trim()}:{' '}
                            {value === Infinity ? 'Unlimited' : value.toLocaleString()}
                          </span>
                        </div>
                      );
                    })}
                  </div>

                  <div className="flex items-center gap-2 text-xs mb-4">
                    {isUpgrade && (
                      <span className="flex items-center gap-1 text-green-400">
                        <ArrowUp className="h-3 w-3" />
                        Upgrade available
                      </span>
                    )}
                    {isDowngrade && (
                      <span className="flex items-center gap-1 text-amber-400">
                        <ArrowDown className="h-3 w-3" />
                        Downgrade available
                      </span>
                    )}
                  </div>

                  <Button
                    variant={isCurrent ? 'outline' : 'default'}
                    className="w-full"
                    disabled={isCurrent}
                    onClick={() => (window.location.href = `/pricing?tab=functions&plan=${planId}`)}
                  >
                    {isCurrent ? 'Current Plan' : isUpgrade ? 'Upgrade' : 'Change Plan'}
                  </Button>
                </div>
              );
            })}
          </div>

          <div className="mt-6 flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-default">
            <div>
              <p className="font-medium">Need more?</p>
              <p className="text-sm text-text-muted">Contact sales for custom pricing and features</p>
            </div>
            <Button variant="outline" className="border-border-strong" onClick={onOpenPortal}>
              Contact Sales
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}