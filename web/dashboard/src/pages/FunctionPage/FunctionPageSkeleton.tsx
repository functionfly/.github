import { Skeleton } from '@/components/ui/skeleton';

export function FunctionPageSkeleton() {
  return (
    <div className="function-page-layout">
      <aside className="function-page-toc-wrapper" style={{ visibility: 'hidden' }}>
        <div className="space-y-2 pt-8">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="h-4 w-24" />
          ))}
        </div>
      </aside>
      <div className="function-page-content">
        {/* Header skeleton */}
        <div className="function-page-section space-y-4">
          <div className="flex items-center gap-4">
            <Skeleton className="h-10 w-10 rounded-full" />
            <div className="space-y-2">
              <Skeleton className="h-6 w-48" />
              <Skeleton className="h-4 w-32" />
            </div>
          </div>
          <div className="flex gap-3">
            <Skeleton className="h-5 w-16 rounded-full" />
            <Skeleton className="h-5 w-20 rounded-full" />
            <Skeleton className="h-5 w-14 rounded-full" />
          </div>
          <Skeleton className="h-4 w-full max-w-md" />
          <Skeleton className="h-4 w-full max-w-sm" />
          <div className="flex gap-3 pt-2">
            <Skeleton className="h-10 w-32 rounded-[var(--radius)]" />
            <Skeleton className="h-10 w-28 rounded-[var(--radius)]" />
            <Skeleton className="h-10 w-28 rounded-[var(--radius)]" />
          </div>
        </div>

        {/* Overview skeleton */}
        <div className="function-page-section function-page-section--delayed space-y-4">
          <Skeleton className="h-40 w-full rounded-[var(--radius-lg)]" />
        </div>

        {/* Stats skeleton */}
        <div className="function-page-section function-page-section--delayed space-y-4">
          <Skeleton className="h-6 w-40" />
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-24 rounded-[var(--radius-lg)]" />
            ))}
          </div>
        </div>

        {/* Trust skeleton */}
        <div className="function-page-section function-page-section--delayed-2 space-y-4">
          <Skeleton className="h-6 w-36" />
          <Skeleton className="h-32 w-full rounded-[var(--radius-lg)]" />
        </div>

        {/* Activity skeleton */}
        <div className="function-page-section function-page-section--delayed-3 space-y-4">
          <Skeleton className="h-6 w-32" />
          <div className="space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-14 w-full rounded-[var(--radius)]" />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
