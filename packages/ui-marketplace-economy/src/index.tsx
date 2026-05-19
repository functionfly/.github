/**
 * @functionfly/ui-marketplace-economy
 * Marketplace Economy Components - Creator payouts, licensing, royalties, and billing
 */

import React, { useState, useMemo } from 'react';
import { cn } from '@functionfly/ui-core';
import {
  DollarSign,
  TrendingUp,
  TrendingDown,
  Users,
  CreditCard,
  Key,
  BarChart3,
  Crown,
  Percent,
  Tag,
  Target,
  Radar,
  ArrowUpRight,
  ArrowDownRight,
  Minus,
  Check,
  X,
  Settings,
  RefreshCw,
  Download,
  Copy,
  ExternalLink,
  ChevronRight,
  Star,
  Clock,
  AlertCircle,
  Plus,
} from 'lucide-react';

// ============================================================================
// CreatorEconomy
// ============================================================================

export const CreatorEconomy: React.FC<CreatorEconomyProps> = ({
  earnings,
  monthlyTrend,
  activeSubscribers,
  topFunctionId,
  onViewDetails,
  className,
}) => {
  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: earnings.currency || 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  const trendIsUp = monthlyTrend >= 0;

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <DollarSign className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Creator Economy</h3>
          </div>
          <div className="flex items-center gap-2">
            <span className={cn('flex items-center gap-1 text-xs font-medium', trendIsUp ? 'text-green-400' : 'text-red-400')}>
              {trendIsUp ? <ArrowUpRight className="w-3 h-3" /> : <ArrowDownRight className="w-3 h-3" />}
              {Math.abs(monthlyTrend).toFixed(1)}%
            </span>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="p-4 bg-aviation-bg-secondary rounded-lg">
            <div className="text-[10px] text-aviation-text-dim mb-1">Total Revenue</div>
            <div className="text-2xl font-bold text-aviation-text-primary">
              {formatCurrency(earnings.totalRevenue)}
            </div>
          </div>

          <div className="p-4 bg-aviation-bg-secondary rounded-lg">
            <div className="text-[10px] text-aviation-text-dim mb-1">Active Subscribers</div>
            <div className="text-2xl font-bold text-aviation-text-primary">
              {activeSubscribers.toLocaleString()}
            </div>
          </div>

          <div className="p-4 bg-aviation-bg-secondary rounded-lg">
            <div className="text-[10px] text-aviation-text-dim mb-1">Subscription Revenue</div>
            <div className="text-lg font-semibold text-aviation-text-primary">
              {formatCurrency(earnings.subscriptionRevenue)}
            </div>
            <div className="text-[10px] text-aviation-text-muted">recurring monthly</div>
          </div>

          <div className="p-4 bg-aviation-bg-secondary rounded-lg">
            <div className="text-[10px] text-aviation-text-dim mb-1">Pending Payout</div>
            <div className="text-lg font-semibold text-aviation-cyan">
              {formatCurrency(earnings.pendingPayout)}
            </div>
            <div className="text-[10px] text-aviation-text-muted">next payout cycle</div>
          </div>
        </div>

        <div className="mt-4 p-4 bg-aviation-bg-instrument rounded-lg">
          <div className="flex items-center justify-between text-xs">
            <span className="text-aviation-text-dim">Revenue Breakdown</span>
          </div>
          <svg className="w-full h-16 mt-2" viewBox="0 0 100 20" preserveAspectRatio="none">
            <rect x="0" y="0" width={(earnings.subscriptionRevenue / earnings.totalRevenue) * 100} height="10" className="fill-aviation-cyan" rx="2" />
            <rect x={(earnings.subscriptionRevenue / earnings.totalRevenue) * 100 + 1} y="0" width={(earnings.oneTimeRevenue / earnings.totalRevenue) * 100} height="10" className="fill-purple-500" rx="2" />
            <rect x={((earnings.subscriptionRevenue + earnings.oneTimeRevenue) / earnings.totalRevenue) * 100 + 1} y="0" width={(earnings.royaltyRevenue / earnings.totalRevenue) * 100} height="10" className="fill-amber-500" rx="2" />
          </svg>
          <div className="flex items-center justify-center gap-4 mt-2 text-[10px]">
            <span className="flex items-center gap-1">
              <div className="w-2 h-2 rounded bg-aviation-cyan" />
              <span className="text-aviation-text-muted">Subscription</span>
            </span>
            <span className="flex items-center gap-1">
              <div className="w-2 h-2 rounded bg-purple-500" />
              <span className="text-aviation-text-muted">One-time</span>
            </span>
            <span className="flex items-center gap-1">
              <div className="w-2 h-2 rounded bg-amber-500" />
              <span className="text-aviation-text-muted">Royalty</span>
            </span>
          </div>
        </div>
      </div>

      <div className="px-4 py-3 border-t border-aviation-border-panel">
        <button
          onClick={onViewDetails}
          className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-aviation-bg-instrument hover:bg-aviation-bg-secondary text-aviation-text-primary rounded-lg transition-colors text-sm"
        >
          View Full Report
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
};

