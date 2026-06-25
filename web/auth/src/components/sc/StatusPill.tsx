import React from "react";

type Status = "live" | "ok" | "pending" | "revoked";

interface StatusPillProps {
  status: Status;
  label: string;
}

const STATUS_STYLES: Record<Status, { text: string; border: string; bg: string }> = {
  live: {
    text: "var(--status-ok)",
    border: "rgba(143,255,208,0.3)",
    bg: "rgba(143,255,208,0.06)",
  },
  ok: {
    text: "var(--status-ok)",
    border: "rgba(143,255,208,0.3)",
    bg: "rgba(143,255,208,0.06)",
  },
  pending: {
    text: "var(--status-pending)",
    border: "rgba(232,196,104,0.3)",
    bg: "rgba(232,196,104,0.06)",
  },
  revoked: {
    text: "var(--status-revoked)",
    border: "rgba(255,107,107,0.3)",
    bg: "rgba(255,107,107,0.06)",
  },
};

export const StatusPill: React.FC<StatusPillProps> = ({ status, label }) => {
  const styles = STATUS_STYLES[status];

  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: "var(--space-1)",
        fontFamily: "var(--font-mono)",
        fontSize: "11px",
        fontWeight: 500,
        textTransform: "uppercase",
        letterSpacing: "0.06em",
        color: styles.text,
        borderColor: styles.border,
        backgroundColor: styles.bg,
        padding: "3px 12px",
        borderRadius: "var(--radius-sm)",
        borderWidth: "1px",
        borderStyle: "solid",
      }}
    >
      <span
        style={{
          display: "inline-block",
          width: "6px",
          height: "6px",
          borderRadius: "50%",
          background: styles.text,
        }}
      />
      {label}
    </span>
  );
};
