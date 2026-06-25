import { Shield } from 'lucide-react';

const clearanceConfig: Record<number, { label: string; color: string; bg: string }> = {
  0: { label: 'Public', color: 'text-gray-400', bg: 'bg-gray-500/20' },
  1: { label: 'Internal', color: 'text-blue-400', bg: 'bg-blue-500/20' },
  2: { label: 'Confidential', color: 'text-yellow-400', bg: 'bg-yellow-500/20' },
  3: { label: 'Restricted', color: 'text-orange-400', bg: 'bg-orange-500/20' },
  4: { label: 'Critical', color: 'text-red-400', bg: 'bg-red-500/20' },
  5: { label: 'Maximum', color: 'text-purple-400', bg: 'bg-purple-500/20' },
};

interface ClearanceBadgeProps {
  level: number;
  className?: string;
}

export function ClearanceBadge({ level, className = '' }: ClearanceBadgeProps) {
  const config = clearanceConfig[level] || clearanceConfig[0];

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold ${config.bg} ${config.color} ${className}`}
    >
      <Shield className="h-3 w-3" />
      L{level} &middot; {config.label}
    </span>
  );
}
