import React from 'react';

interface GaugeProps {
  value: string;
  label: string;
  unit?: string;
  size?: 'sm' | 'md' | 'lg';
}

const SIZE_CLASSES = {
  sm: {
    value: 'text-lg',
    label: 'text-[9px]',
    gap: 'gap-0.5',
    dotSize: 'w-1 h-1',
  },
  md: {
    value: 'text-xl',
    label: 'text-[10px]',
    gap: 'gap-1',
    dotSize: 'w-1.5 h-1.5',
  },
  lg: {
    value: 'text-2xl',
    label: 'text-xs',
    gap: 'gap-1.5',
    dotSize: 'w-1.5 h-1.5',
  },
};

export const Gauge: React.FC<GaugeProps> = ({
  value,
  label,
  unit,
  size = 'md',
}) => {
  const sizeClass = SIZE_CLASSES[size];

  return (
    <div
      className={`flex flex-col items-center justify-center ${sizeClass.gap}`}
      style={{ padding: 'var(--space-5) 0 var(--space-6)' }}
    >
      <div className="flex items-baseline gap-2">
        <span
          className={`inline-block ${sizeClass.dotSize} rounded-full`}
          style={{
            backgroundColor: 'var(--status-ok)',
            boxShadow: '0 0 6px rgba(143,255,208,0.6)',
            marginTop: '2px',
          }}
        />
        <span
          className={`
            font-[var(--font-mono)] font-medium text-[var(--status-ok)]
            tracking-tight leading-none
            ${sizeClass.value}
          `}
        >
          {value}
          {unit && (
            <span className="text-[var(--text-dim)] text-[0.7em] font-normal ml-0.5">
              {unit}
            </span>
          )}
        </span>
      </div>
      <span
        className={`
          font-[var(--font-mono)] uppercase tracking-widest
          text-[var(--text-faint)]
          ${sizeClass.label}
        `}
        style={{ marginTop: 'var(--space-1)' }}
      >
        {label}
      </span>
    </div>
  );
};
