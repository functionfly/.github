import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Check } from 'lucide-react';

interface FreeTierSectionProps {
  displayPlan: string;
  planOptions: Array<{
    id: string;
    name: string;
    tier: number;
    isCurrent: boolean;
    isUpgrade: boolean;
    isDowngrade: boolean;
  }>;
  billingPortalLoading: boolean;
  openPortal: (urlPath: string) => void;
}

export function FreeTierSection({
  displayPlan,
  planOptions,
  billingPortalLoading,
  openPortal,
}: FreeTierSectionProps) {
  const isFreeTier = displayPlan === 'free' || displayPlan.toLowerCase() === 'free';

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-brand-500/10 to-brand-600/10 border border-border-default">
        <div>
          <h3 className="font-semibold text-text-primary capitalize">{displayPlan} Plan</h3>
          <p className="text-sm text-text-secondary mt-1">
            {isFreeTier ? (
              <>
                <Badge variant="default" className="ff-badge-primary mr-2 font-medium">
                  Free Forever
                </Badge>
                Basic features included
              </>
            ) : (
              <>
                <Badge variant="default" className="ff-badge-primary mr-2 font-medium">
                  {displayPlan}
                </Badge>
                Active
              </>
            )}
          </p>
        </div>
        <Badge variant="success" className="ff-badge-success font-semibold px-3 py-1">
          Current
        </Badge>
      </div>

      {/* Free tier features list */}
      {isFreeTier && (
        <>
          <div className="p-4 rounded-lg bg-bg-secondary border border-border-default">
            <p className="font-medium text-text-primary mb-2">Your Free Plan includes:</p>
            <ul className="space-y-1 text-sm text-text-secondary">
              <li className="flex items-center gap-2">
                <span className="w-5 h-5 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center shrink-0">
                  <Check className="w-3.5 h-3.5 text-green-400" />
                </span>
                Basic function deployment
              </li>
              <li className="flex items-center gap-2">
                <span className="w-5 h-5 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center shrink-0">
                  <Check className="w-3.5 h-3.5 text-green-400" />
                </span>
                Community support
              </li>
              <li className="flex items-center gap-2">
                <span className="w-5 h-5 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center shrink-0">
                  <Check className="w-3.5 h-3.5 text-green-400" />
                </span>
                Registry access
              </li>
              <li className="flex items-center gap-2">
                <span className="w-5 h-5 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center shrink-0">
                  <Check className="w-3.5 h-3.5 text-green-400" />
                </span>
                Up to 5 functions
              </li>
            </ul>
          </div>

          {/* Upgrade prompt for free users */}
          <div className="p-4 rounded-lg bg-gradient-to-br from-brand-500/10 to-brand-600/5 border border-brand-500/20">
            <p className="text-sm font-medium text-text-primary mb-2">Ready to unlock more?</p>
            <ul className="space-y-1 text-xs text-text-secondary mb-3">
              <li>- Unlimited executions</li>
              <li>- Priority support</li>
              <li>- Advanced analytics</li>
            </ul>
            <Button
              size="sm"
              onClick={() => (window.location.href = '/pricing')}
              className="ff-btn-velocity"
            >
              View Plans & Pricing
            </Button>
          </div>
        </>
      )}

      {/* Free tier plan comparison */}
      {!isFreeTier && (
        <div className="space-y-3">
          <p className="text-sm font-medium text-text-primary">Choose Your Plan</p>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {planOptions
              .filter((p) => p.tier > 0)
              .map((plan) => {
                const recommended = plan.id === 'starter';
                return (
                  <button
                    key={plan.id}
                    onClick={() => (window.location.href = '/pricing')}
                    disabled={billingPortalLoading}
                    className={`p-4 rounded-lg border text-left transition-colors ${
                      plan.isCurrent
                        ? 'border-brand-500 bg-brand-500/10'
                        : 'border-border-default bg-bg-secondary hover:border-border-strong'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium">{plan.name}</span>
                      {plan.isCurrent && <Badge variant="success" className="ff-badge-success">Current</Badge>}
                      {recommended && !plan.isCurrent && (
                        <Badge variant="secondary" className="ff-badge-primary">Recommended</Badge>
                      )}
                    </div>
                    {plan.isUpgrade && <span className="text-xs text-green-400">Upgrade</span>}
                    {plan.isDowngrade && <span className="text-xs text-amber-400">Downgrade</span>}
                  </button>
                );
              })}
          </div>
        </div>
      )}
    </div>
  );
}
