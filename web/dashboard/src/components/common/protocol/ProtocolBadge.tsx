/**
 * ProtocolBadge — renders a visual badge for MCP or A2A protocol.
 * Used everywhere a call, receipt, or task appears in the UI.
 */

import { Zap, Users } from 'lucide-react';

interface ProtocolBadgeProps {
  protocol: 'mcp' | 'a2a' | string;
  className?: string;
  size?: 'sm' | 'md';
}

export function ProtocolBadge({ protocol, className = '', size = 'sm' }: ProtocolBadgeProps) {
  const isMCP = protocol === 'mcp';
  const Icon = isMCP ? Zap : Users;
  const label = isMCP ? 'MCP' : 'A2A';

  const sizeClasses = size === 'sm'
    ? 'px-2 py-0.5 text-xs gap-1'
    : 'px-3 py-1 text-sm gap-1.5';

  return (
    <span
      className={`inline-flex items-center font-medium rounded-full ${
        isMCP
          ? 'bg-brand-500/10 text-brand-400 border border-brand-500/20'
          : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
      } ${sizeClasses} ${className}`}
    >
      <Icon className={size === 'sm' ? 'w-3 h-3' : 'w-4 h-4'} />
      {label}
    </span>
  );
}
