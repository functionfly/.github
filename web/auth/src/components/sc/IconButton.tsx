import React, { forwardRef } from "react";

interface IconButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  active?: boolean;
  label: string;
}

export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ active = false, label, className = "", children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={`icon-button ${active ? "icon-button-active" : ""} ${className}`}
        aria-label={label}
        title={label}
        style={{
          width: "36px",
          height: "36px",
          display: "inline-flex",
          alignItems: "center",
          justifyContent: "center",
          background: "transparent",
          border: "1px solid var(--steel)",
          borderRadius: "var(--radius)",
          color: active ? "var(--status-ok)" : "var(--text-dim)",
          cursor: "pointer",
          transition:
            "border-color var(--duration-fast) var(--ease-out), color var(--duration-fast) var(--ease-out)",
        }}
        {...props}
      >
        {children}
      </button>
    );
  },
);

IconButton.displayName = "IconButton";
