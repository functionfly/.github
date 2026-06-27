import type { ReactNode } from "react";
import { cn } from "../lib/utils";
import { StatusPill } from "./sc";

interface Props {
  title: string;
  subtitle?: string;
  tagline?: string;
  children: ReactNode;
  error?: string;
  success?: string;
  footer?: ReactNode;
  className?: string;
}

export default function FormCard({
  title,
  subtitle,
  tagline,
  children,
  error,
  success,
  footer,
  className,
}: Props) {
  return (
    <div className={cn("form-card", className)}>
      <div className="form-card-header">
        <h1
          className="form-card-title"
          style={{
            fontFamily: "var(--font-display)",
            fontSize: "22px",
            fontWeight: 500,
            lineHeight: 1.25,
            color: "var(--text)",
            letterSpacing: 0,
          }}
        >
          {title}
        </h1>
        {subtitle && (
          <p
            className="form-card-subtitle"
            style={{
              fontFamily: "var(--font-body)",
              fontSize: "15px",
              lineHeight: 1.6,
              color: "var(--text-dim)",
              marginTop: "var(--space-2)",
            }}
          >
            {subtitle}
          </p>
        )}
        {tagline && (
          <p
            className="form-card-tagline"
            style={{
              fontFamily: "var(--font-body)",
              fontSize: "13px",
              lineHeight: 1.6,
              color: "var(--text-faint)",
              fontStyle: "italic",
              borderLeft: "2px solid var(--accent)",
              paddingLeft: "var(--space-3)",
              marginTop: "var(--space-4)",
            }}
          >
            {tagline}
          </p>
        )}
      </div>

      {error && (
        <div className="form-card-status" role="alert">
          <StatusPill status="revoked" label="Error" />
          <span
            style={{
              fontFamily: "var(--font-body)",
              fontSize: "14px",
              color: "var(--text)",
              marginLeft: "var(--space-2)",
            }}
          >
            {error}
          </span>
        </div>
      )}

      {success && (
        <div className="form-card-status" role="status">
          <StatusPill status="live" label="Success" />
          <span
            style={{
              fontFamily: "var(--font-body)",
              fontSize: "14px",
              color: "var(--text)",
              marginLeft: "var(--space-2)",
            }}
          >
            {success}
          </span>
        </div>
      )}

      {children}

      {footer && (
        <div
          className="form-card-footer"
          style={{
            marginTop: "var(--space-5)",
            paddingTop: "var(--space-5)",
            textAlign: "center",
            fontFamily: "var(--font-body)",
            fontSize: "14px",
            color: "var(--text-faint)",
            borderTop: "1px solid var(--panel-edge)",
          }}
        >
          {footer}
        </div>
      )}

      <style>{`
        .form-card {
          padding: 0;
        }
        .form-card-status {
          display: flex;
          align-items: center;
          margin-bottom: var(--space-4);
        }
      `}</style>
    </div>
  );
}
