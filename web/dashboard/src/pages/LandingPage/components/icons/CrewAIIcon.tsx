export function CrewAIIcon({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg" className={className}>
      <rect width="40" height="40" rx="8" fill="#2A1A2E" />
      <circle cx="20" cy="16" r="5" stroke="#C084FC" strokeWidth="2" fill="none" />
      <circle cx="12" cy="27" r="3.5" stroke="#C084FC" strokeWidth="1.8" fill="none" />
      <circle cx="28" cy="27" r="3.5" stroke="#C084FC" strokeWidth="1.8" fill="none" />
      <line x1="16" y1="20" x2="13" y2="24" stroke="#C084FC" strokeWidth="1.5" />
      <line x1="24" y1="20" x2="27" y2="24" stroke="#C084FC" strokeWidth="1.5" />
      <line x1="15.5" y1="27" x2="24.5" y2="27" stroke="#C084FC" strokeWidth="1.5" strokeDasharray="2 2" />
    </svg>
  );
}
