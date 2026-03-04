/**
 * ExecutionStatusIndicator - Running/success/error states
 */

import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import {
  Loader2,
  CheckCircle2,
  XCircle,
  Clock,
  Play,
  AlertCircle,
  Check,
} from 'lucide-react';
import type { ExecutionStatus } from '../types';

interface ExecutionStatusIndicatorProps {
  status: ExecutionStatus;
  progress?: number;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

const statusConfig: Record<ExecutionStatus, {
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  color: string;
  bgColor: string;
  animate?: boolean;
}> = {
  idle: {
    label: 'Ready',
    icon: Play,
    color: 'text-slate-400',
    bgColor: 'bg-slate-800',
  },
  pending: {
    label: 'Pending',
    icon: Clock,
    color: 'text-slate-400',
    bgColor: 'bg-slate-800',
    animate: true,
  },
  queued: {
    label: 'Queued',
    icon: Clock,
    color: 'text-amber-400',
    bgColor: 'bg-amber-500/10',
    animate: true,
  },
  running: {
    label: 'Running',
    icon: Loader2,
    color: 'text-blue-400',
    bgColor: 'bg-blue-500/10',
    animate: true,
  },
  completed: {
    label: 'Success',
    icon: CheckCircle2,
    color: 'text-emerald-400',
    bgColor: 'bg-emerald-500/10',
  },
  failed: {
    label: 'Failed',
    icon: XCircle,
    color: 'text-red-400',
    bgColor: 'bg-red-500/10',
  },
};

export function ExecutionStatusIndicator({
  status,
  progress,
  className,
  size = 'md',
}: ExecutionStatusIndicatorProps) {
  const config = statusConfig[status];
  const Icon = config.icon;

  const sizeClasses = {
    sm: 'h-2 w-2',
    md: 'h-4 w-4',
    lg: 'h-6 w-6',
  };

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div className={cn(
        'flex items-center justify-center rounded-full',
        config.bgColor,
        size === 'sm' ? 'p-1' : 'p-2'
      )}>
        <Icon className={cn(
          sizeClasses[size],
          config.color,
          config.animate && 'animate-spin'
        )} />
      </div>

      <div className="flex flex-col">
        <span className={cn(
          'font-medium',
          config.color,
          size === 'sm' ? 'text-xs' : 'text-sm'
        )}>
          {config.label}
        </span>

        {status === 'running' && progress !== undefined && (
          <div className="flex items-center gap-2">
            <Progress
              value={progress}
              className="h-1 w-24 bg-slate-800"
            />
            <span className="text-xs text-slate-500">{progress}%</span>
          </div>
        )}
      </div>
    </div>
  );
}

/**
 * Compact badge version for inline display
 */
export function ExecutionStatusBadge({
  status,
  className,
}: {
  status: ExecutionStatus;
  className?: string;
}) {
  const config = statusConfig[status];

  return (
    <Badge
      variant="outline"
      className={cn(
        'gap-1.5 font-medium',
        config.bgColor,
        config.color.replace('text-', 'border-').replace('400', '500/30'),
        config.color,
        className
      )}
    >
      <config.icon className={cn('h-3 w-3', config.animate && 'animate-spin')} />
      {config.label}
    </Badge>
  );
}

/**
 * Verified status indicator
 */
export function VerifiedStatusIndicator({
  isVerified,
  score,
  className,
}: {
  isVerified: boolean;
  score?: number;
  className?: string;
}) {
  if (!isVerified) {
    return (
      <div className={cn('flex items-center gap-2 text-slate-500', className)}>
        <AlertCircle className="h-4 w-4" />
        <span className="text-sm">Not Verified</span>
      </div>
    );
  }

  return (
    <div className={cn('flex items-center gap-2 text-indigo-400', className)}>
      <div className="flex h-5 w-5 items-center justify-center rounded-full bg-indigo-500/10">
        <Check className="h-3 w-3" />
      </div>
      <span className="text-sm font-medium">
        Verified {score !== undefined && `(${score}/100)`}
      </span>
    </div>
  );
}
