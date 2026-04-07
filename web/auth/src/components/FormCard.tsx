import type { ReactNode } from "react";
import React from "react";

interface Props {
  title: string;
  subtitle?: string;
  children: ReactNode;
  error?: string;
  success?: string;
  footer?: ReactNode;
}

export default function FormCard({
  title,
  subtitle,
  children,
  error,
  success,
  footer,
}: Props) {
  return (
    <div className="card">
      <div className="card-header">
        <h1 className="card-title">{title}</h1>
        {subtitle && <p className="card-subtitle">{subtitle}</p>}
      </div>

      {error && (
        <div className="alert alert-error" role="alert">
          {error}
        </div>
      )}
      {success && (
        <div className="alert alert-success" role="status">
          {success}
        </div>
      )}

      {children}

      {footer && <div className="card-footer">{footer}</div>}

      <style>{`
        .card {
          background: #18181b;
          border: 1px solid #27272a;
          border-radius: 12px;
          padding: 2rem;
          width: 100%;
          max-width: 100%;
          box-sizing: border-box;
          overflow: hidden;
        }
        .card-header { margin-bottom: 1.5rem; }
        .card-title {
          font-size: 1.375rem;
          font-weight: 600;
          color: #fafafa;
          letter-spacing: -0.02em;
        }
        .card-subtitle {
          margin-top: 0.375rem;
          color: #71717a;
          font-size: 0.9375rem;
        }
        .alert {
          padding: 0.75rem 1rem;
          border-radius: 8px;
          font-size: 0.875rem;
          margin-bottom: 1rem;
        }
        .alert-error {
          background: rgba(239,68,68,0.1);
          border: 1px solid rgba(239,68,68,0.3);
          color: #fca5a5;
        }
        .alert-success {
          background: rgba(34,197,94,0.1);
          border: 1px solid rgba(34,197,94,0.3);
          color: #86efac;
        }
        .card-footer {
          margin-top: 1.25rem;
          padding-top: 1.25rem;
          border-top: 1px solid #27272a;
          text-align: center;
          font-size: 0.875rem;
          color: #71717a;
        }
      `}</style>
    </div>
  );
}
