import React from 'react';
import {
  Chamber,
  GaugeStrip,
  Gauge,
} from '@/components/containment';
import type { CostSummary } from '@/api/usageAnalytics';
import { TrendingUp, Zap, Calendar, DollarSign, BarChart3, AlertCircle } from 'lucide-react';

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
    <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      {/* Usage Summary Cards */}
      <div className="sc-billing-grid sc-billing-grid-4">
        {[
          { icon: Zap, label: 'Total Requests', value: isLoading ? '—' : (usageData?.total_events ?? 0).toLocaleString(), color: 'var(--text)' },
          { icon: DollarSign, label: 'Total Cost', value: isLoading ? '—' : formatUsd(usageData?.total_cost_usd ?? 0), color: 'var(--status-ok)' },
          { icon: Calendar, label: 'Days Remaining', value: isLoading ? '—' : String(projectedBilling?.daysRemaining ?? 0), color: 'var(--text)' },
          { icon: TrendingUp, label: 'Projected Total', value: isLoading ? '—' : (projectedBilling?.projectedTotal.toLocaleString() ?? '0'), color: 'var(--text)' },
        ].map((stat, idx) => {
          const Icon = stat.icon;
          return (
            <Chamber nested key={idx}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-2)' }}>
                <Icon style={{ width: 14, height: 14, color: 'var(--text-faint)' }} />
                <span className="sc-billing-stat-label">{stat.label}</span>
              </div>
              <div className="sc-billing-stat-value" style={{ color: stat.color }}>
                {stat.value}
              </div>
            </Chamber>
          );
        })}
      </div>

      {/* Usage Details */}
      <Chamber nested>
        <div className="sc-billing-card-header" style={{ margin: 'calc(-1 * var(--space-5))', marginBottom: 'var(--space-5)', padding: 'var(--space-4) var(--space-5)' }}>
          <div className="sc-billing-card-title">
            <BarChart3 style={{ width: 14, height: 14 }} />
            Usage Details
          </div>
          <div className="sc-billing-card-description">Breakdown of your API usage this billing period</div>
        </div>

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            <div style={{ height: 96, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
            <div style={{ height: 96, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
          </div>
        ) : usageMetrics.length > 0 ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
            {usageMetrics.map((metric, idx) => {
              const percent = metric.limit > 0 ? Math.min((metric.current / metric.limit) * 100, 100) : 0;
              const isOver = metric.current > metric.limit && metric.limit > 0;
              return (
                <div key={idx} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <span style={{ fontWeight: 500, color: 'var(--text)' }}>{metric.label}</span>
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontVariantNumeric: 'tabular-nums', color: isOver ? 'var(--status-revoked)' : 'var(--text-dim)' }}>
                      {metric.current.toLocaleString()} / {metric.limit.toLocaleString()}
                    </span>
                  </div>
                  <div className="sc-billing-progress" style={{ height: 8 }}>
                    <div
                      className={`sc-billing-progress-bar ${isOver ? 'danger' : percent > 80 ? 'warning' : 'success'}`}
                      style={{ width: `${Math.min(percent, 100)}%` }}
                    />
                  </div>
                  {percent > 80 && (
                    <p style={{ fontSize: 11, color: isOver ? 'var(--status-revoked)' : 'var(--status-pending)' }}>
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
          <div className="empty-state" style={{ minHeight: 120, flexDirection: 'column', gap: 'var(--space-3)' }}>
            <BarChart3 style={{ width: 48, height: 48, color: 'var(--text-faint)' }} />
            <p style={{ color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.06em' }}>No usage data available</p>
          </div>
        )}
      </Chamber>

      {/* Cost Breakdown */}
      {costData && (
        <Chamber nested>
          <div className="sc-billing-card-title" style={{ marginBottom: 'var(--space-4)' }}>Cost Breakdown</div>
          <div className="sc-billing-grid sc-billing-grid-2">
            <div style={{ padding: 'var(--space-4)', borderRadius: 'var(--radius)', background: 'var(--panel)', border: '1px solid var(--panel-edge)' }}>
              <p className="sc-billing-stat-label">Current Period</p>
              <p style={{ fontFamily: 'var(--font-mono)', fontSize: 20, fontWeight: 700, color: 'var(--status-ok)' }}>
                {formatUsd(costData.total_cost_cents / 100)}
              </p>
            </div>
            {costData.previous_period_cost_usd !== undefined && costData.previous_period_cost_usd !== null && (
              <div style={{ padding: 'var(--space-4)', borderRadius: 'var(--radius)', background: 'var(--panel)', border: '1px solid var(--panel-edge)' }}>
                <p className="sc-billing-stat-label">Previous Period</p>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: 20, fontWeight: 700, color: 'var(--text)' }}>
                  {formatUsd(costData.previous_period_cost_usd)}
                </p>
              </div>
            )}
          </div>

          {costData.previous_period_cost_usd !== undefined && costData.previous_period_cost_usd !== null && (
            <div className="sc-billing-info" style={{ marginTop: 'var(--space-4)' }}>
              {costData.total_cost_cents / 100 > costData.previous_period_cost_usd ? (
                <>
                  <TrendingUp style={{ width: 14, height: 14, color: 'var(--status-pending)' }} />
                  <div className="sc-billing-info-content">
                    <span className="sc-billing-info-text" style={{ color: 'var(--status-pending)' }}>
                      ↑ {formatUsd(costData.total_cost_cents / 100 - costData.previous_period_cost_usd)} more than last period
                    </span>
                  </div>
                </>
              ) : (
                <>
                  <TrendingUp style={{ width: 14, height: 14, transform: 'rotate(180deg)' }} />
                  <div className="sc-billing-info-content">
                    <span className="sc-billing-info-text">
                      ↓ {formatUsd(costData.previous_period_cost_usd - costData.total_cost_cents / 100)} less than last period
                    </span>
                  </div>
                </>
              )}
            </div>
          )}
        </Chamber>
      )}

      {/* Projected Usage */}
      {projectedBilling && projectedBilling.daysRemaining > 0 && (
        <Chamber nested>
          <div className="sc-billing-card-title" style={{ marginBottom: 'var(--space-4)' }}>Projected Usage</div>
          <div className="sc-billing-grid sc-billing-grid-2">
            <div>
              <p style={{ fontSize: 11, color: 'var(--text-faint)', marginBottom: 'var(--space-1)' }}>By period end</p>
              <p style={{ fontFamily: 'var(--font-mono)', fontSize: 18, fontWeight: 600, color: 'var(--text)' }}>
                {projectedBilling.projectedTotal.toLocaleString()}
                <span style={{ fontSize: 11, color: 'var(--text-dim)', marginLeft: 'var(--space-1)' }}>requests</span>
              </p>
            </div>
            <div>
              <p style={{ fontSize: 11, color: 'var(--text-faint)', marginBottom: 'var(--space-1)' }}>Daily rate</p>
              <p style={{ fontFamily: 'var(--font-mono)', fontSize: 18, fontWeight: 600, color: 'var(--text)' }}>
                {projectedBilling.dailyRate > 0 ? Math.round(projectedBilling.dailyRate).toLocaleString() : '—'}
                <span style={{ fontSize: 11, color: 'var(--text-dim)', marginLeft: 'var(--space-1)' }}>/day</span>
              </p>
            </div>
          </div>

          {usageMetrics.length > 0 && projectedBilling.projectedTotal > usageMetrics[0].limit && (
            <div className="sc-billing-info sc-billing-info-warning" style={{ marginTop: 'var(--space-4)' }}>
              <AlertCircle style={{ width: 14, height: 14 }} />
              <div className="sc-billing-info-content">
                <span className="sc-billing-info-text">
                  Projected to exceed your plan limit by {(projectedBilling.projectedTotal - usageMetrics[0].limit).toLocaleString()} requests. Consider upgrading to avoid overage charges.
                </span>
              </div>
            </div>
          )}
        </Chamber>
      )}
    </div>
  );
}
