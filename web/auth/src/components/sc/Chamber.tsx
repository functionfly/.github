import React from "react";

interface ChamberProps {
  children: React.ReactNode;
  ribs?: boolean;
  nested?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

export const Chamber: React.FC<ChamberProps> = ({
  children,
  ribs = false,
  nested = false,
  className = "",
  style,
}) => {
  const bgColor = nested ? "var(--panel-raised)" : "var(--panel)";
  const radialGradient = nested
    ? undefined
    : "radial-gradient(140% 100% at 15% 0%, var(--glass-tint), transparent 55%)";

  return (
    <div
      className={className}
      style={{
        position: "relative",
        borderRadius: "var(--radius-lg)",
        background: radialGradient
          ? `${radialGradient}, ${bgColor}`
          : bgColor,
        boxShadow: nested ? undefined : "var(--shadow-chamber)",
        border: "1px solid var(--panel-edge)",
        padding: "var(--space-7)",
        ...style,
      }}
    >
      {ribs && (
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            pointerEvents: "none",
            userSelect: "none",
            opacity: 0.025,
            backgroundImage:
              "repeating-linear-gradient(90deg, transparent 0px, transparent 119px, rgba(255,255,255,0.025) 120px)",
          }}
          aria-hidden="true"
        />
      )}
      {children}
    </div>
  );
};
