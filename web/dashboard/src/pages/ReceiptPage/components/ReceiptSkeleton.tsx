// ReceiptSkeleton — shimmer placeholders matching the real layout to
// avoid CLS during the initial fetch.
import { Card, CardContent, CardHeader } from "@/components/ui/card";

import { ReceiptPoweredBy } from "./ReceiptPoweredBy";

export function ReceiptSkeleton() {
  return (
    <div className="space-y-6" data-testid="receipt-skeleton">
      <div className="flex items-start gap-4">
        <div className="h-12 w-12 animate-pulse rounded-full bg-muted" aria-hidden />
        <div className="flex-1 space-y-2">
          <div className="h-7 w-2/3 animate-pulse rounded bg-muted" aria-hidden />
          <div className="h-4 w-1/2 animate-pulse rounded bg-muted" aria-hidden />
          <div className="h-3 w-1/3 animate-pulse rounded bg-muted" aria-hidden />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Card key={i}>
            <CardContent className="pt-6">
              <div className="h-7 w-16 animate-pulse rounded bg-muted" aria-hidden />
              <div className="mt-2 h-3 w-20 animate-pulse rounded bg-muted" aria-hidden />
            </CardContent>
          </Card>
        ))}
      </div>
      <Card>
        <CardHeader>
          <div className="h-5 w-32 animate-pulse rounded bg-muted" aria-hidden />
          <div className="h-3 w-1/2 animate-pulse rounded bg-muted" aria-hidden />
        </CardHeader>
        <CardContent>
          <div className="h-40 w-full animate-pulse rounded bg-muted" aria-hidden />
        </CardContent>
      </Card>
      <ReceiptPoweredBy />
    </div>
  );
}

export default ReceiptSkeleton;
