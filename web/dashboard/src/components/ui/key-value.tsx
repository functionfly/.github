import * as React from 'react';
import { cn } from '@/lib/utils';

interface KeyValueProps {
  label: string;
  value: React.ReactNode;
  /** Optional icon shown before the label */
  icon?: React.ReactNode;
  /** Render the value as monospace (default true) */
  mono?: boolean;
  className?: string;
}

/**
 * Compact, scanline-friendly key/value pair. Used inside credential cards
 * for "Issued · 2024-03-14", "Expires · 2027-03-14", etc.
 */
export function KeyValue({
  label,
  value,
  icon,
  mono = true,
  className,
}: KeyValueProps) {
  return (
    <div className={cn('flex flex-col gap-0.5', className)}>
      <span className="flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-[0.14em] text-text-muted">
        {icon}
        {label}
      </span>
      <span
        className={cn(
          'text-sm text-text-primary',
          mono && 'font-mono tabular-nums'
        )}
      >
        {value}
      </span>
    </div>
  );
}
