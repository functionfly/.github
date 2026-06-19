/**
 * PageSkeleton - Content-shaped skeleton loading for initial page loads
 * 
 * Provides a more polished loading experience than spinners by showing
 * placeholder content in the shape of the actual page layout.
 */

import { cn } from '@/lib/utils';

interface PageSkeletonProps {
  /** Optional className for custom styling */
  className?: string;
  /** Show header skeleton */
  showHeader?: boolean;
  /** Show sidebar skeleton */
  showSidebar?: boolean;
  /** Number of content cards/rows to show */
  cardCount?: number;
}

export function PageSkeleton({
  className,
  showHeader = true,
  showSidebar = false,
  cardCount = 3,
}: PageSkeletonProps) {
  return (
    <div className={cn('animate-pulse space-y-6 p-6', className)}>
      {showHeader && (
        <div className="space-y-3">
          <div className="h-8 w-48 bg-muted rounded-md" />
          <div className="h-4 w-64 bg-muted rounded-md" />
        </div>
      )}
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: cardCount }).map((_, i) => (
          <div key={i} className="h-32 bg-muted rounded-lg" />
        ))}
      </div>
      
      <div className="space-y-3">
        <div className="h-4 w-full bg-muted rounded" />
        <div className="h-4 w-5/6 bg-muted rounded" />
        <div className="h-4 w-4/6 bg-muted rounded" />
      </div>
      
      {showSidebar && (
        <div className="space-y-3">
          <div className="h-6 w-32 bg-muted rounded" />
          <div className="h-24 w-full bg-muted rounded-lg" />
          <div className="h-24 w-full bg-muted rounded-lg" />
        </div>
      )}
    </div>
  );
}

/**
 * DashboardPageSkeleton - Skeleton for dashboard-style pages
 */
export function DashboardPageSkeleton() {
  return (
    <div className="animate-pulse space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <div className="h-8 w-48 bg-muted rounded-md" />
          <div className="h-4 w-64 bg-muted rounded-md" />
        </div>
        <div className="h-10 w-32 bg-muted rounded-md" />
      </div>
      
      <div className="grid gap-4 md:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="h-28 bg-muted rounded-lg" />
        ))}
      </div>
      
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="h-64 bg-muted rounded-lg" />
        <div className="h-64 bg-muted rounded-lg" />
      </div>
      
      <div className="space-y-3">
        <div className="h-6 w-40 bg-muted rounded" />
        <div className="h-48 w-full bg-muted rounded-lg" />
      </div>
    </div>
  );
}

/**
 * TablePageSkeleton - Skeleton for table/list pages
 */
export function TablePageSkeleton({ rowCount = 5 }: { rowCount?: number }) {
  return (
    <div className="animate-pulse space-y-4 p-6">
      <div className="flex items-center justify-between">
        <div className="h-8 w-48 bg-muted rounded-md" />
        <div className="h-10 w-40 bg-muted rounded-md" />
      </div>
      
      <div className="space-y-2">
        <div className="flex gap-4 p-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-4 bg-muted rounded flex-1" />
          ))}
        </div>
        {Array.from({ length: rowCount }).map((_, i) => (
          <div key={i} className="flex gap-4 p-4 border-t border-border">
            {Array.from({ length: 5 }).map((_, j) => (
              <div key={j} className="h-4 bg-muted rounded flex-1" />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

/**
 * DetailPageSkeleton - Skeleton for detail/view pages
 */
export function DetailPageSkeleton() {
  return (
    <div className="animate-pulse space-y-6 p-6">
      <div className="flex items-center gap-2">
        <div className="h-4 w-4 bg-muted rounded" />
        <div className="h-4 w-24 bg-muted rounded" />
      </div>
      
      <div className="space-y-2">
        <div className="h-8 w-64 bg-muted rounded-md" />
        <div className="h-4 w-48 bg-muted rounded-md" />
      </div>
      
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-4">
          <div className="h-64 bg-muted rounded-lg" />
          <div className="h-32 bg-muted rounded-lg" />
        </div>
        <div className="space-y-4">
          <div className="h-48 bg-muted rounded-lg" />
          <div className="h-48 bg-muted rounded-lg" />
        </div>
      </div>
    </div>
  );
}
