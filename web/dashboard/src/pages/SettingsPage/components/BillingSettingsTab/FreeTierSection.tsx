import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Chamber, CornerBrace, FrameButton, SealedButton, StatusPill } from '@/components/sc';
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
      <div className="flex items-center justify-between p-4 rounded-lg" style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}>
        <div>
          <h3 className="font-semibold font-display capitalize" style={{ color: 'var(--text)' }}>{displayPlan} Plan</h3>
          <div className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
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
          </div>
        </div>
        <StatusPill status="live" label="Current" />
      </div>

      {/* Free tier features list */}
      {isFreeTier && (
        <>
          <div className="p-4 rounded-lg" style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}>
            <p className="font-medium mb-2" style={{ color: 'var(--text)' }}>Your Free Plan includes:</p>
            <ul className="space-y-1 text-sm" style={{ color: 'var(--text-dim)' }}>
              <li className="flex items-center gap-2">
                <span className="w-5 h-5 rounded-full flex items-center justify-center shrink-0" style={{ background: 'rgba(143, 255, 208, 0.2)', border: '1px solid rgba(143, 255, 208, 0.3)' }}>
                  <Check className="w-3.5 h-3.5" style={{ color: 'var(--status-ok)' }} />
                </span>
                Basic function deployment
              </li>
              <li className="flex items-center gap-2">
                <span className="w-5 h-5 rounded-full flex items-center justify-center shrink-0" style={{ background: 'rgba(143, 255, 208, 0.2)', border: '1px solid rgba(143, 255, 208, 0.3)' }}>
                  <Check className="w-3.5 h-3.5" style={{ color: 'var(--status-ok)' }} />
                </span>
                Community support
              </li>
              <li className="flex items-center gap-2">
                <span className="w-5 h-5 rounded-full flex items-center justify-center shrink-0" style={{ background: 'rgba(143, 255, 208, 0.2)', border: '1px solid rgba(143, 255, 208, 0.3)' }}>
                  <Check className="w-3.5 h-3.5" style={{ color: 'var(--status-ok)' }} />
                </span>
                Registry access
              </li>
              <li className="flex items-center gap-2">
                <span className="w-5 h-5 rounded-full flex items-center justify-center shrink-0" style={{ background: 'rgba(143, 255, 208, 0.2)', border: '1px solid rgba(143, 255, 208, 0.3)' }}>
                  <Check className="w-3.5 h-3.5" style={{ color: 'var(--status-ok)' }} />
                </span>
                Up to 5 functions
              </li>
            </ul>
          </div>

          {/* Upgrade prompt for free users */}
          <div className="p-4 rounded-lg" style={{ background: 'rgba(255, 122, 61, 0.05)', border: '1px solid var(--accent-dim)' }}>
            <p className="text-sm font-medium mb-2" style={{ color: 'var(--text)' }}>Ready to unlock more?</p>
            <ul className="space-y-1 text-xs mb-3" style={{ color: 'var(--text-dim)' }}>
              <li>- Unlimited executions</li>
              <li>- Priority support</li>
              <li>- Advanced analytics</li>
            </ul>
            <SealedButton size="sm" onClick={() => (window.location.href = '/pricing')}>
              View Plans & Pricing
            </SealedButton>
          </div>
        </>
      )}

      {/* Free tier plan comparison */}
      {!isFreeTier && (
        <div className="space-y-3">
          <p className="text-sm font-medium" style={{ color: 'var(--text)' }}>Choose Your Plan</p>
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
                    className="p-4 rounded-lg border text-left transition-colors"
                    style={{
                      background: plan.isCurrent ? 'rgba(59, 130, 246, 0.1)' : 'var(--panel-raised)',
                      borderColor: plan.isCurrent ? 'var(--status-ok)' : 'var(--panel-edge)',
                    }}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium" style={{ color: 'var(--text)' }}>{plan.name}</span>
                      {plan.isCurrent && <StatusPill status="live" label="Current" />}
                      {recommended && !plan.isCurrent && (
                        <Badge variant="secondary" className="ff-badge-primary">Recommended</Badge>
                      )}
                    </div>
                    {plan.isUpgrade && <span className="text-xs" style={{ color: 'var(--status-ok)' }}>Upgrade</span>}
                    {plan.isDowngrade && <span className="text-xs" style={{ color: 'var(--status-pending)' }}>Downgrade</span>}
                  </button>
                );
              })}
          </div>
        </div>
      )}
    </div>
  );
}
