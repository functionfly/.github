import { getTrustColorConfig, getTrustScoreBand } from '@/components/functions/TrustScoreBadge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import {
  Activity,
  AlertTriangle,
  Minus,
  Shield,
  Target,
  TrendingDown,
  TrendingUp,
  Users,
  Zap,
  type LucideIcon,
} from 'lucide-react';
import { TrustBadge, type VerificationLevel } from './TrustBadge';
import { TrustHistory, type TrustHistoryDataPoint } from './TrustHistory';
import { TrustScore } from './TrustScore';
import { VerificationStatus, type VerificationItem } from './VerificationStatus';

/**
 * Trust metrics summary
 */
export interface TrustMetricsSummary {
  /** Overall trust score */
  overallScore: number;
  /** Previous score for trend */
  previousScore?: number;
  /** Reliability component score */
  reliability?: number;
  /** Latency component score */
  latency?: number;
  /** Determinism component score */
  determinism?: number;
  /** Community reputation score */
  communityReputation?: number;
  /** Fraud risk level */
  fraudRisk?: 'low' | 'medium' | 'high';
}

/**
 * TrustDashboardWidget Component Props
 */
export interface TrustDashboardWidgetProps {
  /** Function identifier */
  functionId?: string;
  /** Function name for display */
  functionName?: string;
  /** Trust metrics summary */
  metrics: TrustMetricsSummary;
  /** Verification items */
  verificationItems?: VerificationItem[];
  /** Historical trust data */
  historyData?: TrustHistoryDataPoint[];
  /** Display variant */
  variant?: 'compact' | 'standard' | 'detailed';
  /** Show trend chart */
  showChart?: boolean;
  /** Show verification status */
  showVerification?: boolean;
  /** Height of the chart */
  chartHeight?: number;
  /** Additional CSS classes */
  className?: string;
  /** Loading state */
  loading?: boolean;
}

/**
 * Metric card for dashboard
 */
function MetricCard({
  label,
  value,
  icon: Icon,
  colorClass,
  trend,
}: {
  label: string;
  value: number;
  icon: LucideIcon;
  colorClass: string;
  trend?: 'up' | 'down' | 'stable';
}) {
  const TrendIcon = trend === 'up' ? TrendingUp : trend === 'down' ? TrendingDown : Minus;

  return (
    <div
      className={cn(
        'flex items-center gap-3 p-3 rounded-lg border border-border/50',
        'bg-muted/30'
      )}
    >
      <div className={cn('p-2 rounded-lg', colorClass, 'bg-current/10')}>
        <Icon className={cn('h-4 w-4', colorClass)} aria-hidden="true" />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-[10px] uppercase tracking-wider text-muted-foreground truncate">
          {label}
        </p>
        <div className="flex items-center gap-1.5">
          <span className={cn('text-lg font-bold', colorClass)}>{Math.round(value)}%</span>
          {trend && (
            <TrendIcon
              className={cn(
                'h-3 w-3',
                trend === 'up' && 'text-emerald-400',
                trend === 'down' && 'text-red-400',
                trend === 'stable' && 'text-gray-400'
              )}
            />
          )}
        </div>
      </div>
    </div>
  );
}

/**
 * Compact variant - single trust score with badge
 */
