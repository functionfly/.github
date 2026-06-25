import React from 'react';

interface GaugeProps {
  value: string | number;
  label: string;
  unit?: string;
  className?: string;
}

const VALUE_CLASSES = 'font-mono text-sm font-semibold text-[var(--text-primary)] tracking-tight';
const LABEL_CLASSES = 'font-mono text-[10px] font-medium text-[var(--text-faint)] uppercase tracking-widest';

/**
 * Gauge - Individual stat readout with value and label.
 * Used within GaugeStrip or standalone for metrics display.
 */
export const Gauge: React.FC<GaugeProps> = ({
  value,
  label,
  unit,
  className = '',
}) => {
  return (
    <div className={`flex flex-col gap-0.5 min-w-[4rem] ${className}`}>
      <div className="flex items-baseline gap-0.5">
        <span className={VALUE_CLASSES}>{value}</span>
        {unit && (
          <span className="font-mono text-xs text-[var(--text-secondary)]">{unit}</span>
        )}
      </div>
      <span className={LABEL_CLASSES}>{label}</span>
    </div>
  );
};

Gauge.displayName = 'Gauge';

interface GaugeStripProps {
  gauges: Array<{
    value: string | number;
    label: string;
    unit?: string;
  }>;
  className?: string;
}

const VERTICAL_RULE = (
  <div
    className="w-px h-8 bg-[var(--steel)] self-center"
    aria-hidden="true"
  />
);

/**
 * GaugeStrip - Horizontal strip of stat readouts.
 * Divides by thin vertical rules, not separate cards.
 */
export const GaugeStrip: React.FC<GaugeStripProps> = ({
  gauges,
  className = '',
}) => {
  if (gauges.length === 0) return null;

  return (
    <div
      className={`flex items-start gap-0 ${className}`}
      role="list"
      aria-label="Statistics"
    >
      {gauges.map((gauge, index) => (
        <React.Fragment key={gauge.label}>
          {index > 0 && VERTICAL_RULE}
          <Gauge
            value={gauge.value}
            label={gauge.label}
            unit={gauge.unit}
          />
        </React.Fragment>
      ))}
    </div>
  );
};

GaugeStrip.displayName = 'GaugeStrip';
