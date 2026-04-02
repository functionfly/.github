import React, { type ButtonHTMLAttributes } from "react";

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "ghost";
  loading?: boolean;
  fullWidth?: boolean;
}

export default function Button({
  variant = "primary",
  loading = false,
  fullWidth = false,
  children,
  disabled,
  className = "",
  ...rest
}: Props) {
  return (
    <>
      <button
        disabled={disabled || loading}
        className={`btn btn-${variant}${loading ? " btn-loading" : ""}${fullWidth ? " btn-full" : ""} ${className}`}
        {...rest}
      >
        {loading ? (
          <span className="spinner" aria-hidden="true" />
        ) : null}
        {children}
      </button>
      <style>{`
        .btn {
          display: inline-flex; align-items: center; justify-content: center; gap: 0.5rem;
          padding: 0.625rem 1.25rem; border-radius: 8px; font-size: 0.9375rem; font-weight: 500;
          cursor: pointer; transition: background 0.15s, opacity 0.15s; border: 1px solid transparent;
          font-family: inherit; text-decoration: none;
        }
        .btn:disabled { opacity: 0.5; cursor: not-allowed; }
        .btn-primary { background: #6366f1; color: #fff; }
        .btn-primary:hover:not(:disabled) { background: #4f46e5; }
        .btn-secondary { background: #27272a; color: #fafafa; border-color: #3f3f46; }
        .btn-secondary:hover:not(:disabled) { background: #3f3f46; }
        .btn-ghost { background: transparent; color: #a1a1aa; }
        .btn-ghost:hover:not(:disabled) { background: #27272a; color: #fafafa; }
        .btn-full { width: 100%; }
        .btn-loading { cursor: wait; }
        .spinner {
          width: 16px; height: 16px; border: 2px solid rgba(255,255,255,0.3);
          border-top-color: #fff; border-radius: 50%;
          animation: spin 0.6s linear infinite;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
      `}</style>
    </>
  );
}
