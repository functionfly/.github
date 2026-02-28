import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { TrendSparkline } from "./TrendSparkline";

export interface MetricCardProps {
  title: string;
  value: string | number;
  /** Short label under the value, e.g. "vs last 7d" */
  changeLabel?: string;
  /** Percent change; positive = up, negative = down */
  changePercent?: number;
  /** Optional sparkline data (e.g. last 7–30 points) */
  sparklineData?: number[];
  icon?: React.ReactNode;
  className?: string;
}

export function MetricCard({
  title,
  value,
  changeLabel,
  changePercent,
  sparklineData,
  icon,
  className,
}: MetricCardProps) {
  const trend = changePercent == null ? "neutral" : changePercent >= 0 ? "up" : "down";

  return (
    <Card className={cn("border-theme bg-card overflow-hidden", className)}>
      <CardContent className="p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium text-text-secondary">{title}</p>
            <div className="mt-1 flex items-baseline gap-2">
              <span className="text-2xl font-bold tabular-nums text-text-primary">
                {value}
              </span>
              {changePercent != null && (
                <span
                  className={cn(
                    "text-xs font-medium tabular-nums",
                    trend === "up" && "text-[var(--color-success)]",
                    trend === "down" && "text-[var(--color-error)]",
                    trend === "neutral" && "text-text-muted"
                  )}
                >
                  {changePercent >= 0 ? "+" : ""}
                  {changePercent}%
                </span>
              )}
            </div>
            {changeLabel && (
              <p className="mt-0.5 text-xs text-text-muted">{changeLabel}</p>
            )}
            {sparklineData && sparklineData.length > 0 && (
              <div className="mt-3 h-8 w-full max-w-[140px]">
                <TrendSparkline
                  data={sparklineData}
                  trend={trend}
                  className="h-full w-full"
                />
              </div>
            )}
          </div>
          {icon && (
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-bg-hover text-text-secondary">
              {icon}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
