import { Progress } from '@/components/ui/progress';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { AlertTriangle, CheckCircle, Circle, Clock, XCircle, type LucideIcon } from 'lucide-react';
import React from 'react';

/**
 * Verification status types
 */
export type VerificationStatusType =
  | 'verified'
  | 'pending'
  | 'in_progress'
  | 'failed'
  | 'expired'
  | 'not_started';

/**
 * Individual verification item
 */
export interface VerificationItem {
  /** Unique identifier */
  id: string;
  /** Display label */
  label: string;
  /** Verification status */
  status: VerificationStatusType;
  /** Optional: completion percentage (0-100) */
  progress?: number;
  /** Optional: timestamp of last update */
  lastUpdated?: Date;
  /** Optional: notes or details */
  details?: string;
}

/**
 * VerificationStatus Component Props
 */
export interface VerificationStatusProps {
  /** List of verification items */
  items: VerificationItem[];
  /** Overall verification status (derived from items if not provided) */
  overallStatus?: VerificationStatusType;
  /** Display variant */
  variant?: 'compact' | 'detailed' | 'summary';
  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  /** Show timestamps */
  showTimestamps?: boolean;
  /** Additional CSS classes */
  className?: string;
}

/**
 * Get status configuration
 */
function getStatusConfig(status: VerificationStatusType): {
  icon: LucideIcon;
  colorClass: string;
  bgClass: string;
  label: string;
  progressColor: string;
} {
  const configs = {
    verified: {
      icon: CheckCircle,
      colorClass: 'text-emerald-400',
      bgClass: 'bg-emerald-500/10',
      label: 'Verified',
      progressColor: 'bg-emerald-500',
    },
    pending: {
      icon: Clock,
      colorClass: 'text-amber-400',
      bgClass: 'bg-amber-500/10',
      label: 'Pending',
      progressColor: 'bg-amber-500',
    },
    in_progress: {
      icon: Circle,
      colorClass: 'text-blue-400',
      bgClass: 'bg-blue-500/10',
      label: 'In Progress',
      progressColor: 'bg-blue-500',
    },
    failed: {
      icon: XCircle,
      colorClass: 'text-red-400',
      bgClass: 'bg-red-500/10',
      label: 'Failed',
      progressColor: 'bg-red-500',
    },
    expired: {
      icon: AlertTriangle,
      colorClass: 'text-orange-400',
      bgClass: 'bg-orange-500/10',
      label: 'Expired',
      progressColor: 'bg-orange-500',
    },
    not_started: {
      icon: Circle,
      colorClass: 'text-gray-400',
      bgClass: 'bg-gray-500/10',
      label: 'Not Started',
      progressColor: 'bg-gray-500',
    },
  };
  return configs[status];
}

/**
 * Derive overall status from items
 */
function deriveOverallStatus(items: VerificationItem[]): VerificationStatusType {
  if (items.every((item) => item.status === 'verified')) return 'verified';
  if (items.some((item) => item.status === 'failed')) return 'failed';
  if (items.some((item) => item.status === 'expired')) return 'expired';
  if (items.some((item) => item.status === 'in_progress')) return 'in_progress';
  if (items.some((item) => item.status === 'pending')) return 'pending';
  return 'not_started';
}

/**
 * Compact verification indicator (single status)
 */
function CompactVariant({
  status,
  size,
  className,
}: {
  status: VerificationStatusType;
  size: 'sm' | 'md' | 'lg';
  className?: string;
}) {
  const config = getStatusConfig(status);
  const Icon = config.icon;

  const sizes = {
    sm: 'h-4 w-4',
    md: 'h-5 w-5',
    lg: 'h-6 w-6',
  };

  return (
    <div className={cn('inline-flex', className)}>
      <Icon className={cn(sizes[size], config.colorClass)} aria-label={config.label} />
    </div>
  );
}

/**
 * Summary verification status (single row)
 */
function SummaryVariant({
  items,
  overallStatus,
  size,
  showTimestamps,
  className,
}: {
  items: VerificationItem[];
  overallStatus?: VerificationStatusType;
  size: 'sm' | 'md' | 'lg';
  showTimestamps?: boolean;
  className?: string;
}) {
  const status = overallStatus || deriveOverallStatus(items);
  const config = getStatusConfig(status);
  const Icon = config.icon;

  const completedCount = items.filter((i) => i.status === 'verified').length;
  const totalCount = items.length;

  const sizes = {
    sm: { icon: 'h-3.5 w-3.5', text: 'text-xs' },
    md: { icon: 'h-4 w-4', text: 'text-sm' },
    lg: { icon: 'h-5 w-5', text: 'text-base' },
  };

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Icon className={cn(sizes[size].icon, config.colorClass)} aria-hidden="true" />
      <span className={cn(sizes[size].text, 'font-medium', config.colorClass)}>{config.label}</span>
      <span className={cn(sizes[size].text, 'text-muted-foreground')}>
        ({completedCount}/{totalCount})
      </span>
    </div>
  );
}

/**
 * Detailed verification list
 */
