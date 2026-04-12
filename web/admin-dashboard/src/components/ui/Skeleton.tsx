import React from 'react';
import { cn } from '@/lib/utils';

interface SkeletonProps {
  className?: string;
  variant?: 'text' | 'circular' | 'rectangular' | 'rounded';
  width?: string | number;
  height?: string | number;
  animation?: 'pulse' | 'wave' | 'none';
}

export function Skeleton({
  className,
  variant = 'text',
  width,
  height,
  animation = 'pulse',
}: SkeletonProps) {
  const variantClasses = {
    text: 'rounded',
    circular: 'rounded-full',
    rectangular: 'rounded-none',
    rounded: 'rounded-lg',
  };

  const animationClasses = {
    pulse: 'animate-pulse',
    wave: 'animate-shimmer',
    none: '',
  };

  const style: React.CSSProperties = {
    width: width ? (typeof width === 'number' ? `${width}px` : width) : undefined,
    height: height ? (typeof height === 'number' ? `${height}px` : height) : undefined,
  };

  return (
    <div
      className={cn(
        'bg-gray-200 dark:bg-gray-700',
        variantClasses[variant],
        animationClasses[animation],
        className
      )}
      style={style}
      aria-hidden="true"
    />
  );
}

// Skeleton Card for content cards
interface SkeletonCardProps {
  className?: string;
  hasImage?: boolean;
  hasMeta?: boolean;
  lines?: number;
}

export function SkeletonCard({
  className,
  hasImage = false,
  hasMeta = false,
  lines = 2,
}: SkeletonCardProps) {
  return (
    <div className={cn('p-4 rounded-lg border border-gray-200 bg-white', className)}>
      {hasImage && (
        <Skeleton variant="rounded" width="100%" height={160} className="mb-4" />
      )}
      <Skeleton width="70%" height={24} className="mb-3" />
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          width={i === lines - 1 ? '60%' : '100%'}
          height={16}
          className="mb-2"
        />
      ))}
      {hasMeta && (
        <div className="flex items-center gap-2 mt-4 pt-4 border-t border-gray-100">
          <Skeleton variant="circular" width={32} height={32} />
          <Skeleton width={100} height={16} />
        </div>
      )}
    </div>
  );
}

// Skeleton Table Row for data tables
interface SkeletonTableRowProps {
  columns: number;
  className?: string;
  hasActions?: boolean;
}

export function SkeletonTableRow({
  columns,
  className,
  hasActions = false,
}: SkeletonTableRowProps) {
  return (
    <div className={cn('flex items-center gap-4 p-4', className)}>
      {Array.from({ length: columns }).map((_, i) => (
        <Skeleton
          key={i}
          width={i === 0 ? '40%' : `${80 / columns}%`}
          height={16}
          className="flex-1"
        />
      ))}
      {hasActions && <Skeleton width={80} height={32} variant="rounded" />}
    </div>
  );
}

// Skeleton Stat Card for dashboard stats
interface SkeletonStatProps {
  className?: string;
}

export function SkeletonStat({ className }: SkeletonStatProps) {
  return (
    <div className={cn('p-6 rounded-lg border border-gray-200 bg-white', className)}>
      <div className="flex items-center justify-between">
        <div>
          <Skeleton width={80} height={14} className="mb-2" />
          <Skeleton width={120} height={32} />
        </div>
        <Skeleton variant="circular" width={48} height={48} />
      </div>
    </div>
  );
}

// Skeleton List for list views
interface SkeletonListProps {
  items?: number;
  className?: string;
  hasAvatar?: boolean;
}

export function SkeletonList({
  items = 5,
  className,
  hasAvatar = false,
}: SkeletonListProps) {
  return (
    <div className={cn('space-y-3', className)}>
      {Array.from({ length: items }).map((_, i) => (
        <div key={i} className="flex items-center gap-3 p-3">
          {hasAvatar && <Skeleton variant="circular" width={40} height={40} />}
          <div className="flex-1">
            <Skeleton width="60%" height={16} className="mb-2" />
            <Skeleton width="40%" height={12} />
          </div>
        </div>
      ))}
    </div>
  );
}
