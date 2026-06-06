// ReceiptStats — 4 metric cards (duration, cached, executed-at, views).
import { Calendar, Clock, Eye, Zap } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";

interface ReceiptStatsProps {
  durationMs: number;
  cached: boolean;
  createdAt: string;
  viewCount?: number;
}

function formatTimestamp(iso: string): string {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

export function ReceiptStats({ durationMs, cached, createdAt, viewCount }: ReceiptStatsProps) {
  return (
    <div className="grid grid-cols-2 gap-3 md:grid-cols-4" data-testid="receipt-stats">
      <Card>
        <CardContent className="flex items-center gap-3 pt-6">
          <Clock className="h-5 w-5 text-muted-foreground" aria-hidden />
          <div>
            <p className="text-xl font-semibold tabular-nums">{durationMs}ms</p>
            <p className="text-xs text-muted-foreground">Execution time</p>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent className="flex items-center gap-3 pt-6">
          <Zap className="h-5 w-5 text-muted-foreground" aria-hidden />
          <div>
            <p className="text-xl font-semibold">{cached ? "Cached" : "Live"}</p>
            <p className="text-xs text-muted-foreground">Execution type</p>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent className="flex items-center gap-3 pt-6">
          <Calendar className="h-5 w-5 text-muted-foreground" aria-hidden />
          <div>
            <p className="text-sm font-semibold" title={createdAt}>
              {formatTimestamp(createdAt)}
            </p>
            <p className="text-xs text-muted-foreground">Executed at</p>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent className="flex items-center gap-3 pt-6">
          <Eye className="h-5 w-5 text-muted-foreground" aria-hidden />
          <div>
            <p className="text-xl font-semibold tabular-nums">{(viewCount ?? 0).toLocaleString()}</p>
            <p className="text-xs text-muted-foreground">Public views</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default ReceiptStats;
