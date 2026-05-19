/**
 * @functionfly/ui-graph
 * LoopVisualizer - visualize loop iterations
 */

import * as React from "react";
import { cn } from "./utils";
import type { LoopVisualizerProps, LoopIteration } from "./types";

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${(ms / 60000).toFixed(1)}m`;
}

const STATUS_COLORS = {
  running: "#3b82f6",
  completed: "#10b981",
  failed: "#ef4444",
  skipped: "#6b7280",
};

export function LoopVisualizer({
  loopNodeId,
  iterations,
  maxIterations = 10,
  isExpanded = false,
  currentIteration,
  onIterationClick,
  onToggleExpand,
  className,
}: LoopVisualizerProps) {
  const completedIterations = iterations.filter((i) => i.status === "completed").length;
  const failedIterations = iterations.filter((i) => i.status === "failed").length;
  const progress = maxIterations > 0 ? (completedIterations / maxIterations) * 100 : 0;

  return (
    <div className={cn("flex flex-col rounded-xl border transition-all", isExpanded ? "bg-[#0d0d14] border-[rgba(249,115,22,0.3)]" : "bg-[#14141f] border-[rgba(255,255,255,0.06)]", className)}>
      {/* Header */}
      <button onClick={onToggleExpand} className="flex items-center justify-between px-3 py-2.5 w-full text-left">
        <div className="flex items-center gap-3">
          <div className={cn("size-2 rounded-full", currentIteration !== undefined ? "animate-pulse bg-[#3b82f6]" : "bg-[#6b7280]")} />
          <div>
            <div className="text-xs text-[#e8e8f0]">Loop</div>
            <div className="text-[10px] text-[#6b7280] font-mono">{loopNodeId.slice(0, 8)}</div>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-[10px] text-[#6b7280]">{completedIterations}/{maxIterations}</span>
          <span className={`text-[#6b7280] transition-transform ${isExpanded ? "rotate-180" : ""}`}>▼</span>
        </div>
      </button>

      {/* Progress bar */}
      <div className="px-3 pb-2">
        <div className="h-1.5 bg-[#1a1a28] rounded-full overflow-hidden">
          <div
            className="h-full rounded-full transition-all bg-gradient-to-r from-[#f97316] to-[#fbbf24]"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* Expanded iterations list */}
      {isExpanded && (
        <div className="px-3 pb-3 space-y-1 max-h-48 overflow-y-auto">
          {iterations.length === 0 ? (
            <div className="text-center py-4 text-xs text-[#6b7280]">No iterations yet</div>
          ) : (
            iterations.map((iteration) => (
              <IterationCard key={iteration.id} iteration={iteration} onClick={onIterationClick} />
            ))
          )}
        </div>
      )}
    </div>
  );
}

function IterationCard({
  iteration,
  onClick,
}: {
  iteration: LoopIteration;
  onClick?: LoopVisualizerProps["onIterationClick"];
}) {
  const statusColor = STATUS_COLORS[iteration.status || "skipped"];

  return (
    <div
      onClick={() => onClick?.(iteration)}
      className={cn(
        "flex items-center justify-between px-2 py-1.5 rounded cursor-pointer transition-colors",
        "hover:bg-[rgba(255,255,255,0.04)]",
        iteration.status === "running" && "bg-[rgba(59,130,246,0.1)]",
        iteration.status === "failed" && "bg-[rgba(239,68,68,0.1)]"
      )}
    >
      <div className="flex items-center gap-2">
        <div className="size-5 rounded flex items-center justify-center text-[10px] font-mono" style={{ backgroundColor: `${statusColor}20`, color: statusColor }}>
          {iteration.index + 1}
        </div>
        <span className="text-xs text-[#e8e8f0]">{iteration.nodeId.slice(0, 8)}</span>
      </div>
      <div className="flex items-center gap-2">
        {iteration.duration != null && (
          <span className="text-[10px] text-[#6b7280] font-mono">{formatDuration(iteration.duration)}</span>
        )}
        <div className="size-2 rounded-full" style={{ backgroundColor: statusColor }} />
      </div>
    </div>
  );
}
