import { getTrustColorConfig, getTrustScoreBand } from '@/components/functions/TrustScoreBadge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { Minus, Shield, TrendingDown, TrendingUp, type LucideIcon } from 'lucide-react';
import React, { useMemo } from 'react';

/**
 * Trust score change direction
 */
export type TrustTrend = 'up' | 'down' | 'stable';

/**
 * TrustScore Component Props
 */
export interface TrustScoreProps {
  /** Overall trust score (0-100) */
  score: number;
  /** Previous score for trend calculation */
  previousScore?: number;
  /** Display variant */
  variant?: 'compact' | 'detailed' | 'progress';
  /** Show trend indicator */
  showTrend?: boolean;
  /** Show tier label */
  showTier?: boolean;
  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  /** Additional CSS classes */
  className?: string;
  /** Show tooltip with breakdown */
  showTooltip?: boolean;
  /** Custom tooltip content */
  tooltipContent?: React.ReactNode;
}

/**
 * Get trend configuration
 */
function getTrendConfig(trend: TrustTrend): {
  icon: LucideIcon;
  colorClass: string;
  label: string;
} {
  const configs = {
    up: {
      icon: TrendingUp,
      colorClass: 'text-emerald-400',
      label: 'Improving',
    },
    down: {
      icon: TrendingDown,
      colorClass: 'text-red-400',
      label: 'Declining',
    },
    stable: {
      icon: Minus,
      colorClass: 'text-gray-400',
      label: 'Stable',
    },
  };
  return configs[trend];
}

/**
 * Calculate trend from score change
 */
function calculateTrend(current: number, previous?: number): TrustTrend {
  if (!previous) return 'stable';
  const diff = current - previous;
  if (diff > 2) return 'up';
  if (diff < -2) return 'down';
  return 'stable';
}

/**
 * Circular progress indicator for trust score
 */
function CircularProgress({
  score,
  color,
  size = 48,
  strokeWidth = 4,
}: {
  score: number;
  color: string;
  size?: number;
  strokeWidth?: number;
}) {
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset =
    circumference - (Math.max(0, Math.min(100, score)) / 100) * circumference;

  return (
    <svg width={size} height={size} className="transform -rotate-90" aria-hidden="true">
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke="currentColor"
        strokeWidth={strokeWidth}
        className="text-muted/30"
      />
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke={color}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeDasharray={circumference}
        strokeDashoffset={strokeDashoffset}
        className="transition-all duration-700 ease-out"
      />
    </svg>
  );
}

/**
 * Progress bar variant
 */
function ProgressVariant({
  score,
  colorConfig,
  size,
}: {
  score: number;
  colorConfig: ReturnType<typeof getTrustColorConfig>;
  size: 'sm' | 'md' | 'lg';
}) {
  const heights = {
    sm: 'h-1.5',
    md: 'h-2',
    lg: 'h-3',
  };

  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">Trust Score</span>
        <span className={cn('font-semibold', colorConfig.text)}>{Math.round(score)}%</span>
      </div>
      <div className={cn('w-full bg-muted/30 rounded-full overflow-hidden', heights[size])}>
        <div
          className={cn('h-full rounded-full transition-all duration-700 ease-out', colorConfig.bg)}
          style={{ width: `${Math.max(0, Math.min(100, score))}%` }}
        />
      </div>
    </div>
  );
}

/**
 * TrustScore Component
 *
 * Displays trust score with optional tier indicator and trend.
 * Supports three variants: compact (circular), detailed (with breakdown), progress (bar).
 *
 * @example
 * // Compact circular score
 * <TrustScore score={87} variant="compact" size="md" />
 *
 * // Detailed with trend
 * <TrustScore score={92} previousScore={88} variant="detailed" showTrend showTier />
 *
 * // Progress bar
 * <TrustScore score={75} variant="progress" size="lg" />
 */
