export function AutoGenIcon({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg" className={className}>
      <rect width="40" height="40" rx="8" fill="#1A2A3A" />
      <circle cx="20" cy="20" r="9" stroke="#4A90E2" strokeWidth="2" fill="none" />
      <circle cx="20" cy="14" r="2.5" fill="#4A90E2" />
      <circle cx="26" cy="23" r="2.5" fill="#4A90E2" />
      <circle cx="14" cy="23" r="2.5" fill="#4A90E2" />
      <line x1="20" y1="16.5" x2="25" y2="21" stroke="#4A90E2" strokeWidth="1.5" />
      <line x1="20" y1="16.5" x2="15" y2="21" stroke="#4A90E2" strokeWidth="1.5" />
      <line x1="15" y1="23" x2="25" y2="23" stroke="#4A90E2" strokeWidth="1.5" />
    </svg>
  );
}
