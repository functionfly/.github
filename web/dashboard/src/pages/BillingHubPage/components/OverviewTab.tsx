import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import type { Subscription, PaymentMethod, WalletInfo } from '@/api/billing';
import type { CostSummary } from '@/api/usageAnalytics';
import { formatCurrency, formatDate } from '@/pages/SettingsPage/settings-utils';
import { usePlan } from '@/hooks/usePlan';
import {
  CreditCard,
  Zap,
  Calendar,
  AlertCircle,
  CheckCircle,
  Clock,
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
  planLimits,
  isLoading,
  errors,
  onOpenPortal,
}: OverviewTabProps) {
  const { plan, displayName } = usePlan();
  const defaultPayment = paymentMethods.find((pm) => pm.is_default) ?? paymentMethods[0];

  return (
    <div className="space-y-6">
      {/* Current Plan Card */}
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <Zap className="h-5 w-5 text-brand-500" />
            Current Plan
          </CardTitle>
          <CardDescription>Your active subscription details</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading.subscription ? (
            <div className="space-y-3">
              <Skeleton className="h-6 w-32" />
              <Skeleton className="h-4 w-48" />
            </div>
          ) : subscription ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Badge
                    variant={subscription.status === 'active' ? 'success' : 'secondary'}
                    className={subscription.status === 'active' ? 'ff-badge-success' : ''}
                  >
                    {subscription.status}
                  </Badge>
                  {subscription.cancel_at_period_end && (
                    <Badge variant="outline" className="border-amber-500/50 text-amber-600">
                      Cancels at period end
                    </Badge>
                  )}
                  {subscription.is_trialing && (
                    <Badge variant="secondary" className="border-blue-500/50 text-blue-600">
                      Trial - {subscription.trial_days_remaining} days left
                    </Badge>
                  )}
                </div>
                <span className="text-2xl font-bold capitalize">{displayName}</span>
              </div>

              <div className="grid grid-cols-2 gap-4 rounded-lg bg-bg-secondary p-4">
                <div>
                  <p className="text-xs text-text-muted uppercase tracking-wide">Current Period</p>
                  <p className="text-sm font-medium mt-1">
                    {subscription.current_period_start
                      ? formatDate(subscription.current_period_start)
                      : '—'}{' '}
                    —{' '}
                    {subscription.current_period_end
                      ? formatDate(subscription.current_period_end)
                      : '—'}
                  </p>
                </div>
                {subscription.status === 'active' && (
                  <div>
                    <p className="text-xs text-text-muted uppercase tracking-wide">Next Billing</p>
                    <p className="text-sm font-medium mt-1">
                      {subscription.current_period_end
                        ? formatDate(subscription.current_period_end)
                        : '—'}
                    </p>
                  </div>
                )}
              </div>

              {defaultPayment && (
                <div className="flex items-center gap-3 p-3 rounded-lg bg-bg-secondary border border-border-default">
                  <CreditCard className="h-4 w-4 text-text-muted" />
                  <span className="text-sm">
                    {defaultPayment.brand} ending in {defaultPayment.last4}
                  </span>
                  {defaultPayment.is_default && (
                    <Badge variant="success" className="ml-auto text-xs">
                      Default
                    </Badge>
                  )}
                </div>
              )}
            </div>
          ) : (
            <div className="flex items-center gap-3 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
              <AlertCircle className="h-5 w-5 text-amber-500" />
              <div>
                <p className="font-medium text-amber-400">Free Plan</p>
                <p className="text-sm text-amber-400/80">
                  You are not subscribed to any paid plan.
                </p>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Wallet Balance Card */}
      {isLoading.wallet ? (
        <Card className="ff-card-velocity">
          <CardHeader>
            <CardTitle className="font-display flex items-center gap-2">
              <Wallet className="h-5 w-5 text-amber-500" />
              Registry Credits
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Skeleton className="h-8 w-32" />
          </CardContent>
        </Card>
      ) : walletInfo ? (
        <Card className="ff-card-velocity">
          <CardHeader>
            <CardTitle className="font-display flex items-center gap-2">
              <Wallet className="h-5 w-5 text-amber-500" />
              Registry Credits
            </CardTitle>
            <CardDescription>Prepaid balance for registry fees</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-4">
              <div>
                <p className="text-xs text-text-muted uppercase tracking-wide">Balance</p>
                <p className="text-2xl font-bold font-mono text-amber-500">
                  {formatUsd(walletInfo.balance_usd)}
                </p>
              </div>
              <div>
                <p className="text-xs text-text-muted uppercase tracking-wide">Lifetime Earned</p>
                <p className="text-lg font-medium font-mono">
                  {formatUsd(walletInfo.lifetime_earnings_usd)}
                </p>
              </div>
              <div>
                <p className="text-xs text-text-muted uppercase tracking-wide">Fees Paid</p>
                <p className="text-lg font-medium font-mono">
                  {formatUsd(walletInfo.lifetime_fees_usd)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      ) : null}

      {/* Usage Summary Card */}
      {usageMetrics.length > 0 && (
        <Card className="ff-card-velocity">
          <CardHeader>
            <CardTitle className="font-display flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-brand-500" />
              Usage This Period
            </CardTitle>
            {projectedBilling && (
              <CardDescription>
                {projectedBilling.daysRemaining} days remaining
              </CardDescription>
            )}
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {usageMetrics.map((metric, idx) => {
                const percent =
                  metric.limit > 0 ? Math.min((metric.current / metric.limit) * 100, 100) : 0;
                const isOver = metric.current > metric.limit && metric.limit > 0;
                return (
                  <div key={idx} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">{metric.label}</span>
                      <span
                        className={`font-mono tabular-nums ${isOver ? 'text-red-400' : 'text-text-secondary'}`}
                      >
                        {metric.current.toLocaleString()} / {metric.limit.toLocaleString()}
                      </span>
                    </div>
                    <div className="relative h-2 rounded-full bg-bg-tertiary overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all ${
                          isOver
                            ? 'bg-gradient-to-r from-red-500 to-red-400'
                            : percent > 80
                              ? 'bg-gradient-to-r from-amber-500 to-orange-400'
                              : 'bg-gradient-to-r from-brand-500 to-indigo-500'
                        }`}
                        style={{ width: `${Math.min(percent, 100)}%` }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>

            {projectedBilling && projectedBilling.daysRemaining > 0 && (
              <div className="mt-4 pt-4 border-t border-border-default">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-text-muted">Projected total by period end</span>
                  <span className="font-mono font-medium">
                    {projectedBilling.projectedTotal.toLocaleString()} requests
                  </span>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Quick Actions */}
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display">Quick Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            <Button
              variant="outline"
              className="border-border-strong"
              onClick={onOpenPortal}
            >
              <CreditCard className="mr-2 h-4 w-4" />
              Manage Billing
            </Button>
            <Button
              variant="outline"
              className="border-border-strong"
              onClick={() => (window.location.href = '/pricing')}
            >
              <Zap className="mr-2 h-4 w-4" />
              Upgrade Plan
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}