import { cn } from "@/lib/utils";

export interface DeterministicReliabilityBadgeProps {
  /** Reliability percentage (e.g., 99.9987) */
  percentage: number;
  /** URL to view history */
  historyUrl?: string;
  /** Show trend indicator */
  showTrend?: boolean;
  /** Trend direction */
  trend?: "up" | "down" | "stable";
  /** Custom className */
  className?: string;
}

export function DeterministicReliabilityBadge({
  percentage,
  historyUrl,
  showTrend = false,
  trend = "stable",
  className,
}: DeterministicReliabilityBadgeProps) {
  const formattedPercentage = percentage.toFixed(4);

  const getColor = () => {
    if (percentage >= 99.9) return "text-green-500";
    if (percentage >= 99) return "text-blue-500";
    if (percentage >= 95) return "text-yellow-500";
    return "text-red-500";
  };

  const getTrendIcon = () => {
    if (!showTrend) return null;
    
    if (trend === "up") return "↑";
    if (trend === "down") return "↓";
    return "→";
  };

  const content = (
    <div className={cn("flex items-center gap-2", className)}>
      <span className={cn("font-mono font-bold", getColor())}>
        {formattedPercentage}%
      </span>
      <span className="text-sm text-muted-foreground">
        Deterministic Reliability
      </span>
      {showTrend && (
        <span className={cn(
          "text-xs font-medium",
          trend === "up" && "text-green-500",
          trend === "down" && "text-red-500",
          trend === "stable" && "text-muted-foreground"
        )}>
          {getTrendIcon()}
        </span>
      )}
    </div>
  );

  if (historyUrl) {
    return (
      <a
        href={historyUrl}
        className="hover:underline cursor-pointer"
      >
        {content}
      </a>
    );
  }

  return content;
}
