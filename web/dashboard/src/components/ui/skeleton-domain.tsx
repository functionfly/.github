'use client';

import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const skeletonCardVariants = cva('', {
  variants: {
    variant: {
      default: '',
      glass: 'glass-card',
      gradient: 'gradient-border',
    },
  },
  defaultVariants: {
    variant: 'default',
  },
});

export interface SkeletonCardProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof skeletonCardVariants> {
  header?: boolean;
  footer?: boolean;
  contentLines?: number;
  showImage?: boolean;
  imageHeight?: number;
}

const SkeletonCard = React.forwardRef<HTMLDivElement, SkeletonCardProps>(
  ({ className, variant, header = true, footer = false, contentLines = 3, showImage = false, imageHeight = 160, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'rounded-lg border border-border-default bg-card p-6 shadow-sm',
          skeletonCardVariants({ variant, className })
        )}
        {...props}
      >
        {header && (
          <div className="mb-4 flex items-center justify-between">
            <div className="h-5 w-1/3 animate-pulse rounded-md bg-bg-secondary" />
            <div className="h-4 w-16 animate-pulse rounded-md bg-bg-secondary" />
          </div>
        )}
        {showImage && (
          <div
            className="mb-4 w-full animate-pulse rounded-md bg-bg-secondary"
            style={{ height: imageHeight }}
          />
        )}
        <div className="space-y-3">
          {Array.from({ length: contentLines }).map((_, i) => (
            <div
              key={i}
              className={cn(
                'h-4 animate-pulse rounded-md bg-bg-secondary',
                i === contentLines - 1 ? 'w-2/3' : 'w-full'
              )}
            />
          ))}
        </div>
        {footer && (
          <div className="mt-6 flex items-center justify-between pt-4 border-t border-border-subtle">
            <div className="h-9 w-24 animate-pulse rounded-md bg-bg-secondary" />
            <div className="h-9 w-9 animate-pulse rounded-full bg-bg-secondary" />
          </div>
        )}
      </div>
    );
  }
);
SkeletonCard.displayName = 'SkeletonCard';

const SkeletonListVariants = cva('', {
  variants: {
    variant: {
      default: '',
      compact: 'py-2',
      spacious: 'py-4',
    },
  },
  defaultVariants: {
    variant: 'default',
  },
});

export interface SkeletonListItemProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof SkeletonListVariants> {
  showAvatar?: boolean;
  showMeta?: boolean;
  showAction?: boolean;
  lines?: number;
}

const SkeletonListItem = React.forwardRef<HTMLDivElement, SkeletonListItemProps>(
  ({ className, variant, showAvatar = true, showMeta = true, showAction = true, lines = 2, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'flex items-center gap-4 py-3 px-4 border-b border-border-subtle last:border-0',
          SkeletonListVariants({ variant, className })
        )}
        {...props}
      >
        {showAvatar && (
          <div className="h-10 w-10 flex-shrink-0 animate-pulse rounded-full bg-bg-secondary" />
        )}
        <div className="flex-1 min-w-0">
          <div className="h-4 w-1/3 animate-pulse rounded-md bg-bg-secondary mb-2" />
          {Array.from({ length: lines - 1 }).map((_, i) => (
            <div
              key={i}
              className={cn(
                'h-3 animate-pulse rounded-md bg-bg-secondary/70',
                i === lines - 2 ? 'w-2/3' : 'w-full',
                i > 0 && 'mt-1'
              )}
            />
          ))}
        </div>
        {showMeta && (
          <div className="hidden sm:block">
            <div className="h-3 w-20 animate-pulse rounded-md bg-bg-secondary/50" />
          </div>
        )}
        {showAction && (
          <div className="h-8 w-8 animate-pulse rounded-md bg-bg-secondary" />
        )}
      </div>
    );
  }
);
SkeletonListItem.displayName = 'SkeletonListItem';

