import * as React from 'react';
import { cn } from '@/lib/utils';

interface StatTileProps {
  label: string;
  value: React.ReactNode;
  /** Tiny annotation beneath the value, e.g. "2 within 30 days" */
  hint?: React.ReactNode;
  /** Trailing element (often a status pill) */
  trailing?: React.ReactNode;
  /** Show a colored vertical accent bar on the left edge */
  accent?: 'brand' | 'success' | 'warning' | 'danger' | 'none';
  className?: string;
}

const ACCENT_CLASS: Record<NonNullable<StatTileProps['accent']>, string> = {
  brand: 'before:bg-brand-500',
  success: 'before:bg-emerald-500',
  warning: 'before:bg-amber-500',
  danger: 'before:bg-red-500',
  none: 'before:bg-transparent',
};

/**
 * Compact stat tile used in the page header strip.
 * Vertical 1px accent on the left, monospace label, large display value,
 * subtle hint line — replaces the boring "0 active · 0 expiring · 23 exams"
 * subtitle text with a proper KPI row.
 */
export function StatTile({
  label,
  value,
  hint,
  trailing,
  accent = 'none',
  className,
}: StatTileProps) {
  return (
    <div
      className={cn(
        'relative flex flex-col gap-1 pl-3 pr-4 py-3',
        'before:absolute before:left-0 before:top-2 before:bottom-2 before:w-px',
        ACCENT_CLASS[accent],
        className
      )}
    >
      <div className="flex items-center justify-between gap-2">
        <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-text-muted">
          {label}
        </p>
        {trailing}
      </div>
      <p className="font-display text-2xl font-medium tabular-nums text-text-primary">
        {value}
      </p>
      {hint && <p className="text-xs text-text-secondary">{hint}</p>}
    </div>
  );
}
