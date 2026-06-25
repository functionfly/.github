interface StatusPillProps {
  status: 'live' | 'pending' | 'revoked';
  label: string;
}

export function StatusPill({ status, label }: StatusPillProps) {
  return (
    <span className={`status-pill status-pill--${status}`}>
      <span className="status-pill__dot" />
      <span className="sr-only">{status}: </span>
      {label}
    </span>
  );
}
