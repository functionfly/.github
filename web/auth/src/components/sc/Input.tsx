import React, { forwardRef, type ReactNode } from "react";

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: string;
  label?: string;
  children?: ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ error, label, className = "", id, children, ...props }, ref) => {
    const inputId = id || `input-${Math.random().toString(36).slice(2, 9)}`;
    const hasError = Boolean(error);

    return (
      <div
        className={`input-wrapper ${className}`}
        style={{
          display: "flex",
          flexDirection: "column",
          gap: "var(--space-2)",
        }}
      >
        {label && (
          <label
            htmlFor={inputId}
            style={{
              fontFamily: "var(--font-body)",
              fontSize: "13px",
              fontWeight: 500,
              color: "var(--text-dim)",
            }}
          >
            {label}
          </label>
        )}
        <div
          style={{
            position: "relative",
            display: "flex",
            alignItems: "center",
          }}
        >
          <input
            ref={ref}
            id={inputId}
            className="sealed-input"
            aria-invalid={hasError}
            aria-describedby={hasError ? `${inputId}-error` : undefined}
            style={{
              fontFamily: "var(--font-body)",
              fontSize: "15px",
              lineHeight: 1.6,
              color: "var(--text)",
              background: "var(--panel-raised)",
              border: `1px solid ${hasError ? "var(--status-revoked)" : "var(--steel)"}`,
              borderRadius: "var(--radius)",
              padding: "var(--space-3) var(--space-4)",
              paddingRight: children ? "var(--space-10)" : undefined,
              boxShadow: hasError
                ? "var(--shadow-input-error)"
                : "var(--shadow-input-rest)",
              outline: "none",
              transition:
                "border-color var(--duration-fast) var(--ease-out), box-shadow var(--duration-fast) var(--ease-out)",
              width: "100%",
              boxSizing: "border-box",
            }}
            {...props}
          />
          {children}
        </div>
        {hasError && (
          <span
            id={`${inputId}-error`}
            role="alert"
            style={{
              fontFamily: "var(--font-body)",
              fontSize: "13px",
              lineHeight: 1.5,
              color: "var(--status-revoked)",
            }}
          >
            {error}
          </span>
        )}
      </div>
    );
  },
);

Input.displayName = "Input";
