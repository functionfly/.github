import * as React from 'react';
import { cn } from '@/lib/utils';
import { StatTile } from './stat-tile';

interface StatStripProps {
  children: React.ReactNode;
  className?: string;
}

/**
 * Hairline-divided row of stat tiles. The container itself is just a
 * flex/grid frame — divider lines come from the children, so the strip
 * reads as one continuous piece of instrumentation.
 */
export function StatStrip({ children, className }: StatStripProps) {
  return (
    <div
      className={cn(
        'grid grid-cols-2 sm:grid-cols-4 divide-x divide-white/[0.06] rounded-lg border border-white/[0.06] bg-white/[0.015]',
        className
      )}
    >
      {children}
    </div>
  );
}

export { StatTile };
