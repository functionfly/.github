import React from "react";

/**
 * PageGrid - Blueprint paper texture backdrop
 * Renders once per page as the base layer, fixed behind all content
 */
export const PageGrid: React.FC<{ className?: string }> = ({
  className = "",
}) => {
  return (
    <div
      className={`page-grid ${className}`}
      style={{
        position: "fixed",
        inset: 0,
        backgroundImage:
          "linear-gradient(rgba(255,255,255,0.025) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.025) 1px, transparent 1px)",
        backgroundSize: "48px 48px",
        pointerEvents: "none",
        zIndex: "var(--z-base)",
      }}
      aria-hidden="true"
    />
  );
};
