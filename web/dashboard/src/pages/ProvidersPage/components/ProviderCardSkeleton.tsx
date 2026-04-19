import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';

export function ProviderCardSkeleton() {
  return (
    <Card className="relative overflow-hidden">
      <div className="absolute top-0 left-0 right-0 h-1 bg-slate-200 dark:bg-slate-700" />
      <CardContent className="p-6">
        <div className="flex items-start gap-4 mb-5">
          <Skeleton className="w-14 h-14 rounded-2xl" />
          <div className="flex-1">
            <Skeleton className="h-5 w-32 mb-2" />
            <Skeleton className="h-4 w-20" />
          </div>
        </div>
        <Skeleton className="h-4 w-40 mb-5" />
        <Skeleton className="h-10 w-full" />
      </CardContent>
    </Card>
  );
}
