import React from "react";
import type { ChamberProps } from "./index";

interface CardProps extends Omit<ChamberProps, "nested"> {
  /** Show trust seal (16px variant) next to title */
  trustSeal?: boolean;
  sealLabel?: string;
}

export const Card: React.FC<CardProps> = ({
  children,
  trustSeal = false,
  sealLabel,
  className = "",
  style,
  ...chamberProps
}) => {
  return (
    <div
      className={`card ${className}`}
      style={{
        background: "var(--panel-raised)",
        borderRadius: "var(--radius)",
        border: "1px solid var(--panel-edge)",
        padding: "var(--space-5)",
        ...style,
      }}
    >
      {trustSeal && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "var(--space-2)",
          }}
        >
          {/* TrustSeal 16px variant would be imported here */}
          <div
            style={{
              width: "16px",
              height: "16px",
              borderRadius: "50%",
              background:
                "conic-gradient(from 140deg, var(--foil-a), var(--foil-b), var(--foil-c), var(--foil-d), var(--foil-a))",
              boxShadow: "var(--shadow-seal)",
              flexShrink: 0,
            }}
            aria-hidden="true"
          />
          {sealLabel && (
            <span
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: "11px",
                fontWeight: 500,
                textTransform: "uppercase",
                letterSpacing: "0.06em",
                color: "var(--text-dim)",
              }}
            >
              {sealLabel}
            </span>
          )}
        </div>
      )}
      {children}
    </div>
  );
};
