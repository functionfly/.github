import React from 'react';

interface GaugeProps {
  value: string;
  label: string;
  unit?: string;
  size?: 'sm' | 'md' | 'lg';
}

const SIZES = {
  sm: { value: '18px', label: '9px', dotSize: 4 },
  md: { value: '26px', label: '10px', dotSize: 5 },
  lg: { value: '32px', label: '12px', dotSize: 6 },
};

export const Gauge: React.FC<GaugeProps> = ({
  value,
  label,
  unit,
  size = 'md',
}) => {
  const s = SIZES[size];

  return (
    <div className="flex flex-col items-center justify-center" style={{ padding: 'var(--space-5) 0 var(--space-6)' }}>
      <div className="flex items-baseline gap-2">
        <span
          className="inline-block rounded-full"
          style={{
            width: s.dotSize,
            height: s.dotSize,
            backgroundColor: 'var(--status-ok)',
            boxShadow: '0 0 6px rgba(143,255,208,0.6)',
            marginTop: '2px',
          }}
        />
        <span
          style={{
            fontFamily: 'var(--font-mono)',
            fontSize: s.value,
            fontWeight: 500,
            color: 'var(--status-ok)',
            letterSpacing: '-0.02em',
            lineHeight: 1.2,
          }}
        >
          {value}
          {unit && (
            <span style={{ color: 'var(--text-dim)', fontSize: '0.7em', fontWeight: 400, marginLeft: '2px' }}>
              {unit}
            </span>
          )}
        </span>
      </div>
      <span
        className="uppercase"
        style={{
          fontFamily: 'var(--font-mono)',
          fontSize: s.label,
          fontWeight: 500,
          letterSpacing: '0.06em',
          color: 'var(--text-faint)',
          marginTop: 'var(--space-1)',
        }}
      >
        {label}
      </span>
    </div>
  );
};
