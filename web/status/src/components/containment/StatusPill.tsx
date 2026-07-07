import React from 'react';

type Status = 'live' | 'ok' | 'pending' | 'revoked';

interface StatusPillProps {
  status: Status;
  label: string;
  className?: string;
}

export const StatusPill: React.FC<StatusPillProps> = ({ status, label, className = '' }) => {
  const classes = ['status-pill', `status-pill--${status}`, className].filter(Boolean).join(' ');

  return (
    <span className={classes}>
      <span className="status-pill__dot" />
      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{label}</span>
    </span>
  );
};