// ============================================================================
// RevenueAnalytics
// ============================================================================

export const RevenueAnalytics: React.FC<RevenueAnalyticsProps> = ({
  data,
  totalRevenue,
  period,
  onPeriodChange,
  onDataPointSelect,
  className,
}) => {
  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
    }).format(value);
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const maxRevenue = useMemo(() => Math.max(...data.map(d => d.revenue), 1), [data]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <TrendingUp className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Revenue Analytics</h3>
          </div>
          <div className="flex items-center gap-1">
            {(['7d', '30d', '90d', '1y'] as const).map(p => (
              <button
                key={p}
                onClick={() => onPeriodChange?.(p)}
                className={cn(
                  'px-2 py-1 text-xs rounded transition-colors',
                  period === p
                    ? 'bg-aviation-cyan text-aviation-bg-primary'
                    : 'text-aviation-text-muted hover:text-aviation-text-primary'
                )}
              >
                {p}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center justify-between">
          <span className="text-xs text-aviation-text-dim">Total Revenue</span>
          <span className="text-lg font-bold text-aviation-text-primary">{formatCurrency(totalRevenue)}</span>
        </div>
      </div>

      <div className="flex-1 overflow-hidden p-4">
        <svg className="w-full h-full" viewBox="0 0 100 50" preserveAspectRatio="none">
          {[0, 25, 50, 75, 100].map(y => (
            <line key={y} x1="0" y1={y} x2="100" y2={y} className="stroke-aviation-border-panel" strokeWidth="0.5" strokeDasharray="2 2" />
          ))}

          <path
            d={data.map((d, i) => {
              const x = (i / (data.length - 1)) * 100;
              const y = 50 - (d.revenue / maxRevenue) * 50;
              return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
            }).join(' ') + ' L 100 50 L 0 50 Z'}
            className="fill-aviation-cyan/20"
          />

          <path
            d={data.map((d, i) => {
              const x = (i / (data.length - 1)) * 100;
              const y = 50 - (d.revenue / maxRevenue) * 50;
              return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
            }).join(' ')}
            className="fill-none stroke-aviation-cyan"
            strokeWidth="1.5"
          />

          {data.filter((_, i) => i % Math.ceil(data.length / 6) === 0).map((d, i) => {
            const x = (i * Math.ceil(data.length / 6) / (data.length - 1)) * 100;
            const y = 50 - (d.revenue / maxRevenue) * 50;
            return (
              <g key={i}>
                <circle cx={x} cy={y} r="1.5" className="fill-aviation-cyan" />
              </g>
            );
          })}
        </svg>
        <div className="flex justify-between mt-2 text-[10px] text-aviation-text-dim">
          <span>{formatTime(data[0]?.timestamp)}</span>
          <span>{formatTime(data[data.length - 1]?.timestamp)}</span>
        </div>
      </div>

      <div className="px-4 py-3 border-t border-aviation-border-panel">
        <div className="flex items-center justify-center gap-4 text-[10px]">
          <span className="flex items-center gap-1">
            <div className="w-3 h-0.5 bg-aviation-cyan" />
            <span className="text-aviation-text-muted">Total</span>
          </span>
          <span className="flex items-center gap-1">
            <div className="w-3 h-0.5 bg-purple-500" />
            <span className="text-aviation-text-muted">Subscription</span>
          </span>
          <span className="flex items-center gap-1">
            <div className="w-3 h-0.5 bg-amber-500" />
            <span className="text-aviation-text-muted">One-time</span>
          </span>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// SubscriptionManager
// ============================================================================

export const SubscriptionManager: React.FC<SubscriptionManagerProps> = ({
  subscriptions,
  activeCount,
  cancelledCount,
  pastDueCount,
  onSubscriptionSelect,
  className,
}) => {
  const getStatusColor = (status: SubscriptionStatus) => {
    switch (status) {
      case 'active': return { bg: 'bg-green-500/20', text: 'text-green-400', border: 'border-green-500' };
      case 'cancelled': return { bg: 'bg-red-500/20', text: 'text-red-400', border: 'border-red-500' };
      case 'past_due': return { bg: 'bg-amber-500/20', text: 'text-amber-400', border: 'border-amber-500' };
      case 'trialing': return { bg: 'bg-purple-500/20', text: 'text-purple-400', border: 'border-purple-500' };
      default: return { bg: 'bg-aviation-bg-secondary', text: 'text-aviation-text-muted', border: 'border-aviation-border-panel' };
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Subscriptions</h3>
          </div>
        </div>
      </div>

      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="grid grid-cols-3 gap-4 text-center">
          <div>
            <div className="text-lg font-bold text-green-400">{activeCount}</div>
            <div className="text-[10px] text-aviation-text-dim">Active</div>
          </div>
          <div>
            <div className="text-lg font-bold text-amber-400">{pastDueCount}</div>
            <div className="text-[10px] text-aviation-text-dim">Past Due</div>
          </div>
          <div>
            <div className="text-lg font-bold text-red-400">{cancelledCount}</div>
            <div className="text-[10px] text-aviation-text-dim">Cancelled</div>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {subscriptions.map(sub => {
          const colors = getStatusColor(sub.status);
          return (
            <div
              key={sub.id}
              onClick={() => onSubscriptionSelect?.(sub)}
              className="p-4 border-b border-aviation-border-panel cursor-pointer hover:bg-aviation-bg-secondary transition-colors"
            >
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-aviation-text-primary">{sub.customerName}</span>
                <span className={cn('px-2 py-0.5 text-[10px] rounded uppercase', colors.bg, colors.text)}>
                  {sub.status}
                </span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-aviation-text-muted">{sub.plan}</span>
                <span className="text-aviation-text-primary font-medium">
                  ${sub.amount.toFixed(2)}/{sub.billingCycle}
                </span>
              </div>
              <div className="mt-2 flex items-center gap-2 text-[10px] text-aviation-text-dim">
                <Clock className="w-3 h-3" />
                <span>Renews {new Date(sub.currentPeriodEnd).toLocaleDateString()}</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// UsageBillingPanel
// ============================================================================

export const UsageBillingPanel: React.FC<UsageBillingPanelProps> = ({
  metrics,
  totalCost,
  billingCycle,
  currency,
  onUpgradeClick,
  className,
}) => {
  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD',
    }).format(value);
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <CreditCard className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Usage Billing</h3>
          </div>
          <span className="text-xs text-aviation-text-muted capitalize">{billingCycle}</span>
        </div>
      </div>

      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center justify-between">
          <span className="text-xs text-aviation-text-dim">Total Cost</span>
          <span className="text-lg font-bold text-aviation-text-primary">{formatCurrency(totalCost)}</span>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {metrics.map(metric => {
          const usagePercent = (metric.used / metric.limit) * 100;
          const isHigh = usagePercent > 80;

          return (
            <div key={metric.name} className="mb-4">
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm text-aviation-text-primary">{metric.name}</span>
                <span className="text-sm text-aviation-text-primary">{formatCurrency(metric.cost)}</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="flex-1 h-2 bg-aviation-bg-instrument rounded-full overflow-hidden">
                  <div
                    className={cn('h-full transition-all', isHigh ? 'bg-amber-500' : 'bg-aviation-cyan')}
                    style={{ width: `${usagePercent}%` }}
                  />
                </div>
                <span className="text-[10px] text-aviation-text-dim w-20 text-right">
                  {metric.used.toLocaleString()} / {metric.limit.toLocaleString()} {metric.unit}
                </span>
              </div>
            </div>
          );
        })}
      </div>

      <div className="px-4 py-3 border-t border-aviation-border-panel">
        <button
          onClick={onUpgradeClick}
          className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-aviation-cyan text-aviation-bg-primary rounded-lg hover:bg-aviation-cyan/90 transition-colors text-sm"
        >
          Upgrade Plan
        </button>
      </div>
    </div>
  );
};

// ============================================================================
// LicenseManager
// ============================================================================

export const LicenseManager: React.FC<LicenseManagerProps> = ({
  licenses,
  totalActive,
  totalRevoked,
  onLicenseSelect,
  onLicenseRevoke,
  onLicenseGenerate,
  className,
}) => {
  const getLicenseTypeColor = (type: LicenseType) => {
    switch (type) {
      case 'open': return 'text-green-400';
      case 'restricted': return 'text-amber-400';
      case 'commercial': return 'text-purple-400';
      default: return 'text-aviation-text-muted';
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Key className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Licenses</h3>
          </div>
          <button
            onClick={onLicenseGenerate}
            className="flex items-center gap-1 px-2 py-1 bg-aviation-bg-instrument hover:bg-aviation-bg-secondary rounded text-xs text-aviation-text-primary transition-colors"
          >
            <Plus className="w-3 h-3" />
            Generate
          </button>
        </div>
      </div>

      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="grid grid-cols-2 gap-4 text-center">
          <div>
            <div className="text-lg font-bold text-green-400">{totalActive}</div>
            <div className="text-[10px] text-aviation-text-dim">Active</div>
          </div>
          <div>
            <div className="text-lg font-bold text-red-400">{totalRevoked}</div>
            <div className="text-[10px] text-aviation-text-dim">Revoked</div>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {licenses.map(license => (
          <div
            key={license.id}
            className="p-4 border-b border-aviation-border-panel cursor-pointer hover:bg-aviation-bg-secondary transition-colors"
          >
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-aviation-text-primary">{license.functionName}</span>
                <span className={cn('text-[10px] uppercase', getLicenseTypeColor(license.type))}>
                  {license.type}
                </span>
              </div>
              {!license.revoked && (
                <button
                  onClick={(e) => { e.stopPropagation(); onLicenseRevoke?.(license.id); }}
                  className="p-1 hover:bg-aviation-bg-panel rounded"
                >
                  <X className="w-3 h-3 text-red-400" />
                </button>
              )}
            </div>
            <div className="flex items-center gap-2 mb-2">
              <code className="text-[10px] bg-aviation-bg-instrument px-2 py-1 rounded text-aviation-text-muted">
                {license.key.slice(0, 8)}...{license.key.slice(-4)}
              </code>
              {license.maxActivations && (
                <span className="text-[10px] text-aviation-text-dim">
                  {license.activationCount}/{license.maxActivations} used
                </span>
              )}
            </div>
            <div className="text-xs text-aviation-text-dim">
              to {license.purchaserName} • {new Date(license.issuedAt).toLocaleDateString()}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// CreatorProfile
// ============================================================================

export const CreatorProfile: React.FC<CreatorProfileProps> = ({
  creatorId,
  creatorName,
  creatorAvatar,
  stats,
  payoutInfo,
  className,
}) => {
  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: payoutInfo.currency || 'USD',
    }).format(value);
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <Crown className="w-5 h-5 text-amber-400" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Creator Profile</h3>
        </div>
      </div>

      <div className="p-4 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center gap-4">
          <div className="w-16 h-16 rounded-full bg-aviation-bg-instrument flex items-center justify-center text-2xl font-bold text-aviation-cyan">
            {creatorAvatar ? '👤' : creatorName[0].toUpperCase()}
          </div>
          <div>
            <h4 className="text-lg font-medium text-aviation-text-primary">{creatorName}</h4>
            <p className="text-xs text-aviation-text-muted">ID: {creatorId.slice(0, 8)}</p>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-2 gap-3">
          <div className="p-3 bg-aviation-bg-secondary rounded-lg text-center">
            <div className="text-lg font-bold text-aviation-text-primary">{stats.totalSales.toLocaleString()}</div>
            <div className="text-[10px] text-aviation-text-dim">Total Sales</div>
          </div>
          <div className="p-3 bg-aviation-bg-secondary rounded-lg text-center">
            <div className="text-lg font-bold text-aviation-text-primary">{formatCurrency(stats.totalRevenue)}</div>
            <div className="text-[10px] text-aviation-text-dim">Total Revenue</div>
          </div>
          <div className="p-3 bg-aviation-bg-secondary rounded-lg text-center">
            <div className="flex items-center justify-center gap-1">
              <Star className="w-4 h-4 text-amber-400" />
              <span className="text-lg font-bold text-aviation-text-primary">{stats.averageRating.toFixed(1)}</span>
            </div>
            <div className="text-[10px] text-aviation-text-dim">{stats.totalReviews} reviews</div>
          </div>
          <div className="p-3 bg-aviation-bg-secondary rounded-lg text-center">
            <div className="text-lg font-bold text-aviation-text-primary">{stats.functionsPublished}</div>
            <div className="text-[10px] text-aviation-text-dim">Functions</div>
          </div>
        </div>

        <div className="mt-4 p-4 bg-aviation-bg-instrument rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <DollarSign className="w-4 h-4 text-aviation-cyan" />
            <span className="text-sm font-medium text-aviation-text-primary">Next Payout</span>
          </div>
          <div className="flex items-center justify-between">
            <div>
              <div className="text-xl font-bold text-aviation-text-primary">{formatCurrency(payoutInfo.nextPayoutAmount)}</div>
              <div className="text-[10px] text-aviation-text-dim">
                {new Date(payoutInfo.nextPayoutDate).toLocaleDateString()}
              </div>
            </div>
            <div className="text-right">
              <div className="text-xs text-aviation-text-muted">{payoutInfo.payoutMethod}</div>
              <div className="text-[10px] text-aviation-text-dim">Min: {formatCurrency(payoutInfo.minimumPayout)}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// MarketplaceLeaderboard
// ============================================================================

export const MarketplaceLeaderboard: React.FC<MarketplaceLeaderboardProps> = ({
  entries,
  category,
  timeRange,
  onEntrySelect,
  className,
}) => {
  const getTrendIcon = (direction: TrendDirection) => {
    switch (direction) {
      case 'up': return <ArrowUpRight className="w-3 h-3 text-green-400" />;
      case 'down': return <ArrowDownRight className="w-3 h-3 text-red-400" />;
      default: return <Minus className="w-3 h-3 text-aviation-text-muted" />;
    }
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
    }).format(value);
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Crown className="w-5 h-5 text-amber-400" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Leaderboard</h3>
          </div>
          <div className="flex items-center gap-2">
            <select
              value={category}
              className="text-xs bg-aviation-bg-instrument border border-aviation-border-panel rounded px-2 py-1 text-aviation-text-primary"
            >
              <option value="creators">Creators</option>
              <option value="functions">Functions</option>
              <option value="revenue">Revenue</option>
            </select>
            <select
              value={timeRange}
              className="text-xs bg-aviation-bg-instrument border border-aviation-border-panel rounded px-2 py-1 text-aviation-text-primary"
            >
              <option value="7d">7 days</option>
              <option value="30d">30 days</option>
              <option value="90d">90 days</option>
              <option value="all">All time</option>
            </select>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {entries.map(entry => (
          <div
            key={entry.rank}
            onClick={() => onEntrySelect?.(entry)}
            className="p-4 border-b border-aviation-border-panel cursor-pointer hover:bg-aviation-bg-secondary transition-colors"
          >
            <div className="flex items-center gap-3">
              <div className={cn(
                'w-8 h-8 flex items-center justify-center rounded-lg font-bold text-sm',
                entry.rank === 1 ? 'bg-amber-500/20 text-amber-400' :
                entry.rank === 2 ? 'bg-gray-400/20 text-gray-300' :
                entry.rank === 3 ? 'bg-orange-600/20 text-orange-400' :
                'bg-aviation-bg-instrument text-aviation-text-muted'
              )}>
                #{entry.rank}
              </div>
              <div className="flex-1">
                <div className="text-sm font-medium text-aviation-text-primary">
                  {category === 'creators' ? entry.creatorName : entry.functionName}
                </div>
                <div className="text-xs text-aviation-text-muted">
                  {category === 'creators' ? `${entry.functionName} • ${entry.sales.toLocaleString()} sales` : `${entry.creatorName} • ${entry.sales.toLocaleString()} sales`}
                </div>
              </div>
              <div className="text-right">
                <div className="text-sm font-bold text-aviation-text-primary">{formatCurrency(entry.revenue)}</div>
                <div className="flex items-center justify-end gap-1">
                  <Star className="w-3 h-3 text-amber-400" />
                  <span className="text-[10px] text-aviation-text-dim">{entry.rating.toFixed(1)}</span>
                  {getTrendIcon(entry.trend)}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// FunctionRoyaltiesPanel
// ============================================================================

export const FunctionRoyaltiesPanel: React.FC<FunctionRoyaltiesPanelProps> = ({
  royalties,
  totalEarned,
  totalPending,
  currency,
  onRoyaltySelect,
  onClaimPending,
  className,
}) => {
  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD',
    }).format(value);
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <Percent className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Function Royalties</h3>
        </div>
      </div>

      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="grid grid-cols-2 gap-4 text-center">
          <div>
            <div className="text-lg font-bold text-green-400">{formatCurrency(totalEarned)}</div>
            <div className="text-[10px] text-aviation-text-dim">Total Earned</div>
          </div>
          <div>
            <div className="text-lg font-bold text-amber-400">{formatCurrency(totalPending)}</div>
            <div className="text-[10px] text-aviation-text-dim">Pending</div>
          </div>
        </div>
      </div>

      {totalPending > 0 && (
        <div className="px-4 py-2 border-b border-aviation-border-panel">
          <button
            onClick={onClaimPending}
            className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-amber-500/20 text-amber-400 rounded-lg hover:bg-amber-500/30 transition-colors text-sm"
          >
            <DollarSign className="w-4 h-4" />
            Claim Pending Royalties
          </button>
        </div>
      )}

      <div className="flex-1 overflow-y-auto">
        {royalties.map(royalty => (
          <div
            key={royalty.id}
            onClick={() => onRoyaltySelect?.(royalty)}
            className="p-4 border-b border-aviation-border-panel cursor-pointer hover:bg-aviation-bg-secondary transition-colors"
          >
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-aviation-text-primary">{royalty.functionName}</span>
              <span className="text-sm font-bold text-green-400">{formatCurrency(royalty.royaltyAmount)}</span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-aviation-text-muted">{royalty.licensee} • {royalty.licenseType}</span>
              <span className="text-aviation-text-dim">{royalty.royaltyPercentage}% royalty</span>
            </div>
            <div className="mt-2 flex items-center gap-2">
              <span className="text-[10px] text-aviation-text-dim">
                Sale: {formatCurrency(royalty.saleAmount)} on {new Date(royalty.saleDate).toLocaleDateString()}
              </span>
              {royalty.paidOut ? (
                <span className="flex items-center gap-1 text-[10px] text-green-400">
                  <Check className="w-3 h-3" /> Paid
                </span>
              ) : (
                <span className="flex items-center gap-1 text-[10px] text-amber-400">
                  <Clock className="w-3 h-3" /> Pending
                </span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// AssetPricingEditor
// ============================================================================

export const AssetPricingEditor: React.FC<AssetPricingEditorProps> = ({
  functionId,
  functionName,
  currentPricing,
  onPriceUpdate,
  onSave,
  className,
}) => {
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Tag className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Pricing Editor</h3>
          </div>
        </div>
      </div>

      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="text-sm text-aviation-text-primary font-medium">{functionName}</div>
        <div className="text-[10px] text-aviation-text-muted">ID: {functionId.slice(0, 8)}</div>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {currentPricing.map((tier, index) => (
          <div key={tier.id} className="mb-4 p-4 bg-aviation-bg-secondary rounded-lg">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-aviation-text-primary">{tier.name}</span>
                {tier.popular && (
                  <span className="px-2 py-0.5 bg-amber-500/20 text-amber-400 text-[10px] rounded uppercase">Popular</span>
                )}
              </div>
              <div className="flex items-center gap-2">
                <span className="text-lg font-bold text-aviation-text-primary">
                  ${tier.price.toFixed(2)}
                </span>
                <span className="text-[10px] text-aviation-text-muted">USD</span>
              </div>
            </div>
            <div className="space-y-1">
              {tier.features.map((feature, i) => (
                <div key={i} className="flex items-center gap-2 text-xs text-aviation-text-muted">
                  <Check className="w-3 h-3 text-green-400" />
                  <span>{feature}</span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="px-4 py-3 border-t border-aviation-border-panel flex gap-2">
        <button
          onClick={() => onPriceUpdate?.(currentPricing)}
          className="flex-1 px-4 py-2 bg-aviation-bg-instrument hover:bg-aviation-bg-secondary text-aviation-text-primary rounded-lg transition-colors text-sm"
        >
          Reset
        </button>
        <button
          onClick={onSave}
          className="flex-1 px-4 py-2 bg-aviation-cyan text-aviation-bg-primary rounded-lg hover:bg-aviation-cyan/90 transition-colors text-sm"
        >
          Save Changes
        </button>
      </div>
    </div>
  );
};

// ============================================================================
// SalesConversionAnalytics
// ============================================================================

export const SalesConversionAnalytics: React.FC<SalesConversionAnalyticsProps> = ({
  funnelSteps,
  totalVisitors,
  totalPurchases,
  overallConversionRate,
  averageOrderValue,
  className,
}) => {
  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(value);
  };

  const maxCount = Math.max(...funnelSteps.map(s => s.count));

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <Target className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Conversion Analytics</h3>
        </div>
      </div>

      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="grid grid-cols-4 gap-3 text-center">
          <div>
            <div className="text-lg font-bold text-aviation-text-primary">{totalVisitors.toLocaleString()}</div>
            <div className="text-[10px] text-aviation-text-dim">Visitors</div>
          </div>
          <div>
            <div className="text-lg font-bold text-aviation-text-primary">{totalPurchases.toLocaleString()}</div>
            <div className="text-[10px] text-aviation-text-dim">Purchases</div>
          </div>
          <div>
            <div className="text-lg font-bold text-aviation-cyan">{(overallConversionRate * 100).toFixed(2)}%</div>
            <div className="text-[10px] text-aviation-text-dim">Conv. Rate</div>
          </div>
          <div>
            <div className="text-lg font-bold text-aviation-text-primary">{formatCurrency(averageOrderValue)}</div>
            <div className="text-[10px] text-aviation-text-dim">Avg Order</div>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        <div className="flex items-end justify-between gap-2 h-48">
          {funnelSteps.map((step, index) => {
            const height = (step.count / maxCount) * 100;
            return (
              <div key={step.stage} className="flex-1 flex flex-col items-center">
                <div className="w-full flex-1 flex items-end justify-center">
                  <div
                    className="w-12 bg-aviation-cyan/60 rounded-t-lg transition-all"
                    style={{ height: `${height}%` }}
                  >
                    <div className="h-full flex items-center justify-center">
                      <span className="text-xs font-bold text-aviation-text-primary">
                        {step.count >= 1000 ? `${(step.count / 1000).toFixed(1)}k` : step.count}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="mt-2 text-center">
                  <div className="text-[10px] text-aviation-text-primary font-medium">{step.stage}</div>
                  <div className="text-[10px] text-aviation-text-dim">{(step.conversionRate * 100).toFixed(1)}%</div>
                </div>
              </div>
            );
          })}
        </div>

        <div className="mt-4 p-3 bg-aviation-bg-instrument rounded-lg">
          <div className="text-xs text-aviation-text-muted mb-2">Drop-off Analysis</div>
          {funnelSteps.slice(0, -1).map((step, index) => (
            <div key={step.stage} className="flex items-center justify-between text-xs py-1">
              <span className="text-aviation-text-primary">{step.stage}</span>
              <span className="text-red-400">-{(step.dropOffRate * 100).toFixed(1)}%</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// MonetizationOptimizer
// ============================================================================

export const MonetizationOptimizer: React.FC<MonetizationOptimizerProps> = ({
  suggestions,
  currentRevenue,
  projectedRevenue,
  onSuggestionApply,
  onSuggestionDismiss,
  className,
}) => {
  const getPriorityColor = (priority: 'high' | 'medium' | 'low') => {
    switch (priority) {
      case 'high': return { bg: 'bg-red-500/20', text: 'text-red-400', border: 'border-red-500' };
      case 'medium': return { bg: 'bg-amber-500/20', text: 'text-amber-400', border: 'border-amber-500' };
      default: return { bg: 'bg-green-500/20', text: 'text-green-400', border: 'border-green-500' };
    }
  };

  const getEffortColor = (effort: 'low' | 'medium' | 'high') => {
    switch (effort) {
      case 'low': return 'text-green-400';
      case 'medium': return 'text-amber-400';
      default: return 'text-red-400';
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <TrendingUp className="w-5 h-5 text-aviation-cyan" />
          <h3 className="text-sm font-medium text-aviation-text-primary">Monetization Optimizer</h3>
        </div>
      </div>

      {projectedRevenue && (
        <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between">
            <span className="text-xs text-aviation-text-dim">Projected Revenue</span>
            <div className="flex items-center gap-2">
              <span className="text-lg font-bold text-aviation-cyan">{projectedRevenue.toLocaleString()}</span>
              {currentRevenue > 0 && (
                <span className="text-xs text-green-400">
                  +{(((projectedRevenue - currentRevenue) / currentRevenue) * 100).toFixed(1)}%
                </span>
              )}
            </div>
          </div>
        </div>
      )}

      <div className="flex-1 overflow-y-auto">
        {suggestions.map(suggestion => {
          const colors = getPriorityColor(suggestion.priority);
          return (
            <div
              key={suggestion.id}
              className="p-4 border-b border-aviation-border-panel"
            >
              <div className="flex items-start justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className={cn('px-2 py-0.5 text-[10px] rounded uppercase', colors.bg, colors.text)}>
                    {suggestion.priority}
                  </span>
                  <span className="text-xs text-aviation-text-muted uppercase">{suggestion.type}</span>
                </div>
                <button
                  onClick={() => onSuggestionDismiss?.(suggestion.id)}
                  className="p-1 hover:bg-aviation-bg-panel rounded"
                >
                  <X className="w-3 h-3 text-aviation-text-muted" />
                </button>
              </div>
              <h4 className="text-sm font-medium text-aviation-text-primary mb-1">{suggestion.title}</h4>
              <p className="text-xs text-aviation-text-muted mb-3">{suggestion.description}</p>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3 text-[10px]">
                  <span className="text-aviation-text-dim">
                    Impact: <span className="text-green-400">+{suggestion.potentialImpact}%</span>
                  </span>
                  <span className="text-aviation-text-dim">
                    Effort: <span className={getEffortColor(suggestion.effort)}>{suggestion.effort}</span>
                  </span>
                </div>
                {suggestion.actionable && (
                  <button
                    onClick={() => onSuggestionApply?.(suggestion)}
                    className="px-3 py-1 bg-aviation-cyan/20 text-aviation-cyan text-xs rounded hover:bg-aviation-cyan/30 transition-colors"
                  >
                    Apply
                  </button>
                )}
              </div>
            </div>
          );
        })}

        {suggestions.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <Check className="w-8 h-8 mb-2 text-green-400" />
            <p className="text-sm">No optimization suggestions</p>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// MarketplaceTrendRadar
// ============================================================================

export const MarketplaceTrendRadar: React.FC<MarketplaceTrendRadarProps> = ({
  trends,
  timeRange,
  selectedCategory,
  onTrendSelect,
  onCategoryChange,
  className,
}) => {
  const [hoveredTrend, setHoveredTrend] = useState<string | null>(null);

  const categories = useMemo(() => {
    return Array.from(new Set(trends.map(t => t.category)));
  }, [trends]);

  const getDirectionIcon = (direction: TrendDirection) => {
    switch (direction) {
      case 'up': return <ArrowUpRight className="w-3 h-3 text-green-400" />;
      case 'down': return <ArrowDownRight className="w-3 h-3 text-red-400" />;
      default: return <Minus className="w-3 h-3 text-aviation-text-muted" />;
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Radar className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Trend Radar</h3>
          </div>
          <div className="flex items-center gap-1">
            {(['7d', '30d', '90d'] as const).map(t => (
              <button
                key={t}
                onClick={() => onCategoryChange?.(t)}
                className={cn(
                  'px-2 py-1 text-xs rounded transition-colors',
                  timeRange === t
                    ? 'bg-aviation-cyan text-aviation-bg-primary'
                    : 'text-aviation-text-muted hover:text-aviation-text-primary'
                )}
              >
                {t}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center gap-2 overflow-x-auto">
          <button
            onClick={() => onCategoryChange?.('')}
            className={cn(
              'px-2 py-1 text-xs rounded whitespace-nowrap transition-colors',
              !selectedCategory
                ? 'bg-aviation-cyan text-aviation-bg-primary'
                : 'text-aviation-text-muted hover:text-aviation-text-primary'
            )}
          >
            All
          </button>
          {categories.map(cat => (
            <button
              key={cat}
              onClick={() => onCategoryChange?.(cat)}
              className={cn(
                'px-2 py-1 text-xs rounded whitespace-nowrap transition-colors',
                selectedCategory === cat
                  ? 'bg-aviation-cyan text-aviation-bg-primary'
                  : 'text-aviation-text-muted hover:text-aviation-text-primary'
              )}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        <div className="space-y-3">
          {trends.map(trend => {
            const isHovered = hoveredTrend === trend.id;
            return (
              <div
                key={trend.id}
                onClick={() => onTrendSelect?.(trend)}
                onMouseEnter={() => setHoveredTrend(trend.id)}
                onMouseLeave={() => setHoveredTrend(null)}
                className={cn(
                  'p-4 rounded-lg border transition-all cursor-pointer',
                  isHovered
                    ? 'border-aviation-cyan bg-aviation-bg-instrument'
                    : 'border-aviation-border-panel bg-aviation-bg-secondary'
                )}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-aviation-text-primary">{trend.name}</span>
                    {getDirectionIcon(trend.direction)}
                  </div>
                  <div className="flex items-center gap-2 text-[10px]">
                    <span className="text-aviation-text-muted">Vol: {trend.searchVolume.toLocaleString()}</span>
                    <span className="text-aviation-text-muted">${trend.avgPrice}</span>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <span className="px-2 py-0.5 bg-aviation-bg-instrument text-aviation-text-muted text-[10px] rounded uppercase">
                    {trend.category}
                  </span>
                  <div className="flex items-center gap-1">
                    <span className="text-[10px] text-aviation-text-dim">Momentum:</span>
                    <div className="w-16 h-1.5 bg-aviation-bg-instrument rounded-full overflow-hidden">
                      <div
                        className={cn(
                          'h-full',
                          trend.momentum > 0 ? 'bg-green-400' : trend.momentum < 0 ? 'bg-red-400' : 'bg-aviation-text-muted'
                        )}
                        style={{ width: `${Math.min(100, Math.abs(trend.momentum) * 100)}%` }}
                      />
                    </div>
                  </div>
                </div>
                {isHovered && trend.relatedFunctions.length > 0 && (
                  <div className="mt-3 pt-3 border-t border-aviation-border-panel">
                    <div className="text-[10px] text-aviation-text-dim mb-1">Related Functions</div>
                    <div className="flex flex-wrap gap-1">
                      {trend.relatedFunctions.slice(0, 3).map((fn, i) => (
                        <span key={i} className="px-2 py-0.5 bg-aviation-bg-panel text-aviation-text-muted text-[10px] rounded">
                          {fn}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};
