/**
 * Skeleton Components for ProfilePage
 *
 * Loading states for ProfilePage components.
 */

import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent } from "@/components/ui/card";

/**
 * Loading state for ProfileHeader
 */
export function ProfileHeaderSkeleton() {
  return (
    <div className="animate-pulse">
      {/* Cover */}
      <div className="h-48 md:h-64 bg-gradient-to-r from-gray-700 to-gray-800 rounded-t-xl" />

      <div className="px-4 md:px-8 pb-6">
        <div className="flex flex-col md:flex-row md:items-end -mt-16 md:-mt-20 gap-4 md:gap-6">
          {/* Avatar */}
          <Skeleton className="w-32 h-32 md:w-40 md:h-40 rounded-full border-4 border-background shrink-0" />

          {/* Info */}
          <div className="flex-1 min-w-0 space-y-3">
            <Skeleton className="h-8 w-48" />
            <Skeleton className="h-5 w-32" />
            <Skeleton className="h-4 w-full max-w-md" />
            <div className="flex gap-4">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-24" />
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-2">
            <Skeleton className="h-10 w-24" />
            <Skeleton className="h-10 w-24" />
          </div>
        </div>
      </div>
    </div>
  );
}

/**
 * Loading state for StatsOverview
 */
export function StatsOverviewSkeleton() {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <Card key={i} className="bg-card">
          <CardContent className="p-4">
            <Skeleton className="h-4 w-20 mb-2" />
            <Skeleton className="h-8 w-16 mb-1" />
            <Skeleton className="h-3 w-12" />
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

/**
 * Loading state for TabContent
 */
export function TabContentSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <Skeleton className="h-8 w-48" />
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-40 rounded-lg" />
        ))}
      </div>
    </div>
  );
}
