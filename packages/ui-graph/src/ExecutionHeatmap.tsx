/**
 * @functionfly/ui-graph
 * ExecutionHeatmap - heatmap visualization of execution activity
 */

import * as React from "react";
import { cn } from "./utils";
import type { ExecutionHeatmapProps, HeatmapCell } from "./types";

const COLOR_SCALES = {
  hot: ["#0d0d14", "#1e3a5f", "#2563eb", "#f97316", "#ef4444", "#fef08a"],
  cool: ["#0d0d14", "#0f172a", "#1e40af", "#3b82f6", "#06b6d4", "#a5f3fc"],
  viridis: ["#0d0d14", "#2c1654", "#3b4cc0", "#22b74b", "#f9a825", "#fef08a"],
  plasma: ["#0d0d14", "#3b0774", "#9f23af", "#f47b0a", "#fef08a", "#ffffff"],
};

function getColor(intensity: number, colors: string[]): string {
  const clampedIntensity = Math.max(0, Math.min(1, intensity));
  const index = Math.floor(clampedIntensity * (colors.length - 1));
  return colors[Math.min(index, colors.length - 1)];
}

export function ExecutionHeatmap({
  cells,
  timeRange,
  columns = 24,
  cellSize = 12,
  colorScale = "hot",
  onCellClick,
  className,
}: ExecutionHeatmapProps) {
  const colors = COLOR_SCALES[colorScale];
  const totalDuration = timeRange.end - timeRange.start;
  const rows = Math.ceil(cells.length / columns);

  const getCellAtTime = (timestamp: number): HeatmapCell | undefined => {
    return cells.find(
      (c) => Math.abs(c.timestamp - timestamp) < totalDuration / (columns * rows)
    );
  };

  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <span className="text-xs text-[#6b7280]">Execution Heatmap</span>
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-[#6b7280]">Low</span>
          <div className="flex items-center gap-0.5">
            {colors.slice(1, -1).map((color, i) => (
              <div
                key={i}
                className="size-2 rounded-sm"
                style={{ backgroundColor: color }}
              />
            ))}
          </div>
          <span className="text-[10px] text-[#6b7280]">High</span>
        </div>
      </div>

      {/* Heatmap grid */}
      <div
        className="grid gap-0.5"
        style={{
          gridTemplateColumns: `repeat(${columns}, ${cellSize}px)`,
          gridTemplateRows: `repeat(${rows}, ${cellSize}px)`,
        }}
      >
        {Array.from({ length: columns * rows }).map((_, index) => {
          const timestamp = timeRange.start + (index / (columns * rows)) * totalDuration;
          const cell = getCellAtTime(timestamp);
          const intensity = cell?.intensity ?? 0;
          const bgColor = getColor(intensity, colors);

          return (
            <div
              key={index}
              className={cn(
                "rounded-sm cursor-pointer transition-all",
                cell?.status === "error" && "ring-1 ring-[#ef4444]/30",
                "hover:ring-1 hover:ring-white/30 hover:scale-110"
              )}
              style={{
                backgroundColor: bgColor,
                width: cellSize,
                height: cellSize,
              }}
              onClick={() => cell && onCellClick?.(cell)}
              title={
                cell
                  ? `${cell.nodeId}: ${Math.round(intensity * 100)}% intensity`
                  : "No execution"
              }
            />
          );
        })}
      </div>

      {/* Time range labels */}
      <div className="flex items-center justify-between text-[10px] text-[#6b7280] font-mono">
        <span>{new Date(timeRange.start).toLocaleTimeString()}</span>
        <span>{Math.round(totalDuration / 1000)}s</span>
        <span>{new Date(timeRange.end).toLocaleTimeString()}</span>
      </div>
    </div>
  );
}
