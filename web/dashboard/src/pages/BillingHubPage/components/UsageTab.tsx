import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import type { CostSummary } from '@/api/usageAnalytics';
import { TrendingUp, Zap, Calendar, DollarSign, BarChart3 } from 'lucide-react';

interface UsageTabProps {
  usageData: { total_events: number; total_cost_usd: number } | null;
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
  costData: CostSummary | null;
  isLoading: boolean;
}

function formatUsd(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
}

export function UsageTab({
  usageData,
  projectedBilling,
  usageMetrics,
  costData,
  isLoading,
}: UsageTabProps) {
  return (
    <div className="space-y-6">
      {/* Usage Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card className="ff-card-velocity">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-muted flex items-center gap-2">
              <Zap className="h-4 w-4" />
              Total Requests
            </CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <p className="text-2xl font-bold font-mono">
                {(usageData?.total_events ?? 0).toLocaleString()}
              </p>
            )}
          </CardContent>
        </Card>

        <Card className="ff-card-velocity">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-muted flex items-center gap-2">
              <DollarSign className="h-4 w-4" />
              Total Cost
            </CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <p className="text-2xl font-bold font-mono text-brand-500">
                {formatUsd(usageData?.total_cost_usd ?? 0)}
              </p>
            )}
          </CardContent>
        </Card>

        <Card className="ff-card-velocity">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-muted flex items-center gap-2">
              <Calendar className="h-4 w-4" />
              Days Remaining
            </CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <p className="text-2xl font-bold font-mono">
                {projectedBilling?.daysRemaining ?? 0}
              </p>
            )}
          </CardContent>
        </Card>

        <Card className="ff-card-velocity">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-muted flex items-center gap-2">
              <TrendingUp className="h-4 w-4" />
              Projected Total
            </CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <p className="text-2xl font-bold font-mono">
                {projectedBilling?.projectedTotal.toLocaleString() ?? 0}
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Usage Details */}
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <BarChart3 className="h-5 w-5 text-brand-500" />
            Usage Details
          </CardTitle>
          <CardDescription>Breakdown of your API usage this billing period</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
            </div>
          ) : usageMetrics.length > 0 ? (
            <div className="space-y-6">
              {usageMetrics.map((metric, idx) => {
                const percent =
                  metric.limit > 0 ? Math.min((metric.current / metric.limit) * 100, 100) : 0;
                const isOver = metric.current > metric.limit && metric.limit > 0;
                return (
                  <div key={idx} className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="font-medium">{metric.label}</span>
                      <span
                        className={`font-mono tabular-nums ${isOver ? 'text-red-400' : 'text-text-secondary'}`}
                      >
                        {metric.current.toLocaleString()} / {metric.limit.toLocaleString()}
                      </span>
                    </div>
                    <div className="relative h-3 rounded-full bg-bg-tertiary overflow-hidden">
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
                    {percent > 80 && (
                      <p
                        className={`text-xs ${isOver ? 'text-red-400' : 'text-amber-400'}`}
                      >
                        {isOver
                          ? `Over limit by ${(metric.current - metric.limit).toLocaleString()}`
                          : `${Math.round(100 - percent)}% remaining`}
                      </p>
                    )}
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="text-center py-8">
              <BarChart3 className="h-12 w-12 text-text-muted mx-auto mb-3" />
              <p className="text-text-muted">No usage data available</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Cost Breakdown */}
      {costData && (
        <Card className="ff-card-velocity">
          <CardHeader>
            <CardTitle className="font-display">Cost Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="p-4 rounded-lg bg-bg-secondary border border-border-default">
                <p className="text-xs text-text-muted uppercase tracking-wide mb-1">Current Period</p>
                <p className="text-xl font-bold font-mono text-brand-500">
                  {formatUsd(costData.total_cost_cents / 100)}
                </p>
              </div>
              {costData.previous_period_cost_usd && (
                <div className="p-4 rounded-lg bg-bg-secondary border border-border-default">
                  <p className="text-xs text-text-muted uppercase tracking-wide mb-1">
                    Previous Period
                  </p>
                  <p className="text-xl font-bold font-mono">
                    {formatUsd(costData.previous_period_cost_usd)}
                  </p>
                </div>
              )}
            </div>

            {costData.previous_period_cost_usd && (
              <div className="mt-4 p-3 rounded-lg bg-brand-500/10 border border-brand-500/20">
                <p className="text-sm">
                  {costData.total_cost_cents / 100 > costData.previous_period_cost_usd ? (
                    <span className="text-amber-400">
                      ↑{' '}
                      {formatUsd(costData.total_cost_cents / 100 - costData.previous_period_cost_usd)}{' '}
                      more than last period
                    </span>
                  ) : (
                    <span className="text-green-400">
                      ↓{' '}
                      {formatUsd(costData.previous_period_cost_usd - costData.total_cost_cents / 100)}{' '}
                      less than last period
                    </span>
                  )}
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Projected Usage */}
      {projectedBilling && projectedBilling.daysRemaining > 0 && (
        <Card className="ff-card-velocity">
          <CardHeader>
            <CardTitle className="font-display">Projected Usage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-text-muted mb-1">By period end</p>
                <p className="text-lg font-semibold font-mono">
                  {projectedBilling.projectedTotal.toLocaleString()}
                  <span className="text-xs text-text-muted ml-1">requests</span>
                </p>
              </div>
              <div>
                <p className="text-xs text-text-muted mb-1">Daily rate</p>
                <p className="text-lg font-semibold font-mono">
                  {projectedBilling.dailyRate > 0
                    ? Math.round(projectedBilling.dailyRate).toLocaleString()
                    : '—'}
                  <span className="text-xs text-text-muted ml-1">/day</span>
                </p>
              </div>
            </div>

            {usageMetrics.length > 0 && projectedBilling.projectedTotal > usageMetrics[0].limit && (
              <div className="mt-4 pt-4 border-t border-border-default">
                <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20">
                  <p className="text-sm text-amber-400">
                    Projected to exceed your plan limit by{' '}
                    {(projectedBilling.projectedTotal - usageMetrics[0].limit).toLocaleString()}{' '}
                    requests. Consider upgrading to avoid overage charges.
                  </p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}