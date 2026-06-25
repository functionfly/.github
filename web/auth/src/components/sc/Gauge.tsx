import React from "react";

interface GaugeProps {
  value: string;
  label: string;
  unit?: string;
  size?: "sm" | "md" | "lg";
}

export const Gauge: React.FC<GaugeProps> = ({
  value,
  label,
  unit,
  size = "md",
}) => {
  const SIZES = {
    sm: { valueSize: "1.125rem", labelSize: "9px", gap: "0.125rem" },
    md: { valueSize: "1.25rem", labelSize: "10px", gap: "0.25rem" },
    lg: { valueSize: "1.5rem", labelSize: "12px", gap: "0.375rem" },
  };

  const s = SIZES[size];

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: s.gap,
      }}
    >
      <span
        style={{
          fontFamily: "var(--font-mono)",
          fontWeight: 700,
          color: "var(--status-ok)",
          letterSpacing: "-0.02em",
          lineHeight: 1,
          fontSize: s.valueSize,
        }}
      >
        {value}
        {unit && (
          <span
            style={{
              color: "var(--text-dim)",
              fontSize: "0.7em",
              fontWeight: 400,
              marginLeft: "0.125rem",
            }}
          >
            {unit}
          </span>
        )}
      </span>
      <span
        style={{
          fontFamily: "var(--font-mono)",
          textTransform: "uppercase",
          letterSpacing: "0.14em",
          color: "var(--text-faint)",
          fontSize: s.labelSize,
        }}
      >
        {label}
      </span>
    </div>
  );
};
