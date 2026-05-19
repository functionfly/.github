/**
 * @functionfly/ui-graph
 * DistributedRuntimeMap - visualize distributed runtime regions
 */

import * as React from "react";
import { cn } from "./utils";
import type { DistributedRuntimeMapProps, RuntimeRegion, NodeType } from "./types";

const STATUS_CONFIG = {
  healthy: { color: "#10b981", label: "Healthy", bg: "rgba(16,185,129,0.1)" },
  degraded: { color: "#f59e0b", label: "Degraded", bg: "rgba(245,158,11,0.1)" },
  unhealthy: { color: "#ef4444", label: "Unhealthy", bg: "rgba(239,68,68,0.1)" },
};

const NODE_COLORS: Record<NodeType, string> = {
  function: "#f97316",
  agent: "#8b5cf6",
  api: "#3b82f6",
  memory: "#10b981",
  database: "#ef4444",
  robot: "#f59e0b",
  browser: "#06b6d4",
  gpu: "#ec4899",
  workflow: "#a855f7",
  trigger: "#6366f1",
  condition: "#14b8a6",
  output: "#64748b",
};

export function DistributedRuntimeMap({
  regions,
  selectedRegionId,
  onRegionSelect,
  onRegionExpand,
  className,
}: DistributedRuntimeMapProps) {
  const totalNodes = regions.reduce((sum, r) => sum + r.nodeCount, 0);
  const healthyRegions = regions.filter((r) => r.status === "healthy").length;

  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-base">🌍</span>
          <span className="text-xs text-[#6b7280]">Distributed Runtime</span>
        </div>
        <div className="flex items-center gap-3 text-[10px]">
          <span className="text-[#6b7280]">{regions.length} regions</span>
          <span className="text-[#6b7280]">{totalNodes} nodes</span>
          <span className="text-[#10b981]">{healthyRegions} healthy</span>
        </div>
      </div>

      {/* Region map */}
      {regions.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-center">
          <div className="text-2xl opacity-30 mb-1">🌐</div>
          <p className="text-xs text-[#6b7280]">No runtime regions</p>
          <p className="text-[10px] text-[#4b5563] mt-1">Distributed runtimes will appear here</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-2">
          {regions.map((region) => (
            <RegionCard
              key={region.id}
              region={region}
              isSelected={region.id === selectedRegionId}
              onSelect={onRegionSelect}
              onExpand={onRegionExpand}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function RegionCard({
  region,
  isSelected,
  onSelect,
  onExpand,
}: {
  region: RuntimeRegion;
  isSelected: boolean;
  onSelect?: DistributedRuntimeMapProps["onRegionSelect"];
  onExpand?: DistributedRuntimeMapProps["onRegionExpand"];
}) {
  const statusConfig = STATUS_CONFIG[region.status || "healthy"];
  const healthPercent = region.nodeCount > 0 ? (region.healthyNodes / region.nodeCount) * 100 : 0;

  return (
    <div
      onClick={() => onSelect?.(region.id)}
      className={cn(
        "flex flex-col p-3 rounded-lg border cursor-pointer transition-all",
        isSelected
          ? "border-[rgba(249,115,22,0.4)] bg-[rgba(249,115,22,0.05)]"
          : "border-[rgba(255,255,255,0.06)] hover:border-[rgba(255,255,255,0.1)]"
      )}
    >
      {/* Region header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="text-lg">📍</span>
          <div>
            <div className="text-xs text-[#e8e8f0]">{region.label}</div>
            <div className="text-[10px] text-[#6b7280]">{region.location}</div>
          </div>
        </div>
        <div
          className="px-1.5 py-0.5 rounded text-[9px] font-medium"
          style={{ backgroundColor: statusConfig.bg, color: statusConfig.color }}
        >
          {statusConfig.label}
        </div>
      </div>

      {/* Health bar */}
      <div className="mb-2">
        <div className="flex items-center justify-between text-[10px] mb-1">
          <span className="text-[#6b7280]">Health</span>
          <span className="text-[#e8e8f0] font-mono">{healthPercent.toFixed(0)}%</span>
        </div>
        <div className="h-1.5 bg-[#1a1a28] rounded-full overflow-hidden">
          <div
            className="h-full rounded-full transition-all"
            style={{
              width: `${healthPercent}%`,
              backgroundColor: statusConfig.color,
            }}
          />
        </div>
      </div>

      {/* Node stats */}
      <div className="flex items-center gap-3 text-[10px] mb-2">
        <div className="flex items-center gap-1">
          <div className="size-2 rounded-full bg-[#10b981]" />
          <span className="text-[#e8e8f0] font-mono">{region.healthyNodes}</span>
          <span className="text-[#6b7280]">ok</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="size-2 rounded-full bg-[#ef4444]" />
          <span className="text-[#e8e8f0] font-mono">{region.unhealthyNodes}</span>
          <span className="text-[#6b7280]">err</span>
        </div>
        <div className="flex items-center gap-1">
          <span className="text-[#6b7280]">nodes:</span>
          <span className="text-[#e8e8f0] font-mono">{region.nodeCount}</span>
        </div>
      </div>

      {/* Latency */}
      {region.latency != null && (
        <div className="flex items-center gap-1 text-[10px]">
          <span className="text-[#6b7280]">Latency:</span>
          <span className="text-[#e8e8f0] font-mono">{region.latency}ms</span>
        </div>
      )}

      {/* Expand button */}
      <button
        onClick={(e) => {
          e.stopPropagation();
          onExpand?.(region.id);
        }}
        className="mt-2 w-full py-1 rounded text-[10px] bg-[#14141f] text-[#6b7280] hover:text-[#e8e8f0] hover:bg-[#1a1a28] transition-colors"
      >
        View nodes →
      </button>
    </div>
  );
}
