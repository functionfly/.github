import React from "react";

interface GaugeItem {
  value: string;
  label: string;
}

interface GaugeStripProps {
  items: GaugeItem[];
}

export const GaugeStrip: React.FC<GaugeStripProps> = ({ items }) => {
  return (
    <div
      className="gauge-strip-grid"
      style={{
        display: 'grid',
        gridTemplateColumns: `repeat(${items.length}, 1fr)`,
        borderTop: '1px solid var(--panel-edge)',
      }}
    >
      {items.map((item, i) => (
        <React.Fragment key={item.label}>
          <div
            className="flex flex-col items-center justify-center"
            style={{
              padding: 'var(--space-5) var(--space-4) var(--space-6)',
              borderLeft: i > 0 ? '1px solid var(--panel-edge)' : undefined,
            }}
          >
              <span
                className="font-[var(--font-mono)] text-xl font-medium text-[var(--status-ok)] gauge-strip-value tracking-tight"
                style={{ display: 'flex', alignItems: 'baseline', gap: 'var(--space-3)', paddingLeft: 'var(--space-1)' }}
              >
              <span
                style={{
                  width: '5px',
                  height: '5px',
                  borderRadius: '50%',
                  backgroundColor: 'var(--status-ok)',
                  boxShadow: '0 0 6px rgba(143,255,208,0.6)',
                  flexShrink: 0,
                }}
              />
              {item.value}
            </span>
            <span
              className="font-[var(--font-mono)] text-[10px] uppercase tracking-widest text-[var(--text-faint)] gauge-strip-label"
              style={{ marginTop: 'var(--space-1)' }}
            >
              {item.label}
            </span>
          </div>
        </React.Fragment>
      ))}
    </div>
  );
};
