import { ExternalLink } from 'lucide-react';

interface StatusBadgeProps {
  className?: string;
}

export function StatusBadge({ className = '' }: StatusBadgeProps) {
  return (
    <a
      href="https://status.functionfly.com"
      target="_blank"
      rel="noopener noreferrer"
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium bg-green-500/10 text-green-400 hover:bg-green-500/20 transition-colors ${className}`}
    >
      <span className="w-1.5 h-1.5 rounded-full bg-green-400" />
      All Systems Operational
      <ExternalLink className="w-3 h-3 opacity-60" />
    </a>
  );
}

interface StatusIndicatorProps {
  className?: string;
}

export function StatusIndicator({ className = '' }: StatusIndicatorProps) {
  return (
    <a
      href="https://status.functionfly.com"
      target="_blank"
      rel="noopener noreferrer"
      className={`inline-flex items-center gap-2 ${className}`}
    >
      <span className="w-2 h-2 rounded-full bg-green-400" />
      <span className="text-sm text-text-secondary">Status</span>
    </a>
  );
}