export function TrustScore({
  score,
  previousScore,
  variant = 'compact',
  showTrend = false,
  showTier = false,
  size = 'md',
  className,
  showTooltip = false,
  tooltipContent,
}: TrustScoreProps) {
  const band = useMemo(() => getTrustScoreBand(score), [score]);
  const colorConfig = useMemo(() => getTrustColorConfig(band), [band]);
  const trend = useMemo(() => calculateTrend(score, previousScore), [score, previousScore]);
  const trendConfig = useMemo(() => getTrendConfig(trend), [trend]);
  const TrendIcon = trendConfig.icon;

  const sizeConfigs = {
    sm: { fontSize: 'text-sm', iconSize: 'h-3 w-3' },
    md: { fontSize: 'text-base', iconSize: 'h-4 w-4' },
    lg: { fontSize: 'text-xl', iconSize: 'h-5 w-5' },
  };

  const circularSizes = {
    sm: { size: 32, strokeWidth: 3 },
    md: { size: 48, strokeWidth: 4 },
    lg: { size: 64, strokeWidth: 5 },
  };

  const renderCompact = () => (
    <div className={cn('relative inline-flex items-center justify-center', className)}>
      <CircularProgress score={score} color={colorConfig.primary} {...circularSizes[size]} />
      <div className="absolute inset-0 flex items-center justify-center">
        <Shield className={cn(colorConfig.text, sizeConfigs[size].iconSize)} aria-hidden="true" />
      </div>
    </div>
  );

  const renderDetailed = () => (
    <div className={cn('flex items-center gap-3', className)}>
      <div className="relative">
        <CircularProgress score={score} color={colorConfig.primary} {...circularSizes[size]} />
        <div className="absolute inset-0 flex items-center justify-center">
          <span className={cn('font-bold', sizeConfigs[size].fontSize, colorConfig.text)}>
            {Math.round(score)}
          </span>
        </div>
      </div>

      <div className="flex flex-col gap-0.5">
        {showTier && (
          <span className={cn('font-semibold text-sm', colorConfig.text)}>
            {colorConfig.label} Trust
          </span>
        )}
        <span className="text-xs text-muted-foreground">
          {colorConfig.label} ({Math.round(score)}%)
        </span>

        {showTrend && previousScore !== undefined && (
          <div className={cn('flex items-center gap-1 mt-0.5', trendConfig.colorClass)}>
            <TrendIcon className="h-3 w-3" aria-hidden="true" />
            <span className="text-[10px] font-medium">
              {trendConfig.label} ({Math.abs(score - previousScore).toFixed(1)})
            </span>
          </div>
        )}
      </div>
    </div>
  );

  const renderProgress = () => (
    <div className={cn('w-full', className)}>
      <ProgressVariant score={score} colorConfig={colorConfig} size={size} />
      {showTier && (
        <span className={cn('text-xs font-medium mt-1 block', colorConfig.text)}>
          {colorConfig.label}
        </span>
      )}
    </div>
  );

  const content =
    variant === 'compact'
      ? renderCompact()
      : variant === 'detailed'
        ? renderDetailed()
        : renderProgress();

  if (showTooltip && tooltipContent) {
    return (
      <TooltipProvider delayDuration={200}>
        <Tooltip>
          <TooltipTrigger asChild>{content}</TooltipTrigger>
          <TooltipContent side="top" className="max-w-xs">
            {tooltipContent}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  return content;
}

/**
 * TrustScoreSkeleton Component
 * Loading placeholder for TrustScore
 */
export function TrustScoreSkeleton({
  variant = 'compact',
  className,
}: {
  variant?: 'compact' | 'detailed' | 'progress';
  className?: string;
}) {
  if (variant === 'compact') {
    return (
      <div
        className={cn('animate-pulse bg-muted rounded-full', className)}
        style={{ width: 48, height: 48 }}
        aria-hidden="true"
      />
    );
  }

  if (variant === 'detailed') {
    return (
      <div className={cn('flex items-center gap-3', className)}>
        <div className="animate-pulse bg-muted rounded-full" style={{ width: 48, height: 48 }} />
        <div className="space-y-1.5">
          <div className="h-4 w-24 bg-muted rounded animate-pulse" />
          <div className="h-3 w-16 bg-muted rounded animate-pulse" />
        </div>
      </div>
    );
  }

  return (
    <div className={cn('space-y-1.5', className)}>
      <div className="h-4 w-full bg-muted rounded animate-pulse" />
      <div className="h-2 w-full bg-muted rounded-full animate-pulse" />
    </div>
  );
}

export default TrustScore;
