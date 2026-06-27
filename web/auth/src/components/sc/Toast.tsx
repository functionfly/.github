import React, { useEffect, useState } from "react";

type ToastStatus = "ok" | "pending" | "revoked";

interface ToastProps {
  message: string;
  status?: ToastStatus;
  duration?: number;
  onDismiss: () => void;
}

const STATUS_COLORS: Record<ToastStatus, string> = {
  ok: "var(--status-ok)",
  pending: "var(--status-pending)",
  revoked: "var(--status-revoked)",
};

export const Toast: React.FC<ToastProps> = ({
  message,
  status = "ok",
  duration = 5000,
  onDismiss,
}) => {
  const [isPaused, setIsPaused] = useState(false);

  useEffect(() => {
    if (isPaused) return;
    const timer = setTimeout(onDismiss, duration);
    return () => clearTimeout(timer);
  }, [duration, onDismiss, isPaused]);

  return (
    <div
      role="alert"
      aria-live="polite"
      className="toast"
      onMouseEnter={() => setIsPaused(true)}
      onMouseLeave={() => setIsPaused(false)}
      style={{
        position: "fixed",
        bottom: "var(--space-5)",
        right: "var(--space-5)",
        zIndex: "var(--z-toast)",
        background: "var(--panel-raised)",
        border: "1px solid var(--panel-edge)",
        borderRadius: "var(--radius)",
        borderLeft: `3px solid ${STATUS_COLORS[status]}`,
        padding: "var(--space-3) var(--space-4)",
        minWidth: "280px",
        maxWidth: "400px",
        animation: "toastIn var(--duration-base) var(--ease-out)",
        boxShadow: "var(--shadow-chamber)",
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "flex-start",
          justifyContent: "space-between",
          gap: "var(--space-3)",
        }}
      >
        <span
          style={{
            fontFamily: "var(--font-body)",
            fontSize: "14px",
            color: "var(--text)",
            lineHeight: 1.5,
          }}
        >
          {message}
        </span>
        <button
          onClick={onDismiss}
          aria-label="Dismiss"
          style={{
            background: "none",
            border: "none",
            padding: 0,
            cursor: "pointer",
            color: "var(--text-faint)",
            flexShrink: 0,
            lineHeight: 1,
          }}
        >
          ✕
        </button>
      </div>
    </div>
  );
};
