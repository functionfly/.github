import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export interface MemoryUsageGaugeProps {
  /** Current usage 0–100 */
  percent: number;
  /** Optional label, e.g. "Heap" or "RSS" */
  label?: string;
  /** Size of the gauge ring */
  size?: "sm" | "md" | "lg";
  /** Show warning state when above this threshold (0–100) */
  warningThreshold?: number;
  /** Show danger state when above this threshold (0–100) */
  dangerThreshold?: number;
  className?: string;
}

const sizeConfig = {
  sm: { width: 80, strokeWidth: 6, fontSize: "text-sm" },
  md: { width: 120, strokeWidth: 8, fontSize: "text-xl" },
  lg: { width: 160, strokeWidth: 10, fontSize: "text-2xl" },
};

export function MemoryUsageGauge({
  percent,
  label = "Memory",
  size = "md",
  warningThreshold = 75,
  dangerThreshold = 90,
  className,
}: MemoryUsageGaugeProps) {
  const [animatedPercent, setAnimatedPercent] = useState(0);
  const config = sizeConfig[size];
  const clamped = Math.max(0, Math.min(100, percent));

  useEffect(() => {
    const t = setTimeout(() => setAnimatedPercent(clamped), 50);
    return () => clearTimeout(t);
  }, [clamped]);

  const radius = (config.width - config.strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (animatedPercent / 100) * circumference;

  const statusColor =
    animatedPercent >= dangerThreshold
      ? "var(--color-error)"
      : animatedPercent >= warningThreshold
        ? "var(--color-warning)"
        : "var(--color-success)";

  return (
    <Card className={cn("border-theme bg-card", className)}>
      <CardHeader className="pb-1">
        <CardTitle className="text-sm font-medium text-text-secondary">
          {label}
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center pt-0">
        <div className="relative inline-flex items-center justify-center">
          <svg
            width={config.width}
            height={config.width}
            className="-rotate-90"
          >
            <circle
              cx={config.width / 2}
              cy={config.width / 2}
              r={radius}
              fill="none"
              stroke="var(--color-border-subtle)"
              strokeWidth={config.strokeWidth}
            />
            <circle
              cx={config.width / 2}
              cy={config.width / 2}
              r={radius}
              fill="none"
              stroke={statusColor}
              strokeWidth={config.strokeWidth}
              strokeLinecap="round"
              strokeDasharray={circumference}
              strokeDashoffset={strokeDashoffset}
              className="transition-all duration-700 ease-out"
            />
          </svg>
          <div className="absolute inset-0 flex flex-col items-center justify-center">
            <span
              className={cn("font-bold tabular-nums", config.fontSize)}
              style={{ color: statusColor }}
            >
              {Math.round(animatedPercent)}%
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
