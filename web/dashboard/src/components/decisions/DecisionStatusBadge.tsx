import type { DecisionStatus } from '@/api/decisions';
import { cn } from '@/lib/utils';

interface DecisionStatusBadgeProps {
  status: DecisionStatus;
  className?: string;
}

const statusConfig = {
  pending: {
    label: 'Pending',
    className: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
  },
  approved: {
    label: 'Approved',
    className: 'bg-green-500/10 text-green-500 border-green-500/20',
  },
  superseded: {
    label: 'Superseded',
    className: 'bg-orange-500/10 text-orange-500 border-orange-500/20',
  },
  deprecated: {
    label: 'Deprecated',
    className: 'bg-gray-500/10 text-gray-500 border-gray-500/20',
  },
};

export function DecisionStatusBadge({ status, className }: DecisionStatusBadgeProps) {
  const config = statusConfig[status] || statusConfig.pending;

  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium',
        config.className,
        className
      )}
    >
      {config.label}
    </span>
  );
}
