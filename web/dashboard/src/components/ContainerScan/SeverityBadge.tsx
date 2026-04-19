import { severityConfig, type SeverityLevel } from './types';
import { cn } from '@/lib/utils';

interface SeverityBadgeProps {
  severity: SeverityLevel;
  showIcon?: boolean;
  showLabel?: boolean;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  count?: number;
}

export function SeverityBadge({
  severity,
  showIcon = true,
  showLabel = true,
  size = 'md',
  className,
  count,
}: SeverityBadgeProps) {
  const config = severityConfig[severity];
  const Icon = config.icon;

  const sizeClasses = {
    sm: 'text-xs px-2 py-0.5 gap-1',
    md: 'text-sm px-2.5 py-1 gap-1.5',
    lg: 'text-base px-3 py-1.5 gap-2',
  };

  const iconSizes = {
    sm: 'w-3 h-3',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center font-medium rounded-full border',
        config.color,
        config.bgColor,
        config.borderColor,
        sizeClasses[size],
        className
      )}
    >
      {showIcon && <Icon className={iconSizes[size]} />}
      {showLabel && <span>{config.label}</span>}
      {count !== undefined && count > 0 && (
        <span className={cn('font-semibold', size === 'sm' ? 'text-xs' : 'text-sm')}>
          {count}
        </span>
      )}
    </span>
  );
}

interface SeverityBarProps {
  counts: Record<SeverityLevel, number>;
  total: number;
  className?: string;
}

export function SeverityBar({ counts, total, className }: SeverityBarProps) {
  const severities: SeverityLevel[] = ['critical', 'high', 'medium', 'low', 'info'];
  
  if (total === 0) {
    return (
      <div className={cn('h-2 w-full bg-gray-700 rounded-full', className)}>
        <div className="h-full w-full bg-green-500/30 rounded-full" />
      </div>
    );
  }

  return (
    <div className={cn('h-2 w-full flex rounded-full overflow-hidden', className)}>
      {severities.map((sev) => {
        const count = counts[sev];
        if (count === 0) return null;
        const percentage = (count / total) * 100;
        const config = severityConfig[sev];
        
        return (
          <div
            key={sev}
            className={cn('h-full transition-all duration-300', config.color.replace('text-', 'bg-'))}
            style={{ width: `${percentage}%` }}
            title={`${config.label}: ${count}`}
          />
        );
      })}
    </div>
  );
}

interface SeverityCountProps {
  counts: Record<SeverityLevel, number>;
  className?: string;
}

export function SeverityCounts({ counts, className }: SeverityCountProps) {
  const severities: SeverityLevel[] = ['critical', 'high', 'medium', 'low', 'info'];

  return (
    <div className={cn('flex flex-wrap gap-2', className)}>
      {severities.map((sev) => {
        const count = counts[sev];
        if (count === 0) return null;
        
        return (
          <SeverityBadge
            key={sev}
            severity={sev}
            count={count}
            size="sm"
          />
        );
      })}
    </div>
  );
}
