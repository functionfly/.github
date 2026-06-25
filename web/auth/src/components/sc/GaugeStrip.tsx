import React from "react";

interface GaugeItem {
  value: string;
  label: string;
  tooltip?: string;
}

interface GaugeStripProps {
  items: GaugeItem[];
  className?: string;
}

export const GaugeStrip: React.FC<GaugeStripProps> = ({
  items,
  className = "",
}) => {
  return (
    <div
      className={`gauge-strip-grid ${className}`}
      style={{
        display: "grid",
        gridTemplateColumns: `repeat(${items.length}, 1fr)`,
        borderTop: "1px solid var(--panel-edge)",
      }}
    >
      {items.map((item, i) => (
        <React.Fragment key={item.label}>
          <div
            style={{
              padding: "var(--space-5) 0 var(--space-6)",
              paddingLeft: i === 0 ? 0 : "var(--space-4)",
              borderLeft: i > 0 ? "1px solid var(--panel-edge)" : undefined,
            }}
          >
            <span
              className="gauge-strip-value"
              style={{
                display: "flex",
                alignItems: "baseline",
                gap: "var(--space-2)",
                fontFamily: "var(--font-mono)",
                fontSize: "20px",
                fontWeight: 500,
                color: "var(--status-ok)",
                letterSpacing: "-0.02em",
              }}
            >
              <span
                style={{
                  width: "5px",
                  height: "5px",
                  borderRadius: "50%",
                  backgroundColor: "var(--status-ok)",
                  boxShadow: "0 0 6px rgba(143,255,208,0.6)",
                  flexShrink: 0,
                }}
              />
              {item.value}
            </span>
            <span
              className="gauge-strip-label"
              style={{
                display: "block",
                fontFamily: "var(--font-mono)",
                fontSize: "10px",
                textTransform: "uppercase",
                letterSpacing: "0.1em",
                color: "var(--text-faint)",
                marginTop: "var(--space-1)",
              }}
            >
              {item.label}
            </span>
            {item.tooltip && (
              <span
                style={{
                  display: "block",
                  fontFamily: "var(--font-mono)",
                  fontSize: "8px",
                  color: "var(--text-faint)",
                  opacity: 0.7,
                  marginTop: "2px",
                }}
                title={item.tooltip}
              >
                *{item.tooltip}
              </span>
            )}
          </div>
        </React.Fragment>
      ))}
    </div>
  );
};
