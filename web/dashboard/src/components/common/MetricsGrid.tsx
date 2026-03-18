import { ReactNode } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { Sparkline } from "./Sparkline";
import { TrendingUp, TrendingDown, Minus } from "lucide-react";

interface MetricItem {
  id: string;
  title: string;
  value: string | number;
  previousValue?: string | number;
  change?: { value: number; label: string };
  icon?: ReactNode;
  sparklineData?: number[];
  trend?: "up" | "down" | "neutral";
  color?: string;
  className?: string;
}

interface MetricsGridProps {
  metrics: MetricItem[];
  columns?: number;
  className?: string;
  showSparklines?: boolean;
  showTrends?: boolean;
}

const TrendIcon = ({ trend, className }: { trend?: string; className?: string }) => {
  switch (trend) {
    case "up":
      return <TrendingUp className={cn("w-4 h-4 text-green-400", className)} />;
    case "down":
      return <TrendingDown className={cn("w-4 h-4 text-red-400", className)} />;
    default:
      return <Minus className={cn("w-4 h-4 text-gray-400", className)} />;
  }
};

const calculateChange = (current: string | number, previous?: string | number): { value: number; trend: "up" | "down" | "neutral" } | null => {
  if (!previous) return null;

  const currentNum = typeof current === "string" ? parseFloat(current.replace(/[^\d.-]/g, "")) : current;
  const previousNum = typeof previous === "string" ? parseFloat(previous.replace(/[^\d.-]/g, "")) : previous;

  if (isNaN(currentNum) || isNaN(previousNum)) return null;

  const change = ((currentNum - previousNum) / previousNum) * 100;
  const trend = change > 0 ? "up" : change < 0 ? "down" : "neutral";

  return { value: Math.abs(change), trend };
};

export function MetricsGrid({
  metrics,
  columns = 2,
  className,
  showSparklines = true,
  showTrends = true,
}: MetricsGridProps) {
  return (
    <div
      className={cn(
        "grid gap-4",
        {
          "grid-cols-1": columns === 1,
          "grid-cols-2": columns === 2,
          "grid-cols-3": columns === 3,
          "grid-cols-4": columns === 4,
          "md:grid-cols-2 lg:grid-cols-3": columns === 2,
          "md:grid-cols-3 lg:grid-cols-4": columns === 3,
          "md:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5": columns === 4,
        },
        className
      )}
    >
      {metrics.map((metric) => {
        const calculatedChange = metric.change || (showTrends && metric.previousValue
          ? calculateChange(metric.value, metric.previousValue)
          : null);

        const trend =
          (calculatedChange && "trend" in calculatedChange ? calculatedChange.trend : undefined) ??
          metric.trend ??
          "neutral";

        return (
          <Card
            key={metric.id}
            className={cn(
              "metric-card overflow-hidden transition-all duration-200 hover:shadow-lg",
              metric.className
            )}
          >
            <CardContent className="p-6">
              <div className="flex items-start justify-between mb-4">
                <div className="space-y-2 flex-1">
                  <p className="text-sm font-medium text-text-secondary">
                    {metric.title}
                  </p>
                  <div className="flex items-baseline gap-2">
                    <span
                      className="text-3xl font-bold text-white"
                      style={{ color: metric.color }}
                    >
                      {metric.value}
                    </span>
                    {calculatedChange && showTrends && (
                      <div className="flex items-center gap-1">
                        <TrendIcon trend={trend} />
                        <Badge
                          variant="secondary"
                          className={cn(
                            "text-xs font-medium",
                            trend === "up" && "bg-green-400/10 text-green-400 border-green-400/20",
                            trend === "down" && "bg-red-400/10 text-red-400 border-red-400/20",
                            trend === "neutral" && "bg-gray-400/10 text-gray-400 border-gray-400/20"
                          )}
                        >
                          {calculatedChange.value > 0 ? "+" : ""}{calculatedChange.value.toFixed(1)}%
                        </Badge>
                      </div>
                    )}
                  </div>
                  {calculatedChange && "label" in calculatedChange && calculatedChange.label && (
                    <p className="text-xs text-text-muted">
                      {calculatedChange.label}
                    </p>
                  )}
                </div>

                {metric.icon && (
                  <div className="p-3 rounded-xl bg-linear-to-br from-indigo-500/10 to-purple-500/10 border border-indigo-500/20">
                    <div className="text-indigo-400">
                      {metric.icon}
                    </div>
                  </div>
                )}
              </div>

              {showSparklines && metric.sparklineData && metric.sparklineData.length > 0 && (
                <div className="mt-4">
                  <Sparkline
                    data={metric.sparklineData}
                    height={32}
                    color={metric.color || "#10b981"}
                    className="opacity-80"
                  />
                </div>
              )}
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}