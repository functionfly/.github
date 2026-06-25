import type { ReactNode } from 'react';

interface GaugeStripProps {
  children: ReactNode;
}

export function GaugeStrip({ children }: GaugeStripProps) {
  return (
    <div className="gauge-strip">
      {children}
    </div>
  );
}