export interface SkeletonListProps extends React.HTMLAttributes<HTMLDivElement> {
  count?: number;
  itemProps?: SkeletonListItemProps;
}

const SkeletonList = React.forwardRef<HTMLDivElement, SkeletonListProps>(
  ({ className, count = 5, itemProps, ...props }, ref) => {
    return (
      <div ref={ref} className={cn('divide-y divide-border-subtle', className)} {...props}>
        {Array.from({ length: count }).map((_, i) => (
          <SkeletonListItem key={i} {...itemProps} />
        ))}
      </div>
    );
  }
);
SkeletonList.displayName = 'SkeletonList';

const SkeletonTableVariants = cva('', {
  variants: {
    variant: {
      default: '',
      compact: '',
    },
  },
  defaultVariants: {
    variant: 'default',
  },
});

export interface SkeletonTableProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof SkeletonTableVariants> {
  columns?: number;
  rows?: number;
  showHeader?: boolean;
}

const SkeletonTable = React.forwardRef<HTMLDivElement, SkeletonTableProps>(
  ({ className, variant, columns = 5, rows = 5, showHeader = true, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'w-full rounded-lg border border-border-default overflow-hidden',
          SkeletonTableVariants({ variant, className })
        )}
        {...props}
      >
        <table className="w-full">
          {showHeader && (
            <thead className="bg-bg-secondary">
              <tr>
                {Array.from({ length: columns }).map((_, i) => (
                  <th key={i} className="px-4 py-3">
                    <div className="h-4 animate-pulse rounded-md bg-bg-secondary/50 w-24" />
                  </th>
                ))}
              </tr>
            </thead>
          )}
          <tbody className="divide-y divide-border-subtle">
            {Array.from({ length: rows }).map((_, rowIndex) => (
              <tr key={rowIndex} className="bg-card">
                {Array.from({ length: columns }).map((_, colIndex) => (
                  <td key={colIndex} className="px-4 py-3">
                    <div
                      className={cn(
                        'h-4 animate-pulse rounded-md bg-bg-secondary',
                        colIndex === 0 ? 'w-full' : colIndex === columns - 1 ? 'w-16' : 'w-24'
                      )}
                    />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }
);
SkeletonTable.displayName = 'SkeletonTable';

export interface SkeletonStatsProps extends React.HTMLAttributes<HTMLDivElement> {
  count?: number;
  showIcon?: boolean;
  showTrend?: boolean;
}

const SkeletonStats = React.forwardRef<HTMLDivElement, SkeletonStatsProps>(
  ({ className, count = 4, showIcon = true, showTrend = true, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'grid gap-4',
          count === 1 ? 'grid-cols-1' :
          count === 2 ? 'grid-cols-1 sm:grid-cols-2' :
          count === 3 ? 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3' :
          'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4',
          className
        )}
        {...props}
      >
        {Array.from({ length: count }).map((_, i) => (
          <div
            key={i}
            className="rounded-lg border border-border-default bg-card p-6 shadow-sm"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="h-4 w-24 animate-pulse rounded-md bg-bg-secondary mb-2" />
                <div className="h-8 w-16 animate-pulse rounded-md bg-bg-secondary" />
              </div>
              {showIcon && (
                <div className="h-10 w-10 animate-pulse rounded-lg bg-bg-secondary" />
              )}
            </div>
            {showTrend && (
              <div className="mt-4 flex items-center gap-2">
                <div className="h-4 w-12 animate-pulse rounded-md bg-bg-secondary" />
                <div className="h-3 w-20 animate-pulse rounded-md bg-bg-secondary/50" />
              </div>
            )}
          </div>
        ))}
      </div>
    );
  }
);
SkeletonStats.displayName = 'SkeletonStats';

export interface SkeletonFormProps extends React.HTMLAttributes<HTMLDivElement> {
  fields?: number;
  showSubmit?: boolean;
}

const SkeletonForm = React.forwardRef<HTMLDivElement, SkeletonFormProps>(
  ({ className, fields = 4, showSubmit = true, ...props }, ref) => {
    return (
      <div ref={ref} className={cn('space-y-6', className)} {...props}>
        {Array.from({ length: fields }).map((_, i) => (
          <div key={i} className="space-y-2">
            <div className="h-4 w-24 animate-pulse rounded-md bg-bg-secondary" />
            <div className="h-10 w-full animate-pulse rounded-md bg-bg-secondary" />
            {i === 0 && <div className="h-3 w-48 animate-pulse rounded-md bg-bg-secondary/50" />}
          </div>
        ))}
        {showSubmit && (
          <div className="flex items-center justify-between pt-4">
            <div className="h-10 w-32 animate-pulse rounded-md bg-bg-secondary" />
            <div className="h-10 w-24 animate-pulse rounded-md bg-bg-secondary" />
          </div>
        )}
      </div>
    );
  }
);
SkeletonForm.displayName = 'SkeletonForm';

export interface SkeletonChartProps extends React.HTMLAttributes<HTMLDivElement> {
  height?: number;
  showHeader?: boolean;
  showLegend?: boolean;
}

const SkeletonChart = React.forwardRef<HTMLDivElement, SkeletonChartProps>(
  ({ className, height = 300, showHeader = true, showLegend = true, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'rounded-lg border border-border-default bg-card p-6 shadow-sm',
          className
        )}
        {...props}
      >
        {showHeader && (
          <div className="mb-4 flex items-center justify-between">
            <div className="h-5 w-32 animate-pulse rounded-md bg-bg-secondary" />
            <div className="h-8 w-24 animate-pulse rounded-md bg-bg-secondary" />
          </div>
        )}
        <div
          className="w-full animate-pulse rounded-md bg-bg-secondary/50"
          style={{ height }}
        >
          {/* Chart area placeholder with gradient effect */}
          <div className="relative h-full w-full overflow-hidden rounded-md">
            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-bg-secondary/30 to-transparent animate-pulse" />
            <svg className="absolute inset-0 h-full w-full" preserveAspectRatio="none">
              <path
                d="M0,250 Q50,200 100,220 T200,180 T300,150 T400,200 T500,120 T600,100 T700,140 T800,80 T900,100 L900,300 L0,300 Z"
                className="fill-bg-secondary/30"
              />
              <path
                d="M0,250 Q50,200 100,220 T200,180 T300,150 T400,200 T500,120 T600,100 T700,140 T800,80 T900,100"
                fill="none"
                className="stroke-bg-secondary"
                strokeWidth="2"
              />
            </svg>
          </div>
        </div>
        {showLegend && (
          <div className="mt-4 flex items-center justify-center gap-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-center gap-2">
                <div className="h-3 w-3 animate-pulse rounded-full bg-bg-secondary" />
                <div className="h-3 w-16 animate-pulse rounded-md bg-bg-secondary" />
              </div>
            ))}
          </div>
        )}
      </div>
    );
  }
);
SkeletonChart.displayName = 'SkeletonChart';

export interface SkeletonAvatarProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

const SkeletonAvatar = React.forwardRef<HTMLDivElement, SkeletonAvatarProps>(
  ({ className, size = 'md', ...props }, ref) => {
    const sizeClasses = {
      sm: 'h-8 w-8',
      md: 'h-10 w-10',
      lg: 'h-16 w-16',
      xl: 'h-24 w-24',
    };

    return (
      <div
        ref={ref}
        className={cn(
          'animate-pulse rounded-full bg-bg-secondary',
          sizeClasses[size],
          className
        )}
        {...props}
      />
    );
  }
);
SkeletonAvatar.displayName = 'SkeletonAvatar';

export {
  SkeletonCard,
  SkeletonList,
  SkeletonListItem,
  SkeletonTable,
  SkeletonStats,
  SkeletonForm,
  SkeletonChart,
  SkeletonAvatar,
  skeletonCardVariants,
  SkeletonListVariants,
  SkeletonTableVariants,
};
