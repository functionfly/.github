/**
 * @functionfly/ui-graph
 * RuntimeGraph - runtime execution wrapper around FunctionCanvas
 */

import * as React from "react";
import { cn } from "./utils";
import {
  RuntimeGraphProps,
  RuntimeStatus,
  RuntimeMetrics,
  CanvasViewport,
  NodeData,
  EdgeData,
} from "./types";

const STATUS_CONFIG: Record<RuntimeStatus, { label: string; color: string; bgColor: string }> = {
  idle: { label: "Idle", color: "#6b7280", bgColor: "rgba(107,114,128,0.1)" },
  starting: { label: "Starting", color: "#3b82f6", bgColor: "rgba(59,130,246,0.1)" },
  running: { label: "Running", color: "#10b981", bgColor: "rgba(16,185,129,0.1)" },
  paused: { label: "Paused", color: "#f59e0b", bgColor: "rgba(245,158,11,0.1)" },
  completed: { label: "Completed", color: "#22c55e", bgColor: "rgba(34,197,94,0.1)" },
  failed: { label: "Failed", color: "#ef4444", bgColor: "rgba(239,68,68,0.1)" },
  cancelled: { label: "Cancelled", color: "#a855f7", bgColor: "rgba(168,85,247,0.1)" },
};

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  if (ms < 3600000) return `${(ms / 60000).toFixed(1)}m`;
  return `${(ms / 3600000).toFixed(1)}h`;
}

function formatNumber(num: number): string {
  if (num >= 1_000_000) return `${(num / 1_000_000).toFixed(1)}M`;
  if (num >= 1_000) return `${(num / 1_000).toFixed(1)}K`;
  return num.toString();
}

export function RuntimeGraph({
  nodes,
  edges,
  runtimeStatus = "idle",
  metrics,
  onNodeSelect,
  onNodeDoubleClick,
  onEdgeSelect,
  onCanvasClick,
  onStart,
  onPause,
  onStop,
  onReset,
  className,
}: RuntimeGraphProps) {
  const [viewport, setViewport] = React.useState<CanvasViewport>({ x: 0, y: 0, zoom: 1 });
  const statusConfig = STATUS_CONFIG[runtimeStatus];
  const isRunning = runtimeStatus === "running" || runtimeStatus === "starting";
  const isPaused = runtimeStatus === "paused";
  const isFinished = ["completed", "failed", "cancelled"].includes(runtimeStatus);

  return (
    <div className={cn("flex flex-col h-full bg-[#08080f] rounded-xl overflow-hidden", className)}>
      {/* Runtime Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-[#0d0d14] border-b border-[rgba(255,255,255,0.08)]">
        <div className="flex items-center gap-4">
          {/* Status indicator */}
          <div className="flex items-center gap-2">
            <div
              className={cn("size-2.5 rounded-full", isRunning && "animate-pulse")}
              style={{
                backgroundColor: statusConfig.color,
                boxShadow: isRunning ? `0 0 8px ${statusConfig.color}` : "none",
              }}
            />
            <span className="text-sm font-medium text-[#e8e8f0]">{statusConfig.label}</span>
          </div>

          {/* Metrics */}
          {metrics && (
            <div className="flex items-center gap-4 text-xs text-[#6b7280]">
              <div className="flex items-center gap-1.5">
                <span className="text-[#e8e8f0] font-mono font-medium">
                  {metrics.totalExecutions.toLocaleString()}
                </span>
                <span>exec</span>
              </div>
              <div className="flex items-center gap-1.5">
                <span
                  className={cn(
                    "font-mono font-medium",
                    metrics.failedExecutions > 0 ? "text-[#ef4444]" : "text-[#e8e8f0]"
                  )}
                >
                  {metrics.successfulExecutions}
                </span>
                <span>/</span>
                <span className="text-[#e8e8f0] font-mono font-medium">{metrics.totalExecutions}</span>
                <span>ok</span>
              </div>
              <div className="flex items-center gap-1.5">
                <span className="text-[#e8e8f0] font-mono font-medium">
                  {formatDuration(metrics.averageLatency)}
                </span>
                <span>avg</span>
              </div>
              {metrics.tokensProcessed > 0 && (
                <div className="flex items-center gap-1.5">
                  <span className="text-[#e8e8f0] font-mono font-medium">
                    {formatNumber(metrics.tokensProcessed)}
                  </span>
                  <span>tokens</span>
                </div>
              )}
              {metrics.totalCost > 0 && (
                <div className="flex items-center gap-1.5">
                  <span className="text-[#e8e8f0] font-mono font-medium">
                    ${metrics.totalCost.toFixed(4)}
                  </span>
                  <span>cost</span>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Controls */}
        <div className="flex items-center gap-2">
          {runtimeStatus === "idle" && (
            <button
              onClick={onStart}
              className={cn(
                "flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold",
                "bg-[#10b981] hover:bg-[#0ea472] text-white transition-colors",
                "shadow-[0_0_12px_rgba(16,185,129,0.3)]"
              )}
            >
              <span className="text-sm">▶</span> Start Execution
            </button>
          )}
          {isRunning && (
            <button
              onClick={onPause}
              className={cn(
                "flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold",
                "bg-[#f59e0b] hover:bg-[#d97706] text-white transition-colors",
                "shadow-[0_0_12px_rgba(245,158,11,0.3)]"
              )}
            >
              <span className="text-sm">⏸</span> Pause
            </button>
          )}
          {isPaused && (
            <>
              <button
                onClick={onStart}
                className={cn(
                  "flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold",
                  "bg-[#10b981] hover:bg-[#0ea472] text-white transition-colors"
                )}
              >
                <span className="text-sm">▶</span> Resume
              </button>
              <button
                onClick={onStop}
                className={cn(
                  "flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold",
                  "bg-[#ef4444] hover:bg-[#dc2626] text-white transition-colors"
                )}
              >
                <span className="text-sm">⏹</span> Stop
              </button>
            </>
          )}
          {isFinished && (
            <button
              onClick={onReset}
              className={cn(
                "flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold",
                "bg-[#3b82f6] hover:bg-[#2563eb] text-white transition-colors"
              )}
            >
              <span className="text-sm">↻</span> Reset
            </button>
          )}
        </div>
      </div>

      {/* Canvas Area - placeholder rendered when no children provided */}
      <div className="flex-1 relative overflow-hidden bg-[#08080f]">
        {/* Grid pattern */}
        <div
          className="absolute inset-0 pointer-events-none opacity-30"
          style={{
            backgroundImage: `
              linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px),
              linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px)
            `,
            backgroundSize: "64px 64px",
          }}
        />
        {nodes.length === 0 && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="text-center text-[#6b7280]">
              <div className="text-4xl mb-2 opacity-30">⚡</div>
              <div className="text-sm">No nodes to execute</div>
              <div className="text-xs mt-1 opacity-60">Add nodes to the canvas to begin</div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
