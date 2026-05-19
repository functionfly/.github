/**
 * @functionfly/ui-graph
 * LiveNodeTelemetry - real-time node performance metrics
 */

import * as React from "react";
import { cn } from "./utils";
import type { LiveNodeTelemetryProps, TelemetryMetric } from "./types";

const TREND_ICONS = {
  up: "↑",
  down: "↓",
  stable: "→",
};

const TREND_COLORS = {
  up: "#10b981",
  down: "#ef4444",
  stable: "#6b7280",
};

function formatValue(value: number | string, unit?: string): string {
  if (typeof value === "string") return value;
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(1)}K`;
  return `${value}${unit || ""}`;
}

export function LiveNodeTelemetry({
  nodeId,
  nodeLabel,
  metrics,
  isExpanded = false,
  refreshInterval = 1000,
  onToggleExpand,
  className,
}: LiveNodeTelemetryProps) {
  const [localMetrics, setLocalMetrics] = React.useState(metrics);

  // Simulated live updates
  React.useEffect(() => {
    if (!isExpanded || refreshInterval <= 0) return;

    const interval = setInterval(() => {
      setLocalMetrics((prev) =>
        prev.map((m) => {
          if (typeof m.value === "number") {
            const delta = (Math.random() - 0.5) * 0.1 * (m.value as number);
            return { ...m, value: Math.max(0, (m.value as number) + delta) };
          }
          return m;
        })
      );
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [isExpanded, refreshInterval]);

  return (
    <div
      className={cn(
        "flex flex-col rounded-xl border transition-all",
        isExpanded
          ? "bg-[#0d0d14] border-[rgba(249,115,22,0.3)] shadow-[0_0_20px_rgba(249,115,22,0.1)]"
          : "bg-[#14141f] border-[rgba(255,255,255,0.06)]",
        className
      )}
    >
      {/* Header */}
      <button
        onClick={onToggleExpand}
        className="flex items-center justify-between px-3 py-2.5 w-full text-left"
      >
        <div className="flex items-center gap-2">
          <div className={cn("size-2 rounded-full", isExpanded ? "animate-pulse bg-[#10b981]" : "bg-[#6b7280]")} />
          <div>
            <div className="text-xs text-[#e8e8f0] font-medium truncate max-w-[120px]">{nodeLabel}</div>
            <div className="text-[10px] text-[#6b7280] font-mono">{nodeId.slice(0, 8)}</div>
          </div>
        </div>
        <span className={`text-[#6b7280] transition-transform ${isExpanded ? "rotate-180" : ""}`}>
          ▼
        </span>
      </button>

      {/* Metrics */}
      {isExpanded && (
        <div className="px-3 pb-3 space-y-2">
          {localMetrics.map((metric, index) => (
            <div
              key={index}
              className="flex items-center justify-between px-2 py-1.5 rounded-lg bg-[#14141f]"
            >
              <span className="text-[10px] text-[#6b7280]">{metric.label}</span>
              <div className="flex items-center gap-2">
                <span className="text-xs text-[#e8e8f0] font-mono">
                  {formatValue(metric.value, metric.unit)}
                </span>
                {metric.trend && (
                  <span
                    className="text-[10px]"
                    style={{ color: TREND_COLORS[metric.trend] }}
                  >
                    {TREND_ICONS[metric.trend]}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
