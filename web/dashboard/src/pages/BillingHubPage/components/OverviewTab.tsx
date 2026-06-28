import React from 'react';
import {
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  TrustSeal,
} from '@/components/containment';
import type { Subscription, PaymentMethod, WalletInfo } from '@/api/billing';
import { formatDate } from '@/pages/SettingsPage/settings-utils';
import { usePlan } from '@/hooks/usePlan';
import {
  CreditCard,
  Zap,
  AlertCircle,
  TrendingUp,
  Wallet,
} from 'lucide-react';

interface OverviewTabProps {
  subscription: Subscription | null;
  walletInfo: WalletInfo | null;
  paymentMethods: PaymentMethod[];
  projectedBilling: {
    periodEnd: Date;
    daysRemaining: number;
    usagePercent: number;
    projectedTotal: number;
    dailyRate: number;
    currentUsage: number;
  } | null;
  usageMetrics: Array<{
    label: string;
    current: number;
    limit: number;
    unit: string;
    icon: React.ReactNode;
  }>;
  planLimits: Record<string, unknown> | null;
  isLoading: {
    subscription: boolean;
    wallet: boolean;
    paymentMethods: boolean;
    usage: boolean;
  };
  errors: {
    subscription: Error | null;
    wallet: Error | null;
    paymentMethods: Error | null;
    usage: Error | null;
  };
  onOpenPortal: () => void;
}

function formatUsd(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
}