function DetailedVariant({
  items,
  size,
  showTimestamps,
  className,
}: {
  items: VerificationItem[];
  size: 'sm' | 'md' | 'lg';
  showTimestamps?: boolean;
  className?: string;
}) {
  const spacing = size === 'sm' ? 'space-y-1.5' : size === 'md' ? 'space-y-2' : 'space-y-3';

  return (
    <div className={cn(spacing, className)}>
      {items.map((item) => {
        const config = getStatusConfig(item.status);
        const Icon = config.icon;

        return (
          <div
            key={item.id}
            className={cn(
              'flex items-start gap-3 p-2 rounded-lg border border-border/50',
              'bg-muted/30'
            )}
          >
            <div className="mt-0.5">
              <Icon className={cn('h-4 w-4', config.colorClass)} aria-hidden="true" />
            </div>

            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between gap-2">
                <span
                  className={cn(
                    'font-medium truncate',
                    size === 'sm' && 'text-xs',
                    size === 'md' && 'text-sm',
                    size === 'lg' && 'text-base'
                  )}
                >
                  {item.label}
                </span>
                <span
                  className={cn(
                    'font-medium flex-shrink-0',
                    size === 'sm' && 'text-[10px]',
                    size === 'md' && 'text-xs',
                    size === 'lg' && 'text-sm',
                    config.colorClass
                  )}
                >
                  {config.label}
                </span>
              </div>

              {item.progress !== undefined && item.status === 'in_progress' && (
                <div className="mt-1.5">
                  <Progress value={item.progress} className="h-1" />
                  <span className="text-[10px] text-muted-foreground mt-0.5">
                    {item.progress}% complete
                  </span>
                </div>
              )}

              {item.details && (
                <p
                  className={cn(
                    'text-muted-foreground mt-0.5',
                    size === 'sm' && 'text-[10px]',
                    size === 'md' && 'text-xs',
                    size === 'lg' && 'text-sm'
                  )}
                >
                  {item.details}
                </p>
              )}

              {showTimestamps && item.lastUpdated && (
                <span
                  className={cn(
                    'text-muted-foreground/70 block mt-0.5',
                    size === 'sm' && 'text-[9px]',
                    size === 'md' && 'text-[10px]',
                    size === 'lg' && 'text-xs'
                  )}
                >
                  Updated {item.lastUpdated.toLocaleDateString()}
                </span>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

/**
 * VerificationStatus Component
 *
 * Displays verification status with support for multiple verification items.
 * Supports three variants: compact (icon), summary (single row), detailed (list).
 *
 * @example
 * // Summary status
 * <VerificationStatus
 *   items={verificationItems}
 *   variant="summary"
 * />
 *
 * // Detailed list
 * <VerificationStatus
 *   items={verificationItems}
 *   variant="detailed"
 *   showTimestamps
 * />
 *
 * // Compact icon
 * <VerificationStatus items={verificationItems} variant="compact" />
 */
export function VerificationStatus({
  items,
  overallStatus,
  variant = 'summary',
  size = 'md',
  showTimestamps = false,
  className,
}: VerificationStatusProps) {
  const content =
    variant === 'compact' ? (
      <CompactVariant
        status={overallStatus || deriveOverallStatus(items)}
        size={size}
        className={className}
      />
    ) : variant === 'summary' ? (
      <SummaryVariant
        items={items}
        overallStatus={overallStatus}
        size={size}
        showTimestamps={showTimestamps}
        className={className}
      />
    ) : (
      <DetailedVariant
        items={items}
        size={size}
        showTimestamps={showTimestamps}
        className={className}
      />
    );

  // Wrap with tooltip for detailed info
  if (variant !== 'detailed' && items.length > 0) {
    return (
      <TooltipProvider delayDuration={200}>
        <Tooltip>
          <TooltipTrigger asChild>{content}</TooltipTrigger>
          <TooltipContent side="top" className="max-w-xs">
            <div className="space-y-1">
              <p className="font-semibold">Verification Status</p>
              {items.map((item) => (
                <div key={item.id} className="flex items-center gap-2 text-xs">
                  {getStatusConfig(item.status).icon &&
                    React.createElement(getStatusConfig(item.status).icon, {
                      className: cn('h-3 w-3', getStatusConfig(item.status).colorClass),
                    })}
                  <span className="text-muted-foreground">{item.label}:</span>
                  <span className={getStatusConfig(item.status).colorClass}>
                    {getStatusConfig(item.status).label}
                  </span>
                </div>
              ))}
            </div>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  return content;
}

/**
 * VerificationStatusSkeleton Component
 * Loading placeholder for VerificationStatus
 */
export function VerificationStatusSkeleton({
  variant = 'summary',
  count = 3,
  className,
}: {
  variant?: 'compact' | 'detailed' | 'summary';
  count?: number;
  className?: string;
}) {
  if (variant === 'compact') {
    return (
      <div
        className={cn('animate-pulse bg-muted rounded', className)}
        style={{ width: 20, height: 20 }}
        aria-hidden="true"
      />
    );
  }

  if (variant === 'summary') {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <div className="animate-pulse bg-muted rounded" style={{ width: 20, height: 20 }} />
        <div className="h-4 w-24 bg-muted rounded animate-pulse" />
      </div>
    );
  }

  return (
    <div className={cn('space-y-2', className)}>
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="flex items-start gap-3 p-2 rounded-lg bg-muted/30">
          <div className="animate-pulse bg-muted rounded" style={{ width: 16, height: 16 }} />
          <div className="flex-1 space-y-1.5">
            <div className="h-3 w-32 bg-muted rounded animate-pulse" />
            <div className="h-2 w-full bg-muted rounded animate-pulse" />
          </div>
        </div>
      ))}
    </div>
  );
}

export default VerificationStatus;
