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
      className="grid"
      style={{
        gridTemplateColumns: `repeat(${items.length}, 1fr)`,
        borderTop: '1px solid var(--panel-edge)',
      }}
    >
      {items.map((item, i) => (
        <div
          key={item.label}
          className="flex flex-col items-center justify-center"
          style={{
            padding: 'var(--space-5) var(--space-4) var(--space-6)',
            borderLeft: i > 0 ? '1px solid var(--panel-edge)' : undefined,
          }}
        >
          <span
            className="flex items-baseline"
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '26px',
              fontWeight: 500,
              color: 'var(--status-ok)',
              gap: 'var(--space-3)',
              paddingLeft: 'var(--space-1)',
              letterSpacing: '-0.02em',
              lineHeight: 1.2,
            }}
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
            className="uppercase"
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '10px',
              fontWeight: 500,
              letterSpacing: '0.06em',
              color: 'var(--text-faint)',
              marginTop: 'var(--space-1)',
            }}
          >
            {item.label}
          </span>
        </div>
      ))}
      <style>{`
        @media (max-width: 479px) {
          .grid {
            grid-template-columns: 1fr !important;
            border-top: none;
          }
          .grid > div {
            border-left: none !important;
            border-top: 1px solid var(--panel-edge);
            padding-top: var(--space-5);
          }
          .grid > div:first-child {
            border-top: none;
            padding-top: 0;
          }
        }
      `}</style>
    </div>
  );
};
