export type Status = 'live' | 'pending' | 'revoked';

export interface StatusPillProps {
  status: Status;
  label?: string;
  className?: string;
}

const STATUS_LABELS: Record<Status, string> = {
  live: 'Live',
  pending: 'Pending',
  revoked: 'Revoked',
};

export function StatusPill({ status, label, className = '' }: StatusPillProps) {
  const statusLabel = label || STATUS_LABELS[status];

  return (
    <span className={`status-pill status-pill--${status} ${className}`}>
      <span className="status-pill__dot" aria-hidden="true" />
      <span className="status-pill__label">{statusLabel}</span>
    </span>
  );
}