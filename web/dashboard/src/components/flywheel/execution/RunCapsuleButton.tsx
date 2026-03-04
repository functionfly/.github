/**
 * RunCapsuleButton - Execute code button with various states
 */

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Play,
  RotateCcw,
  Loader2,
  Check,
  X,
  Sparkles,
} from 'lucide-react';
import type { ExecutionStatus } from '../types';

interface RunCapsuleButtonProps {
  status: ExecutionStatus;
  onRun: () => void;
  onVerify?: () => void;
  disabled?: boolean;
  size?: 'sm' | 'default' | 'lg';
  variant?: 'default' | 'outline' | 'ghost';
  showVerify?: boolean;
  className?: string;
}

export function RunCapsuleButton({
  status,
  onRun,
  onVerify,
  disabled = false,
  size = 'default',
  variant = 'default',
  showVerify = false,
  className,
}: RunCapsuleButtonProps) {
  const isRunning = status === 'running' || status === 'queued' || status === 'pending';
  const isCompleted = status === 'completed';
  const isFailed = status === 'failed';

  const sizeClasses = {
    sm: 'h-8 text-xs',
    md: 'h-9 text-sm',
    lg: 'h-10 text-sm',
  };

  const iconSizes = {
    sm: 'h-3.5 w-3.5',
    md: 'h-4 w-4',
    lg: 'h-4 w-4',
  };

  // Running state
  if (isRunning) {
    return (
      <Button
        disabled
        size={size}
        className={cn(
          'gap-2 bg-blue-500/20 text-blue-400',
          sizeClasses[size],
          className
        )}
      >
        <Loader2 className={cn(iconSizes[size], 'animate-spin')} />
        {status === 'queued' ? 'Queued...' : 'Running...'}
      </Button>
    );
  }

  // Show Run and Verify buttons side by side
  if (showVerify && onVerify && isCompleted) {
    return (
      <div className="flex items-center gap-2">
        <Button
          onClick={onRun}
          size={size}
          variant="outline"
          className={cn(
            'gap-2 border-slate-700 text-slate-300 hover:bg-slate-800',
            sizeClasses[size],
            className
          )}
        >
          <RotateCcw className={iconSizes[size]} />
          Run Again
        </Button>
        <Button
          onClick={onVerify}
          size={size}
          className={cn(
            'gap-2 bg-indigo-600 hover:bg-indigo-500',
            sizeClasses[size],
            className
          )}
        >
          <Sparkles className={iconSizes[size]} />
          Verify
        </Button>
      </div>
    );
  }

  // Completed state - Run Again
  if (isCompleted) {
    return (
      <Button
        onClick={onRun}
        size={size}
        variant={variant}
        className={cn(
          'gap-2',
          variant === 'default' && 'bg-emerald-600 hover:bg-emerald-500',
          sizeClasses[size],
          className
        )}
      >
        <Check className={iconSizes[size]} />
        Run Again
      </Button>
    );
  }

  // Failed state - Retry
  if (isFailed) {
    return (
      <Button
        onClick={onRun}
        size={size}
        variant={variant}
        className={cn(
          'gap-2',
          variant === 'default' && 'bg-red-600 hover:bg-red-500',
          sizeClasses[size],
          className
        )}
      >
        <X className={iconSizes[size]} />
        Retry
      </Button>
    );
  }

  // Default state - Run
  return (
    <Button
      onClick={onRun}
      disabled={disabled}
      size={size}
      variant={variant}
      className={cn(
        'gap-2',
        variant === 'default' && 'bg-indigo-600 hover:bg-indigo-500',
        sizeClasses[size],
        className
      )}
    >
      <Play className={iconSizes[size]} />
      Run
    </Button>
  );
}

/**
 * Compact version for inline use
 */
export function RunCapsuleButtonCompact({
  status,
  onRun,
  disabled = false,
}: {
  status: ExecutionStatus;
  onRun: () => void;
  disabled?: boolean;
}) {
  const isRunning = status === 'running' || status === 'queued' || status === 'pending';

  if (isRunning) {
    return (
      <div className="flex h-8 w-8 items-center justify-center rounded-md bg-blue-500/10">
        <Loader2 className="h-4 w-4 animate-spin text-blue-400" />
      </div>
    );
  }

  return (
    <button
      onClick={onRun}
      disabled={disabled}
      className="flex h-8 w-8 items-center justify-center rounded-md bg-indigo-500/10 text-indigo-400 transition-colors hover:bg-indigo-500/20 disabled:opacity-50"
      aria-label="Run code"
    >
      <Play className="h-4 w-4" />
    </button>
  );
}
