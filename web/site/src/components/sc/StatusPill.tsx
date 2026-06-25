import React from 'react';

type Status = 'live' | 'ok' | 'pending' | 'revoked';

interface StatusPillProps {
  status: Status;
  label: string;
}

const STATUS_STYLES: Record<Status, { text: string; border: string; bg: string }> = {
  live: {
    text: 'var(--status-ok)',
    border: 'rgba(143,255,208,0.3)',
    bg: 'rgba(143,255,208,0.06)',
  },
  ok: {
    text: 'var(--status-ok)',
    border: 'rgba(143,255,208,0.3)',
    bg: 'rgba(143,255,208,0.06)',
  },
  pending: {
    text: 'var(--status-pending)',
    border: 'rgba(232,196,104,0.3)',
    bg: 'rgba(232,196,104,0.06)',
  },
  revoked: {
    text: 'var(--status-revoked)',
    border: 'rgba(255,107,107,0.3)',
    bg: 'rgba(255,107,107,0.06)',
  },
};

export const StatusPill: React.FC<StatusPillProps> = ({ status, label }) => {
  const styles = STATUS_STYLES[status];

  return (
    <span
      className="inline-flex items-center gap-1 font-[var(--font-mono)] text-[11px] font-medium uppercase tracking-widest"
      style={{
        color: styles.text,
        borderColor: styles.border,
        backgroundColor: styles.bg,
        padding: '3px 12px',
        borderRadius: 'var(--radius-sm)',
        borderWidth: '1px',
        borderStyle: 'solid',
      }}
    >
      <span
        className="inline-block w-1.5 h-1.5 rounded-full"
        style={{ backgroundColor: styles.text }}
      />
      {label}
    </span>
  );
};
