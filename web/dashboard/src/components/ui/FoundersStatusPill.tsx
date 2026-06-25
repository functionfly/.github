import React from 'react';
import { cn } from '@/lib/utils';
import { Check } from 'lucide-react';

type StatusType = 'live' | 'pending' | 'revoked' | 'claimed' | 'available';

interface StatusPillProps {
  status: StatusType;
  label?: string;
  showDot?: boolean;
  pulse?: boolean;
  className?: string;
}

const statusConfig: Record<
  StatusType,
  { className: string; defaultLabel: string; dotClass: string }
> = {
  live: {
    className: 'status-pill--live',
    defaultLabel: 'Live',
    dotClass: 'status-pill__dot--pulse',
  },
  pending: {
    className: 'status-pill--pending',
    defaultLabel: 'Pending',
    dotClass: '',
  },
  revoked: {
    className: 'status-pill--revoked',
    defaultLabel: 'Revoked',
    dotClass: '',
  },
  claimed: {
    className: 'status-pill--live',
    defaultLabel: 'Claimed',
    dotClass: '',
  },
  available: {
    className: 'status-pill--pending',
    defaultLabel: 'Available',
    dotClass: '',
  },
};

export function StatusPill({
  status,
  label,
  showDot = true,
  pulse = false,
  className = '',
}: StatusPillProps) {
  const config = statusConfig[status];

  return (
    <span className={cn('status-pill', config.className, className)}>
      {showDot && (
        <span
          className={cn(
            'status-pill__dot',
            pulse && config.dotClass
          )}
        />
      )}
      {label ?? config.defaultLabel}
    </span>
  );
}

interface LiveStatusProps {
  label?: string;
  className?: string;
}

export function LiveStatus({ label = 'Live', className = '' }: LiveStatusProps) {
  return (
    <span className={cn('status-pill status-pill--live', className)}>
      <span className="status-pill__dot status-pill__dot--pulse" />
      {label}
    </span>
  );
}

interface ClaimedBadgeProps {
  className?: string;
}

export function ClaimedBadge({ className = '' }: ClaimedBadgeProps) {
  return (
    <span className={cn('status-pill status-pill--live flex items-center gap-1', className)}>
      <Check size={10} strokeWidth={3} />
      Claimed
    </span>
  );
}

export function AvailableBadge({ className = '' }: ClaimedBadgeProps) {
  return (
    <span className={cn('status-pill status-pill--pending', className)}>
      Available
    </span>
  );
}
