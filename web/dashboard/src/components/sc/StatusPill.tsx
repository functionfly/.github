import React from 'react';

type Status = 'live' | 'pending' | 'revoked';

interface StatusPillProps {
  status: Status;
  label?: string;
  className?: string;
}

const STATUS_STYLES: Record<Status, { dot: string; text: string }> = {
  live: {
    dot: 'bg-[var(--status-ok)]',
    text: 'text-[var(--status-ok)]',
  },
  pending: {
    dot: 'bg-[var(--status-pending)]',
    text: 'text-[var(--text-faint)]',
  },
  revoked: {
    dot: 'bg-[var(--status-revoked)]',
    text: 'text-[var(--status-revoked)]',
  },
};

const DEFAULT_LABELS: Record<Status, string> = {
  live: 'Live',
  pending: 'Pending',
  revoked: 'Revoked',
};

/**
 * StatusPill - Monospace status indicator with leading dot.
 * Used for Live, sandbox states, and agent run states.
 */
export const StatusPill: React.FC<StatusPillProps> = ({
  status,
  label,
  className = '',
}) => {
  const styles = STATUS_STYLES[status];
  const displayLabel = label ?? DEFAULT_LABELS[status];

  return (
    <span
      className={`inline-flex items-center gap-1.5 font-mono text-xs font-semibold uppercase tracking-wider ${styles.text} ${className}`}
    >
      <span
        className={`w-1.5 h-1.5 rounded-full ${styles.dot}`}
        aria-hidden="true"
      />
      <span>{displayLabel}</span>
    </span>
  );
};

StatusPill.displayName = 'StatusPill';
