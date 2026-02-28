import { cn } from "@/lib/utils";

export type SystemHealthStatus = "healthy" | "degraded" | "down" | "unknown";

export interface SystemHealthIndicatorProps {
  status: SystemHealthStatus;
  label?: string;
  size?: "sm" | "md" | "lg";
  showLabel?: boolean;
  className?: string;
}

const statusConfig: Record<
  SystemHealthStatus,
  { label: string; dotClass: string; textClass: string; pulse?: boolean }
> = {
  healthy: {
    label: "Healthy",
    dotClass: "bg-[var(--color-status-online)]",
    textClass: "text-[var(--color-success)]",
  },
  degraded: {
    label: "Degraded",
    dotClass: "bg-[var(--color-status-degraded)]",
    textClass: "text-[var(--color-warning)]",
    pulse: true,
  },
  down: {
    label: "Down",
    dotClass: "bg-[var(--color-status-offline)]",
    textClass: "text-[var(--color-error)]",
    pulse: true,
  },
  unknown: {
    label: "Unknown",
    dotClass: "bg-[var(--color-status-pending)]",
    textClass: "text-text-muted",
  },
};

const sizeConfig = {
  sm: { dot: "h-2 w-2", text: "text-xs" },
  md: { dot: "h-2.5 w-2.5", text: "text-sm" },
  lg: { dot: "h-3 w-3", text: "text-base" },
};

export function SystemHealthIndicator({
  status,
  label,
  size = "md",
  showLabel = true,
  className,
}: SystemHealthIndicatorProps) {
  const config = statusConfig[status];
  const sizeStyles = sizeConfig[size];
  const displayLabel = label ?? config.label;

  return (
    <span
      className={cn(
        "inline-flex items-center gap-2 font-medium",
        config.textClass,
        className
      )}
    >
      <span
        className={cn(
          "shrink-0 rounded-full",
          sizeStyles.dot,
          config.dotClass,
          config.pulse && "animate-pulse"
        )}
        title={config.label}
      />
      {showLabel && (
        <span className={cn(sizeStyles.text, "text-text-secondary")}>
          {displayLabel}
        </span>
      )}
    </span>
  );
}