function CompactWidget({
  metrics,
  verificationItems,
  functionName,
  className,
}: {
  metrics: TrustMetricsSummary;
  verificationItems?: VerificationItem[];
  functionName?: string;
  className?: string;
}) {
  const band = getTrustScoreBand(metrics.overallScore);
  const colorConfig = getTrustColorConfig(band);

  // Derive verification level from verification items
  let verificationLevel: VerificationLevel = 'basic';
  if (verificationItems) {
    const allVerified = verificationItems.every((i) => i.status === 'verified');
    const anyFailed = verificationItems.some((i) => i.status === 'failed');
    const anyPending = verificationItems.some((i) => i.status === 'pending');

    if (allVerified) verificationLevel = 'highly_trusted';
    else if (anyFailed) verificationLevel = 'untrusted';
    else if (anyPending) verificationLevel = 'pending';
    else if (verificationItems.length > 0) verificationLevel = 'verified';
  }

  return (
    <Card className={cn('w-full', className)}>
      <CardContent className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <TrustScore
              score={metrics.overallScore}
              previousScore={metrics.previousScore}
              variant="compact"
              size="md"
              showTrend
            />
            <div>
              <p className="text-sm font-medium">{colorConfig.label} Trust</p>
              {functionName && (
                <p className="text-xs text-muted-foreground truncate max-w-[150px]">
                  {functionName}
                </p>
              )}
            </div>
          </div>
          <TrustBadge level={verificationLevel} variant="inline" size="sm" />
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Standard variant - metrics grid with verification
 */
function StandardWidget({
  metrics,
  verificationItems,
  functionName,
  showVerification,
  className,
}: {
  metrics: TrustMetricsSummary;
  verificationItems?: VerificationItem[];
  functionName?: string;
  showVerification?: boolean;
  className?: string;
}) {
  const band = getTrustScoreBand(metrics.overallScore);
  const colorConfig = getTrustColorConfig(band);

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Shield className={cn('h-5 w-5', colorConfig.text)} aria-hidden="true" />
            <CardTitle className="text-base font-medium">Trust Overview</CardTitle>
          </div>
          {functionName && (
            <span className="text-xs text-muted-foreground truncate max-w-[200px]">
              {functionName}
            </span>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Main score */}
        <div className="flex items-center justify-between">
          <TrustScore
            score={metrics.overallScore}
            previousScore={metrics.previousScore}
            variant="detailed"
            size="lg"
            showTrend
            showTier
          />
          {metrics.fraudRisk && (
            <div
              className={cn(
                'flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium',
                metrics.fraudRisk === 'low' && 'bg-emerald-500/10 text-emerald-400',
                metrics.fraudRisk === 'medium' && 'bg-amber-500/10 text-amber-400',
                metrics.fraudRisk === 'high' && 'bg-red-500/10 text-red-400'
              )}
            >
              <AlertTriangle className="h-3.5 w-3.5" />
              {metrics.fraudRisk.charAt(0).toUpperCase() + metrics.fraudRisk.slice(1)} Risk
            </div>
          )}
        </div>

        {/* Metrics grid */}
        <div className="grid grid-cols-2 gap-2">
          {metrics.reliability !== undefined && (
            <MetricCard
              label="Reliability"
              value={metrics.reliability}
              icon={Activity}
              colorClass="text-blue-400"
            />
          )}
          {metrics.latency !== undefined && (
            <MetricCard
              label="Latency"
              value={metrics.latency}
              icon={Zap}
              colorClass="text-amber-400"
            />
          )}
          {metrics.determinism !== undefined && (
            <MetricCard
              label="Determinism"
              value={metrics.determinism}
              icon={Target}
              colorClass="text-violet-400"
            />
          )}
          {metrics.communityReputation !== undefined && (
            <MetricCard
              label="Community"
              value={metrics.communityReputation}
              icon={Users}
              colorClass="text-cyan-400"
            />
          )}
        </div>

        {/* Verification status */}
        {showVerification && verificationItems && verificationItems.length > 0 && (
          <div className="pt-3 border-t border-border/50">
            <p className="text-xs font-medium text-muted-foreground mb-2">Verification</p>
            <VerificationStatus items={verificationItems} variant="summary" size="sm" />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

/**
 * Detailed variant - includes chart
 */
function DetailedWidget({
  metrics,
  verificationItems,
  historyData,
  functionName,
  showVerification,
  chartHeight = 200,
  className,
}: {
  metrics: TrustMetricsSummary;
  verificationItems?: VerificationItem[];
  historyData?: TrustHistoryDataPoint[];
  functionName?: string;
  showVerification?: boolean;
  chartHeight?: number;
  className?: string;
}) {
  const band = getTrustScoreBand(metrics.overallScore);
  const colorConfig = getTrustColorConfig(band);

  return (
    <div className={cn('space-y-4', className)}>
      {/* Main widget */}
      <StandardWidget
        metrics={metrics}
        verificationItems={verificationItems}
        functionName={functionName}
        showVerification={showVerification}
      />

      {/* History chart */}
      {historyData && historyData.length > 0 && (
        <TrustHistory
          data={historyData}
          title="Trust Score Trend"
          variant="area"
          showComponents
          showTrend
          height={chartHeight}
        />
      )}
    </div>
  );
}

/**
 * TrustDashboardWidget Component
 *
 * Comprehensive trust metrics widget for the dashboard.
 * Displays trust score, component metrics, verification status, and optional trend chart.
 *
 * @example
 * // Compact single-line widget
 * <TrustDashboardWidget
 *   metrics={trustMetrics}
 *   variant="compact"
 * />
 *
 * // Standard with verification
 * <TrustDashboardWidget
 *   metrics={trustMetrics}
 *   verificationItems={items}
 *   variant="standard"
 *   showVerification
 * />
 *
 * // Detailed with chart
 * <TrustDashboardWidget
 *   metrics={trustMetrics}
 *   historyData={history}
 *   variant="detailed"
 *   showVerification
 * />
 */
export function TrustDashboardWidget({
  functionId,
  functionName,
  metrics,
  verificationItems,
  historyData,
  variant = 'standard',
  showChart = false,
  showVerification = false,
  chartHeight = 200,
  className,
  loading = false,
}: TrustDashboardWidgetProps) {
  if (loading) {
    return (
      <Card className={cn('w-full', className)}>
        <CardHeader className="pb-3">
          <div className="h-5 w-32 bg-muted rounded animate-pulse" />
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <div
              className="animate-pulse bg-muted rounded-full"
              style={{ width: 64, height: 64 }}
            />
            <div className="space-y-2">
              <div className="h-4 w-24 bg-muted rounded animate-pulse" />
              <div className="h-3 w-16 bg-muted rounded animate-pulse" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="h-16 bg-muted/50 rounded-lg animate-pulse" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (variant === 'compact') {
    return (
      <CompactWidget
        metrics={metrics}
        verificationItems={verificationItems}
        functionName={functionName}
        className={className}
      />
    );
  }

  if (variant === 'detailed' || showChart) {
    return (
      <DetailedWidget
        metrics={metrics}
        verificationItems={verificationItems}
        historyData={historyData}
        functionName={functionName}
        showVerification={showVerification}
        chartHeight={chartHeight}
        className={className}
      />
    );
  }

  return (
    <StandardWidget
      metrics={metrics}
      verificationItems={verificationItems}
      functionName={functionName}
      showVerification={showVerification}
      className={className}
    />
  );
}

/**
 * TrustDashboardWidgetSkeleton Component
 * Loading placeholder for TrustDashboardWidget
 */
export function TrustDashboardWidgetSkeleton({
  variant = 'standard',
  className,
}: {
  variant?: 'compact' | 'standard' | 'detailed';
  className?: string;
}) {
  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-3">
        <div className="h-5 w-32 bg-muted rounded animate-pulse" />
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-4">
          <div className="animate-pulse bg-muted rounded-full" style={{ width: 64, height: 64 }} />
          <div className="space-y-2">
            <div className="h-4 w-24 bg-muted rounded animate-pulse" />
            <div className="h-3 w-16 bg-muted rounded animate-pulse" />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-16 bg-muted/50 rounded-lg animate-pulse" />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

export default TrustDashboardWidget;
