import type { ReactNode } from 'react';

interface GaugeProps {
  children: ReactNode;
  first?: boolean;
}

export function Gauge({ children, first = false }: GaugeProps) {
  return (
    <div className={`gauge ${first ? '' : ''}`}>
      {children}
    </div>
  );
}

interface GaugeValueProps {
  children: ReactNode;
}

export function GaugeValue({ children }: GaugeValueProps) {
  return (
    <div className="gauge__value">
      <span className="gauge__tick" />
      {children}
    </div>
  );
}

interface GaugeLabelProps {
  children: ReactNode;
}

export function GaugeLabel({ children }: GaugeLabelProps) {
  return (
    <div className="gauge__label">
      {children}
    </div>
  );
}
