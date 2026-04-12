'use client';

import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const emptyStateVariants = cva(
  'flex flex-col items-center justify-center text-center p-8 rounded-lg',
  {
    variants: {
      variant: {
        default: 'bg-bg-secondary/50 border border-border-subtle',
        ghost: 'bg-transparent',
        card: 'bg-card border border-border-default shadow-sm',
        glass: 'glass-card border border-white/10',
      },
      size: {
        default: 'py-12 px-8',
        sm: 'py-6 px-4',
        lg: 'py-16 px-12',
        full: 'h-full min-h-[300px]',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
);

export interface EmptyStateProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof emptyStateVariants> {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
  secondaryAction?: React.ReactNode;
}

const EmptyState = React.forwardRef<HTMLDivElement, EmptyStateProps>(
  (
    { className, variant, size, icon, title, description, action, secondaryAction, ...props },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(emptyStateVariants({ variant, size, className }))}
        {...props}
      >
        {icon && (
          <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full bg-brand-500/10 text-brand-500">
            {React.isValidElement(icon) && icon}
          </div>
        )}
        <h3 className="text-lg font-semibold text-text-primary">{title}</h3>
        {description && (
          <p className="mt-2 max-w-sm text-sm text-text-secondary">{description}</p>
        )}
        {(action || secondaryAction) && (
          <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
            {action}
            {secondaryAction}
          </div>
        )}
      </div>
    );
  }
);
EmptyState.displayName = 'EmptyState';

export { EmptyState, emptyStateVariants };
