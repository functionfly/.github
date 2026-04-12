import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import {
  ArrowDownRight,
  ArrowUpRight,
  Clock,
  Flame,
  Medal,
  TrendingDown,
  TrendingUp,
  Zap,
} from "lucide-react";

export interface FunctionPerformance {
  id: string;
  name: string;
  avgLatency: number;
  p95Latency: number;
  successRate: number;
  invocations: number;
  trend: "up" | "down" | "stable";
  trendValue?: number;
}

export interface PerformanceLeaderboardProps {
  functions: FunctionPerformance[];
  className?: string;
  maxItems?: number;
  sortBy?: "latency" | "success" | "usage";
}

function PerformanceRow({
  fn,
  index,
  rank,
  highlight,
}: {
  fn: FunctionPerformance;
  index: number;
  rank: number;
  highlight: "good" | "bad" | "neutral";
}) {
  const rankStyles = {
    1: "text-yellow-500",
    2: "text-slate-400",
    3: "text-amber-600",
  };

  const highlightStyles = {
    good: "border-l-2 border-l-[var(--color-success)]",
    bad: "border-l-2 border-l-[var(--color-error)]",
    neutral: "",
  };

  const TrendIcon = fn.trend === "up" ? TrendingUp : TrendingDown;
  const trendColor =
    fn.trend === "up"
      ? highlight === "good"
        ? "text-[var(--color-success)]"
        : "text-[var(--color-error)]"
      : highlight === "good"
        ? "text-[var(--color-error)]"
        : "text-[var(--color-success)]";

  return (
    <motion.div
      initial={{ opacity: 0, x: -10 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.05 }}
      className={cn(
        "flex items-center gap-3 p-3 rounded-lg bg-bg-secondary/50 hover:bg-bg-tertiary transition-colors",
        highlightStyles[highlight]
      )}
    >
      <div
        className={cn(
          "flex items-center justify-center w-6 h-6 rounded-md font-bold text-sm shrink-0",
          rankStyles[rank as keyof typeof rankStyles] || "text-text-muted"
        )}
      >
        {rank <= 3 ? (
          <Medal className="w-5 h-5" />
        ) : (
          <span className="text-xs">{rank}</span>
        )}
      </div>

      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-text-primary truncate">
          {fn.name}
        </p>
        <div className="flex items-center gap-3 mt-0.5">
          <div className="flex items-center gap-1 text-xs text-text-muted">
            <Clock className="w-3 h-3" />
            <span className="tabular-nums">{Math.round(fn.avgLatency)}ms</span>
          </div>
          <div className="flex items-center gap-1 text-xs text-text-muted">
            <Zap className="w-3 h-3" />
            <span className="tabular-nums">
              {fn.successRate.toFixed(1)}% success
            </span>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <div className="text-right">
          <p className="text-xs text-text-muted tabular-nums">
            {fn.invocations.toLocaleString()}
          </p>
          <p className="text-[10px] text-text-dim">calls</p>
        </div>
        <TrendIcon className={cn("w-4 h-4", trendColor)} />
      </div>
    </motion.div>
  );
}

export function PerformanceLeaderboard({
  functions,
  className,
  maxItems = 5,
  sortBy = "latency",
}: PerformanceLeaderboardProps) {
  const sortedFunctions = [...functions].sort((a, b) => {
    switch (sortBy) {
      case "latency":
        return a.avgLatency - b.avgLatency;
      case "success":
        return b.successRate - a.successRate;
      case "usage":
        return b.invocations - a.invocations;
      default:
        return 0;
    }
  });

  const topFunctions = sortedFunctions.slice(0, maxItems);
  const bottomFunctions = sortedFunctions.slice(-maxItems).reverse();

  const hasData = functions.length > 0;

  return (
    <Card className={cn("border-theme bg-card overflow-hidden", className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium text-text-secondary">
            Performance Leaderboard
          </CardTitle>
          <div className="flex items-center gap-1 text-xs text-text-muted">
            <Flame className="w-3.5 h-3.5" />
            <span>Top & Bottom {maxItems}</span>
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        {!hasData ? (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <div className="w-12 h-12 rounded-full bg-bg-tertiary flex items-center justify-center mb-3">
              <Zap className="w-6 h-6 text-text-muted" />
            </div>
            <p className="text-sm text-text-secondary">
              No performance data yet
            </p>
            <p className="text-xs text-text-muted mt-1">
              Function metrics will appear here
            </p>
          </div>
        ) : (
          <div className="space-y-6">
            {/* Top Performers */}
            <div>
              <div className="flex items-center gap-2 mb-3">
                <ArrowUpRight className="w-4 h-4 text-(--color-success)" />
                <span className="text-xs font-medium text-text-secondary uppercase tracking-wide">
                  Top Performers
                </span>
              </div>
              <div className="space-y-1">
                {topFunctions.map((fn, index) => (
                  <PerformanceRow
                    key={fn.id}
                    fn={fn}
                    index={index}
                    rank={index + 1}
                    highlight="good"
                  />
                ))}
              </div>
            </div>

            {/* Bottom Performers */}
            {functions.length > maxItems && (
              <div>
                <div className="flex items-center gap-2 mb-3">
                  <ArrowDownRight className="w-4 h-4 text-(--color-error)" />
                  <span className="text-xs font-medium text-text-secondary uppercase tracking-wide">
                    Needs Attention
                  </span>
                </div>
                <div className="space-y-1">
                  {bottomFunctions.map((fn, index) => (
                    <PerformanceRow
                      key={fn.id}
                      fn={fn}
                      index={index}
                      rank={functions.length - bottomFunctions.length + index + 1}
                      highlight="bad"
                    />
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
