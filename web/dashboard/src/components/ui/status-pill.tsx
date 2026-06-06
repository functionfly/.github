import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const statusPillVariants = cva(
  'inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase tracking-wider transition-colors',
  {
    variants: {
      tone: {
        // Active / verified — calm emerald with a hairline border, not a pill glow
        success:
          'border-emerald-500/30 bg-emerald-500/[0.06] text-emerald-400',
        // Failed / destructive — restrained, not a red badge
        danger:
          'border-red-500/30 bg-red-500/[0.06] text-red-400',
        // In progress / pending — amber, dimmer
        warning:
          'border-amber-500/30 bg-amber-500/[0.06] text-amber-400',
        // Neutral / abandoned / pending payment
        neutral:
          'border-white/10 bg-white/[0.03] text-text-secondary',
        // Brand accent (current/exam in progress)
        brand:
          'border-brand-500/30 bg-brand-500/[0.08] text-brand-400',
        // Info (informational metadata)
        info:
          'border-sky-500/30 bg-sky-500/[0.06] text-sky-400',
      },
    },
    defaultVariants: {
      tone: 'neutral',
    },
  }
);

export interface StatusPillProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof statusPillVariants> {
  /** Show a small status dot before the label (uses tone color). */
  withDot?: boolean;
}

function StatusPill({ className, tone, withDot, children, ...props }: StatusPillProps) {
  const dotColor: Record<NonNullable<VariantProps<typeof statusPillVariants>['tone']>, string> = {
    success: 'bg-emerald-400',
    danger: 'bg-red-400',
    warning: 'bg-amber-400',
    neutral: 'bg-text-muted',
    brand: 'bg-brand-400',
    info: 'bg-sky-400',
  };

  return (
    <span
      className={cn(statusPillVariants({ tone }), className)}
      {...props}
    >
      {withDot && (
        <span
          aria-hidden
          className={cn('h-1.5 w-1.5 rounded-full', dotColor[tone ?? 'neutral'])}
        />
      )}
      {children}
    </span>
  );
}

export { StatusPill, statusPillVariants };
