import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function ServiceCardSkeleton() {
  return (
    <Card className="border-border-subtle bg-bg-tertiary/50">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-center gap-3">
            <Skeleton className="h-9 w-9 rounded-lg" />
            <div>
              <Skeleton className="h-5 w-32 mb-1" />
              <Skeleton className="h-4 w-20" />
            </div>
          </div>
          <Skeleton className="h-6 w-24 rounded-full" />
        </div>
      </CardHeader>
      <CardContent className="pb-4">
        <Skeleton className="h-4 w-full mb-4" />
        <div className="grid grid-cols-2 gap-3">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-4 w-16" />
        </div>
        <Skeleton className="h-6 w-full mt-4 rounded" />
      </CardContent>
    </Card>
  );
}

export function IncidentSkeleton() {
  return (
    <Card className="border-border-subtle bg-bg-tertiary/50">
      <div className="p-4">
        <div className="flex items-start gap-4">
          <Skeleton className="w-3 h-3 rounded-full mt-1" />
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-1">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-5 w-16 rounded-full" />
            </div>
            <Skeleton className="h-4 w-full mb-2" />
            <div className="flex gap-4">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-24" />
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
}

export function MetricsSectionSkeleton() {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-5 gap-4">
      {[1, 2, 3, 4, 5].map((i) => (
        <Card key={i} className="border-border-subtle bg-bg-tertiary/50 p-4">
          <div className="flex items-start justify-between mb-3">
            <Skeleton className="h-9 w-9 rounded-lg" />
            <Skeleton className="h-4 w-12" />
          </div>
          <Skeleton className="h-8 w-20 mb-1" />
          <Skeleton className="h-4 w-24" />
        </Card>
      ))}
    </div>
  );
}

export function HeroSkeleton() {
  return (
    <Card className="border-border-subtle bg-bg-tertiary/50 p-8 md:p-12">
      <div className="flex flex-col md:flex-row items-center gap-6 md:gap-8">
        <Skeleton className="w-16 h-16 rounded-full" />
        <div className="flex-1 text-center md:text-left">
          <Skeleton className="h-10 w-64 mb-2" />
          <Skeleton className="h-6 w-full max-w-md" />
        </div>
        <Skeleton className="h-5 w-32" />
      </div>
    </Card>
  );
}
