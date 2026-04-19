import type { ReactNode } from "react";
import React from "react";
import { cn } from "../lib/utils";

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
    <div className={cn("ff-card", className)}>
      <div className="mb-6">
        <h1 className="text-xl font-semibold text-[var(--ff-primary-text)] tracking-tight">
          {title}
        </h1>
        {subtitle && (
          <p className="mt-1.5 text-[var(--ff-secondary-text)] text-[0.9375rem]">
            {subtitle}
          </p>
        )}
        {tagline && (
          <p className="mt-3 text-[var(--ff-muted-text)] text-sm italic border-l-2 border-[var(--ff-flame)] pl-3">
            {tagline}
          </p>
        )}
      </div>

      {error && (
        <div 
          className="ff-status-banner ff-status-banner--error mb-4" 
          role="alert"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
          {error}
        </div>
      )}
      
      {success && (
        <div 
          className="ff-status-banner ff-status-banner--success mb-4" 
          role="status"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="20 6 9 17 4 12" />
          </svg>
          {success}
        </div>
      )}

      {children}

      {footer && (
        <div className="mt-5 pt-5 text-center text-sm text-[var(--ff-muted-text)] border-t border-[var(--ff-border-default)]">
          {footer}
        </div>
      )}
    </div>
  );
}
