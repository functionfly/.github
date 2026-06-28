import type { Subscription } from '@/api/billing';
import {
  Chamber,
  CornerBrace,
  FrameButton,
  SealedButton,
  StatusPill,
} from '@/components/containment';
import { usePlan } from '@/hooks/usePlan';
import { PLANS } from '@/lib/constants';
import { ArrowDown, ArrowUp, Check, Zap } from 'lucide-react';

interface PlansTabProps {
  subscription: Subscription | null;
  onOpenPortal: () => void;
}

const PLAN_ORDER = ['starter', 'professional', 'enterprise', 'microvm_enterprise'] as const;
type PlanId = (typeof PLAN_ORDER)[number];

export function PlansTab({ subscription, onOpenPortal }: PlansTabProps) {
  const { plan: currentPlan, displayName } = usePlan();
  const userPlanKey = currentPlan?.toUpperCase() as keyof typeof PLANS;

  const currentPlanData = PLANS[userPlanKey];
  const currentTier = PLAN_ORDER.indexOf(currentPlan as PlanId) ?? 0;

  return (
    <div
      className="sc-billing-fade-in"
      style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}
    >
      {/* Current Plan */}
      <Chamber nested>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <div
          className="sc-billing-card-header"
          style={{
            margin: 'calc(-1 * var(--space-5))',
            marginBottom: 'var(--space-5)',
            padding: 'var(--space-4) var(--space-5)',
          }}
        >
          <div className="sc-billing-card-title">
            <Zap style={{ width: 14, height: 14 }} />
            Your Plan
          </div>
          <div className="sc-billing-card-description">Manage your subscription plan</div>
        </div>

        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: 'var(--space-4)',
            borderRadius: 'var(--radius)',
            background: 'linear-gradient(135deg, rgba(143, 255, 208, 0.05), var(--panel))',
            border: '1px solid var(--panel-edge)',
          }}
        >
          <div>
            <h3
              style={{
                fontFamily: 'var(--font-display)',
                fontSize: 18,
                fontWeight: 600,
                color: 'var(--text)',
                textTransform: 'capitalize',
              }}
            >
              {displayName} Plan
            </h3>
            {currentPlanData && (
              <p style={{ fontSize: 13, color: 'var(--text-dim)', marginTop: 'var(--space-1)' }}>
                ${currentPlanData.price}/month
                {currentPlanData.annualDiscount > 0 && (
                  <span style={{ marginLeft: 'var(--space-2)', color: 'var(--status-ok)' }}>
                    {Math.round(currentPlanData.annualDiscount * 100)}% annual discount
                  </span>
                )}
              </p>
            )}
          </div>
          {subscription?.status === 'active' && <StatusPill status="live" label="Active" />}
        </div>

        <div style={{ marginTop: 'var(--space-5)' }}>
          <h4
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: 11,
              fontWeight: 500,
              textTransform: 'uppercase',
              letterSpacing: '0.06em',
              color: 'var(--text-faint)',
              marginBottom: 'var(--space-3)',
            }}
          >
            Plan Limits
          </h4>
          <div className="sc-billing-grid sc-billing-grid-2">
            {currentPlanData?.limits &&
              Object.entries(currentPlanData.limits).map(([key, value]) => {
                if (typeof value !== 'number') return null;
                return (
                  <div
                    key={key}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      padding: 'var(--space-3)',
                      borderRadius: 'var(--radius)',
                      background: 'var(--panel)',
                      border: '1px solid var(--panel-edge)',
                    }}
                  >
                    <span
                      style={{
                        fontFamily: 'var(--font-mono)',
                        fontSize: 11,
                        textTransform: 'uppercase',
                        letterSpacing: '0.04em',
                        color: 'var(--text-faint)',
                      }}
                    >
                      {key.replace(/([A-Z])/g, ' $1').trim()}
                    </span>
                    <span
                      style={{
                        fontFamily: 'var(--font-mono)',
                        fontSize: 13,
                        fontWeight: 500,
                        color: 'var(--text)',
                      }}
                    >
                      {value === Infinity ? 'Unlimited' : value.toLocaleString()}
                    </span>
                  </div>
                );
              })}
          </div>
        </div>
      </Chamber>

      {/* Available Plans */}
      <Chamber nested>
        <div
          className="sc-billing-card-header"
          style={{
            margin: 'calc(-1 * var(--space-5))',
            marginBottom: 'var(--space-5)',
            padding: 'var(--space-4) var(--space-5)',
          }}
        >
          <div className="sc-billing-card-title">Available Plans</div>
          <div className="sc-billing-card-description">Compare plans and upgrade or downgrade</div>
        </div>

        <div className="sc-billing-grid sc-billing-grid-3">
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
                style={{
                  padding: 'var(--space-4)',
                  borderRadius: 'var(--radius)',
                  border: `1px solid ${isCurrent ? 'var(--status-ok)' : 'var(--panel-edge)'}`,
                  background: isCurrent ? 'rgba(143, 255, 208, 0.05)' : 'var(--panel)',
                  transition: 'border-color var(--duration-fast) var(--ease-out)',
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginBottom: 'var(--space-3)',
                  }}
                >
                  <h4
                    style={{
                      fontFamily: 'var(--font-display)',
                      fontSize: 16,
                      fontWeight: 600,
                      color: 'var(--text)',
                      textTransform: 'capitalize',
                    }}
                  >
                    {planData.name}
                  </h4>
                  {isCurrent && <StatusPill status="live" label="Current" />}
                </div>

                <div style={{ marginBottom: 'var(--space-4)' }}>
                  <span
                    style={{
                      fontFamily: 'var(--font-display)',
                      fontSize: 28,
                      fontWeight: 700,
                      color: 'var(--text)',
                    }}
                  >
                    ${planData.price}
                  </span>
                  <span style={{ fontSize: 13, color: 'var(--text-dim)' }}>/month</span>
                </div>

                <div
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 'var(--space-2)',
                    marginBottom: 'var(--space-4)',
                  }}
                >
                  {planData.limits &&
                    Object.entries(planData.limits)
                      .slice(0, 4)
                      .map(([key, value]) => {
                        if (typeof value !== 'number') return null;
                        return (
                          <div
                            key={key}
                            style={{
                              display: 'flex',
                              alignItems: 'center',
                              gap: 'var(--space-2)',
                              fontSize: 12,
                              color: 'var(--text-dim)',
                            }}
                          >
                            <Check
                              style={{
                                width: 12,
                                height: 12,
                                color: 'var(--status-ok)',
                                flexShrink: 0,
                              }}
                            />
                            <span style={{ textTransform: 'capitalize' }}>
                              {key.replace(/([A-Z])/g, ' $1').trim()}:{' '}
                              {value === Infinity ? 'Unlimited' : value.toLocaleString()}
                            </span>
                          </div>
                        );
                      })}
                </div>

                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 'var(--space-2)',
                    fontSize: 12,
                    marginBottom: 'var(--space-4)',
                  }}
                >
                  {isUpgrade && (
                    <span
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 4,
                        color: 'var(--status-ok)',
                      }}
                    >
                      <ArrowUp style={{ width: 12, height: 12 }} />
                      Upgrade available
                    </span>
                  )}
                  {isDowngrade && (
                    <span
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 4,
                        color: 'var(--status-pending)',
                      }}
                    >
                      <ArrowDown style={{ width: 12, height: 12 }} />
                      Downgrade available
                    </span>
                  )}
                </div>

                {isCurrent ? (
                  <FrameButton disabled style={{ width: '100%' }}>
                    Current Plan
                  </FrameButton>
                ) : (
                  <SealedButton
                    style={{ width: '100%' }}
                    onClick={() => {
                      window.location.href = `/pricing?tab=functions&plan=${planId}`;
                    }}
                  >
                    {isUpgrade ? 'Upgrade' : 'Change Plan'}
                  </SealedButton>
                )}
              </div>
            );
          })}
        </div>

        <div className="sc-billing-divider" />

        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: 'var(--space-4)',
            borderRadius: 'var(--radius)',
            background: 'var(--panel)',
            border: '1px solid var(--panel-edge)',
          }}
        >
          <div>
            <p
              style={{
                fontFamily: 'var(--font-body)',
                fontSize: 14,
                fontWeight: 500,
                color: 'var(--text)',
              }}
            >
              Need more?
            </p>
            <p style={{ fontSize: 13, color: 'var(--text-dim)' }}>
              Contact sales for custom pricing and features
            </p>
          </div>
          <FrameButton onClick={onOpenPortal}>Contact Sales</FrameButton>
        </div>
      </Chamber>
    </div>
  );
}
