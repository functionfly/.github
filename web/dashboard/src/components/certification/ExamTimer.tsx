import { cn } from '@/lib/utils';
import { useExamTimer } from '@/hooks/useExamTimer';
import { Clock, AlertTriangle } from 'lucide-react';

interface ExamTimerProps {
  expiresAt: string;
  onExpire?: () => void;
}

export function ExamTimer({ expiresAt, onExpire }: ExamTimerProps) {
  const { formatted, isWarning, isCritical, progress, isExpired } = useExamTimer({
    expiresAt,
    onExpire,
  });

  if (isExpired) {
    return (
      <div className="flex items-center gap-2 rounded-lg bg-red-500/10 border border-red-500/20 px-4 py-2">
        <AlertTriangle className="h-4 w-4 text-red-500" />
        <span className="text-sm font-medium text-red-500">Time Expired</span>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {/* Timer display */}
      <div
        className={cn(
          'flex items-center gap-2 rounded-lg border px-4 py-2 transition-colors',
          isCritical
            ? 'border-red-500/30 bg-red-500/10'
            : isWarning
              ? 'border-amber-500/30 bg-amber-500/10'
              : 'border-theme bg-card'
        )}
      >
        <Clock
          className={cn(
            'h-4 w-4',
            isCritical ? 'text-red-500 animate-pulse' : isWarning ? 'text-amber-500' : 'text-text-muted'
          )}
        />
        <span
          className={cn(
            'font-mono text-lg font-bold',
            isCritical ? 'text-red-500' : isWarning ? 'text-amber-500' : 'text-text-primary'
          )}
        >
          {formatted}
        </span>
        {isCritical && (
          <span className="text-xs text-red-400 ml-auto">Hurry!</span>
        )}
      </div>

      {/* Progress bar */}
      <div className="h-1 w-full rounded-full bg-bg-tertiary overflow-hidden">
        <div
          className={cn(
            'h-full rounded-full transition-all duration-1000',
            isCritical
              ? 'bg-red-500'
              : isWarning
                ? 'bg-amber-500'
                : 'bg-brand-500'
          )}
          style={{ width: `${100 - progress}%` }}
        />
      </div>
    </div>
  );
}
