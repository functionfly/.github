import type { Subscription } from '@/api/billing';
import type { CostSummary } from '@/api/usageAnalytics';
import { Badge } from '@/components/ui/badge';
import { CollapsibleSection } from '@/components/ui/collapsible-section';
import { PaymentMethodCard } from '@/components/ui/payment-method-card';
import { SpendingSummaryWidget } from '@/components/ui/spending-summary-widget';
import { Zap } from 'lucide-react';

interface CurrentPlanSectionProps {
  subscription: Subscription;
  planOptions: Array<{
    id: string;
    name: string;
    tier: number;
    isCurrent: boolean;
    isUpgrade: boolean;
    isDowngrade: boolean;
  }>;
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
  usageLoading: boolean;
  costData: CostSummary | null;
  costLoading: boolean;
  billingPortalLoading: boolean;
  returnUrl: string;
  openPortal: (urlPath: string) => void;
  formatDate: (date: string) => string;
}

export function CurrentPlanSection({
  subscription,
  planOptions,
  projectedBilling,
  usageMetrics,
  usageLoading,
  costData,
  costLoading,
  billingPortalLoading,
  returnUrl,
  openPortal,
  formatDate,
}: CurrentPlanSectionProps) {
  return (
    <>
      <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-brand-500/10 to-brand-600/10 border border-border-default">
        <div>
          <h3 className="font-semibold text-text-primary capitalize">{subscription.plan} Plan</h3>
          <div className="flex items-center gap-2 mt-1">
            <Badge variant={subscription.status === 'active' ? 'success' : 'secondary'} className={subscription.status === 'active' ? 'ff-badge-success' : ''}>
              {subscription.status}
            </Badge>
            {subscription.cancel_at_period_end && (
              <Badge
                variant="outline"
                className="border-amber-500/50 text-amber-600 dark:text-amber-400"
              >
                Cancels at period end
              </Badge>
            )}
          </div>
        </div>
        <Badge variant="success" className="ff-badge-primary font-semibold px-3 py-1">
          Current
        </Badge>
      </div>

      {/* Trial Period Display */}
      {subscription.is_trialing && (
        <div
          className={`p-4 rounded-lg border ${
            subscription.trial_days_remaining <= 3
              ? 'bg-amber-500/10 border-amber-500/20'
              : 'bg-blue-500/10 border-blue-500/20'
          }`}
        >
          <div className="flex items-center gap-2 mb-2">
            <span className="text-sm font-medium">Trial Period</span>
          </div>
          <p className="text-sm">
            <span
              className={
                subscription.trial_days_remaining <= 3
                  ? 'text-amber-400 font-semibold'
                  : 'text-blue-400'
              }
            >
              {subscription.trial_days_remaining} days remaining
            </span>
            {subscription.trial_end && <> · Ends {formatDate(subscription.trial_end)}</>}
          </p>
          {subscription.trial_days_remaining <= 3 && (
            <p className="text-xs mt-2 text-amber-400">
              Your trial ends soon. Choose a plan to continue using premium features.
            </p>
          )}
        </div>
      )}

      {(subscription.current_period_start || subscription.current_period_end) && (
        <div className="grid grid-cols-2 gap-4 p-4 rounded-lg bg-bg-secondary border border-border-default">
          <div className="flex items-center gap-3">
            <div>
              <p className="text-sm text-text-muted">Current Period Start</p>
              <p className="text-text-primary font-medium">
                {formatDate(subscription.current_period_start)}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div>
              <p className="text-sm text-text-muted">Next Billing Date</p>
              <p className="text-text-primary font-medium">
                {formatDate(subscription.current_period_end)}
              </p>
            </div>
          </div>
        </div>
      )}
      {subscription.payment_method && (
        <PaymentMethodCard
          brand={subscription.payment_method.brand}
          last4={subscription.payment_method.last4}
          expMonth={subscription.payment_method.exp_month}
          expYear={subscription.payment_method.exp_year}
          onUpdate={() => openPortal(returnUrl)}
        />
      )}

      {/* Spending Summary Widget */}
      {costData && (
        <SpendingSummaryWidget
          currentSpend={costData.total_cost_cents}
          previousSpend={
            costData.previous_period_cost_usd ? costData.previous_period_cost_usd * 100 : undefined
          }
          loading={costLoading}
        />
      )}

      {/* Usage Progress - Collapsible */}
      {usageMetrics.length > 0 && (
        <CollapsibleSection
          title="Usage This Period"
          icon={<Zap className="w-4 h-4" />}
          defaultOpen={true}
          headerRight={
            projectedBilling && (
              <span className="text-xs text-text-muted">
                {projectedBilling.daysRemaining}d left
              </span>
            )
          }
          variant="default"
        >
          {usageLoading ? (
            <div className="space-y-2">
              {usageMetrics.map((_, i) => (
                <div key={i} className="h-8 rounded-md bg-bg-hover animate-pulse" />
              ))}
            </div>
          ) : (
            <div className="space-y-3">
              {usageMetrics.map((metric, idx) => {
                const percent =
                  metric.limit > 0 ? Math.min((metric.current / metric.limit) * 100, 100) : 0;
                const isOver = metric.current > metric.limit && metric.limit > 0;
                return (
                  <div key={idx} className="space-y-1.5">
                    <div className="flex items-center justify-between text-xs">
                      <div className="flex items-center gap-2 text-text-muted">
                        {metric.icon}
                        <span>{metric.label}</span>
                      </div>
                      <span
                        className={`font-medium tabular-nums ${isOver ? 'text-red-400' : 'text-text-secondary'}`}
                      >
                        {metric.current.toLocaleString()} / {metric.limit.toLocaleString()}
                      </span>
                    </div>
                    <div className="relative">
                      <div
                        className={`h-2 rounded-full overflow-hidden ${
                          isOver ? 'bg-red-500/20' : 'bg-bg-tertiary'
                        }`}
                      >
                        <div
                          className={`h-full rounded-full transition-all duration-500 ${
                            isOver
                              ? 'bg-gradient-to-r from-red-500 to-red-400'
                              : percent > 80
                                ? 'bg-gradient-to-r from-amber-500 to-orange-400'
                                : 'bg-gradient-to-r from-amber-500 to-orange-500'
                          }`}
                          style={{ width: `${Math.min(percent, 100)}%` }}
                        />
                      </div>
                      {percent > 80 && (
                        <div
                          className={`absolute -right-0.5 -top-0.5 w-2 h-2 rounded-full ${
                            isOver ? 'bg-red-500 animate-pulse' : 'bg-amber-500'
                          }`}
                        />
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CollapsibleSection>
      )}

      {/* Projected Usage & Next Billing */}
      {projectedBilling && projectedBilling.daysRemaining > 0 && usageMetrics.length > 0 && (
        <CollapsibleSection
          title="Projected Usage"
          icon={<span className="text-sm">📈</span>}
          defaultOpen={false}
          variant="highlighted"
        >
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-xs text-text-muted mb-1">By period end</p>
              <p className="text-lg font-semibold text-text-primary tabular-nums">
                {projectedBilling.projectedTotal.toLocaleString()}
                <span className="text-xs text-text-muted ml-1">requests</span>
              </p>
            </div>
            <div>
              <p className="text-xs text-text-muted mb-1">Daily rate</p>
              <p className="text-lg font-semibold text-text-primary tabular-nums">
                {projectedBilling.dailyRate > 0
                  ? Math.round(projectedBilling.dailyRate).toLocaleString()
                  : '—'}
                <span className="text-xs text-text-muted ml-1">/day</span>
              </p>
            </div>
          </div>
          {projectedBilling.projectedTotal > 0 && (
            <div className="mt-3 pt-3 border-t border-border-default">
              <div className="flex items-center justify-between text-xs">
                <span className="text-text-muted">Projection based on current usage</span>
                {projectedBilling.projectedTotal > usageMetrics[0].limit ? (
                  <button
                    onClick={() => (window.location.href = '/pricing')}
                    className="text-amber-400 font-medium"
                  >
                    May exceed limit → Upgrade
                  </button>
                ) : (
                  <span className="font-medium text-green-400">Within limits</span>
                )}
              </div>
            </div>
          )}
        </CollapsibleSection>
      )}

      {/* Plan Comparison */}
      <div className="space-y-3">
        <p className="text-sm font-medium text-text-primary">Change Plan</p>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {planOptions
            .filter((p) => p.tier > 0)
            .map((plan) => {
              const recommended = plan.id === 'professional' && !subscription;
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
                    {plan.isCurrent && <Badge variant="success">Current</Badge>}
                    {recommended && !plan.isCurrent && (
                      <Badge variant="secondary">Recommended</Badge>
                    )}
                  </div>
                  {plan.isUpgrade && <span className="text-xs text-green-400">Upgrade</span>}
                  {plan.isDowngrade && <span className="text-xs text-amber-400">Downgrade</span>}
                </button>
              );
            })}
        </div>
      </div>
    </>
  );
}
