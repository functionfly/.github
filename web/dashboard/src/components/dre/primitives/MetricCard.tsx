import { cn } from "@/lib/utils";

export interface MetricCardProps {
  /** Metric label */
  label: string;
  /** Metric value */
  value: string | number;
  /** Optional description */
  description?: string;
  /** Optional trend indicator */
  trend?: {
    value: number;
    direction: "up" | "down";
  };
  /** Optional icon */
  icon?: React.ReactNode;
  /** Custom value className */
  valueClassName?: string;
  /** Custom className */
  className?: string;
  /** Click handler */
  onClick?: () => void;
}

export function MetricCard({
  label,
  value,
  description,
  trend,
  icon,
  valueClassName,
  className,
  onClick,
}: MetricCardProps) {
  return (
    <div
      className={cn(
        "bg-bg-secondary/50 border border-border-subtle rounded-lg p-4",
        onClick && "cursor-pointer hover:bg-bg-secondary transition-colors",
        className
      )}
      onClick={onClick}
    >
      <div className="flex items-start justify-between mb-2">
        <span className="text-sm font-medium text-muted-foreground">{label}</span>
        {icon && <span className="text-muted-foreground">{icon}</span>}
      </div>
      <div className={cn("text-2xl font-bold", valueClassName)}>{value}</div>
      {(description || trend) && (
        <div className="mt-2 flex items-center gap-2">
          {trend && (
            <span
              className={cn(
                "text-xs font-medium",
                trend.direction === "up" ? "text-green-500" : "text-red-500"
              )}
            >
              {trend.direction === "up" ? "↑" : "↓"} {Math.abs(trend.value)}%
            </span>
          )}
          {description && (
            <span className="text-xs text-muted-foreground">{description}</span>
          )}
        </div>
      )}
    </div>
  );
}
