import * as React from 'react';
import { cn } from '@/lib/utils';

interface SectionHeaderProps {
  /** Small uppercase monospace label, e.g. "01 · ACTIVE CREDENTIALS" */
  eyebrow?: string;
  /** Main heading */
  title: string;
  /** Optional supporting text */
  description?: string;
  /** Right-aligned actions */
  actions?: React.ReactNode;
  className?: string;
  /** Render with a hairline border below (terminal section divider) */
  divider?: boolean;
  id?: string;
}

/**
 * Editorial section header used across the credentials surface.
 * Mimics a financial terminal: monospace eyebrow + display title +
 * hairline rule. Substitutes the typical "h3 with icon" pattern with
 * something more deliberate and document-like.
 */
export function SectionHeader({
  eyebrow,
  title,
  description,
  actions,
  className,
  divider = true,
  id,
}: SectionHeaderProps) {
  return (
    <header
      className={cn(
        'flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between',
        divider && 'border-b border-white/[0.06] pb-4',
        className
      )}
    >
      <div className="space-y-1.5">
        {eyebrow && (
          <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-text-muted">
            {eyebrow}
          </p>
        )}
        <h2
          id={id}
          className="font-display text-xl font-medium tracking-tight text-text-primary sm:text-2xl"
        >
          {title}
        </h2>
        {description && (
          <p className="max-w-2xl text-sm text-text-secondary">{description}</p>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </header>
  );
}