export function OverviewTab({
  subscription,
  walletInfo,
  paymentMethods,
  projectedBilling,
  usageMetrics,
  isLoading,
  errors,
  onOpenPortal,
}: OverviewTabProps) {
  const { displayName } = usePlan();
  const defaultPayment = paymentMethods.find((pm) => pm.is_default) ?? paymentMethods[0];

  return (
    <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      {/* Current Plan Card */}
      <Chamber nested>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <div className="sc-billing-card-header" style={{ margin: 'calc(-1 * var(--space-5))', marginBottom: 'var(--space-5)', padding: 'var(--space-4) var(--space-5)' }}>
          <div className="sc-billing-card-title">
            <Zap style={{ width: 14, height: 14 }} />
            Current Plan
          </div>
          <div className="sc-billing-card-description">Your active subscription details</div>
        </div>

        {isLoading.subscription ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            <div style={{ height: 24, width: 128, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
            <div style={{ height: 16, width: 192, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
          </div>
        ) : subscription ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                <StatusPill status={subscription.status === 'active' ? 'live' : subscription.status === 'trialing' ? 'pending' : 'revoked'} label={subscription.status} />
                {subscription.cancel_at_period_end && (
                  <span className="sc-billing-badge sc-billing-badge-warning">Cancels at period end</span>
                )}
                {subscription.is_trialing && (
                  <span className="sc-billing-badge sc-billing-badge-info">Trial — {subscription.trial_days_remaining} days left</span>
                )}
              </div>
              <span style={{ fontFamily: 'var(--font-display)', fontSize: 26, fontWeight: 700, color: 'var(--text)', textTransform: 'capitalize' }}>{displayName}</span>
            </div>

            <div className="sc-billing-grid sc-billing-grid-2" style={{ padding: 'var(--space-4)', background: 'var(--panel)', borderRadius: 'var(--radius)', border: '1px solid var(--panel-edge)' }}>
              <div>
                <p className="sc-billing-stat-label">Current Period</p>
                <p style={{ fontFamily: 'var(--font-body)', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginTop: 'var(--space-1)' }}>
                  {subscription.current_period_start ? formatDate(subscription.current_period_start) : '—'} — {subscription.current_period_end ? formatDate(subscription.current_period_end) : '—'}
                </p>
              </div>
              {subscription.status === 'active' && (
                <div>
                  <p className="sc-billing-stat-label">Next Billing</p>
                  <p style={{ fontFamily: 'var(--font-body)', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginTop: 'var(--space-1)' }}>
                    {subscription.current_period_end ? formatDate(subscription.current_period_end) : '—'}
                  </p>
                </div>
              )}
            </div>

            {defaultPayment && (
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', padding: 'var(--space-3)', borderRadius: 'var(--radius)', background: 'var(--panel)', border: '1px solid var(--panel-edge)' }}>
                <CreditCard style={{ width: 14, height: 14, color: 'var(--text-faint)' }} />
                <span style={{ fontSize: 13, color: 'var(--text)' }}>
                  {defaultPayment.brand} ending in {defaultPayment.last4}
                </span>
                {defaultPayment.is_default && (
                  <span className="sc-billing-badge sc-billing-badge-success" style={{ marginLeft: 'auto' }}>Default</span>
                )}
              </div>
            )}
          </div>
        ) : (
          <div className="sc-billing-info sc-billing-info-warning">
            <AlertCircle style={{ width: 18, height: 18 }} />
            <div className="sc-billing-info-content">
              <div className="sc-billing-info-title">Free Plan</div>
              <div className="sc-billing-info-text">You are not subscribed to any paid plan.</div>
            </div>
          </div>
        )}
      </Chamber>

      {/* Wallet Balance Card */}
      {isLoading.wallet ? (
        <Chamber nested>
          <div className="sc-billing-card-title" style={{ marginBottom: 'var(--space-4)' }}>
            <Wallet style={{ width: 14, height: 14 }} />
            Registry Credits
          </div>
          <div style={{ height: 32, width: 128, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
        </Chamber>
      ) : walletInfo ? (
        <Chamber nested>
          <div className="sc-billing-card-header" style={{ margin: 'calc(-1 * var(--space-5))', marginBottom: 'var(--space-5)', padding: 'var(--space-4) var(--space-5)' }}>
            <div className="sc-billing-card-title">
              <Wallet style={{ width: 14, height: 14, color: 'var(--status-pending)' }} />
              Registry Credits
            </div>
            <div className="sc-billing-card-description">Prepaid balance for registry fees</div>
          </div>
          <GaugeStrip>
            <Gauge data={{ value: formatUsd(walletInfo.balance_usd), label: 'Balance' }} isFirst />
            <Gauge data={{ value: formatUsd(walletInfo.lifetime_earnings_usd), label: 'Lifetime Earned' }} />
            <Gauge data={{ value: formatUsd(walletInfo.lifetime_fees_usd), label: 'Fees Paid' }} />
          </GaugeStrip>
        </Chamber>
      ) : null}

      {/* Usage Summary Card */}
      {usageMetrics.length > 0 && (
        <Chamber nested>
          <div className="sc-billing-card-header" style={{ margin: 'calc(-1 * var(--space-5))', marginBottom: 'var(--space-5)', padding: 'var(--space-4) var(--space-5)' }}>
            <div className="sc-billing-card-title">
              <TrendingUp style={{ width: 14, height: 14 }} />
              Usage This Period
            </div>
            {projectedBilling && (
              <div className="sc-billing-card-description">{projectedBilling.daysRemaining} days remaining</div>
            )}
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            {usageMetrics.map((metric, idx) => {
              const percent = metric.limit > 0 ? Math.min((metric.current / metric.limit) * 100, 100) : 0;
              const isOver = metric.current > metric.limit && metric.limit > 0;
              return (
                <div key={idx} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <span style={{ fontSize: 13, fontWeight: 500, color: 'var(--text)' }}>{metric.label}</span>
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontVariantNumeric: 'tabular-nums', color: isOver ? 'var(--status-revoked)' : 'var(--text-dim)' }}>
                      {metric.current.toLocaleString()} / {metric.limit.toLocaleString()}
                    </span>
                  </div>
                  <div className="sc-billing-progress">
                    <div
                      className={`sc-billing-progress-bar ${isOver ? 'danger' : percent > 80 ? 'warning' : 'success'}`}
                      style={{ width: `${Math.min(percent, 100)}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>

          {projectedBilling && projectedBilling.daysRemaining > 0 && (
            <>
              <div className="sc-billing-divider" />
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontSize: 13 }}>
                <span style={{ color: 'var(--text-dim)' }}>Projected total by period end</span>
                <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 500, color: 'var(--text)' }}>
                  {projectedBilling.projectedTotal.toLocaleString()} requests
                </span>
              </div>
            </>
          )}
        </Chamber>
      )}

      {/* Quick Actions */}
      <Chamber nested>
        <div className="sc-billing-card-title" style={{ marginBottom: 'var(--space-4)' }}>Quick Actions</div>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-3)' }}>
          <FrameButton onClick={onOpenPortal} iconLeft={<CreditCard style={{ width: 14, height: 14 }} />}>
            Manage Billing
          </FrameButton>
          <SealedButton onClick={() => { window.location.href = '/pricing' }} iconLeft={<Zap style={{ width: 14, height: 14 }} />}>
            Upgrade Plan
          </SealedButton>
        </div>
      </Chamber>
    </div>
  );
}
