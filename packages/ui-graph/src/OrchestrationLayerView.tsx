/**
 * @functionfly/ui-graph
 * OrchestrationLayerView - visualize orchestration layers
 */

import * as React from "react";
import { cn } from "./utils";
import type { OrchestrationLayerViewProps, OrchestrationLayer, NodeType } from "./types";

const LAYER_COLORS = [
  "#f97316", // orange
  "#3b82f6", // blue
  "#10b981", // green
  "#8b5cf6", // purple
  "#ec4899", // pink
  "#06b6d4", // cyan
];

const NODE_ICONS: Record<NodeType, string> = {
  function: "⚡",
  agent: "🤖",
  api: "🔌",
  memory: "🧠",
  database: "🗄️",
  robot: "🦾",
  browser: "🌐",
  gpu: "🖥️",
  workflow: "🔀",
  trigger: "🎯",
  condition: "🔀",
  output: "📤",
};

export function OrchestrationLayerView({
  layers,
  onLayerClick,
  onLayerExpand,
  className,
}: OrchestrationLayerViewProps) {
  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      <div className="flex items-center justify-between">
        <span className="text-xs text-[#6b7280]">Orchestration Layers</span>
        <span className="text-[10px] text-[#f97316] font-mono">{layers.length} layers</span>
      </div>

      {layers.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-center">
          <div className="text-2xl opacity-30 mb-1">📊</div>
          <p className="text-xs text-[#6b7280]">No layers defined</p>
          <p className="text-[10px] text-[#4b5563] mt-1">Orchestrated workflows will show their layers here</p>
        </div>
      ) : (
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {layers.map((layer, index) => {
            const color = LAYER_COLORS[index % LAYER_COLORS.length];
            return (
              <LayerCard
                key={layer.id}
                layer={layer}
                color={color}
                onClick={onLayerClick}
                onExpand={onLayerExpand}
              />
            );
          })}
        </div>
      )}
    </div>
  );
}

function LayerCard({
  layer,
  color,
  onClick,
  onExpand,
}: {
  layer: OrchestrationLayer;
  color: string;
  onClick?: OrchestrationLayerViewProps["onLayerClick"];
  onExpand?: OrchestrationLayerViewProps["onLayerExpand"];
}) {
  const hasError = layer.hasError;
  const nodeCount = layer.nodes.length;

  return (
    <div
      className={cn(
        "flex flex-col rounded-lg border transition-all cursor-pointer",
        hasError
          ? "border-[rgba(239,68,68,0.3)] bg-[rgba(239,68,68,0.05)]"
          : "border-[rgba(255,255,255,0.06)] hover:border-[rgba(249,115,22,0.2)]"
      )}
      onClick={() => onClick?.(layer)}
    >
      {/* Layer header */}
      <div className="flex items-center justify-between px-3 py-2">
        <div className="flex items-center gap-2">
          {/* Layer number badge */}
          <div
            className="size-6 rounded-md flex items-center justify-center text-xs font-bold text-white"
            style={{ backgroundColor: color }}
          >
            {layer.depth}
          </div>

          <div>
            <div className="flex items-center gap-2">
              <span className="text-sm text-[#e8e8f0]">{layer.label}</span>
              {layer.isParallel && (
                <span className="text-[9px] px-1.5 py-0.5 rounded bg-[rgba(168,85,247,0.2)] text-[#a855f7]">
                  parallel
                </span>
              )}
              {hasError && <span className="text-[#ef4444]">⚠️</span>}
            </div>
            {layer.description && (
              <p className="text-[10px] text-[#6b7280] mt-0.5">{layer.description}</p>
            )}
          </div>
        </div>

        {/* Node count */}
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-[#6b7280]">{nodeCount} nodes</span>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onExpand?.(layer.id);
            }}
            className="size-5 flex items-center justify-center rounded text-[#6b7280] hover:text-[#e8e8f0] hover:bg-[rgba(255,255,255,0.06)]"
          >
            ▼
          </button>
        </div>
      </div>

      {/* Node list */}
      <div className="flex flex-wrap gap-1 px-3 pb-2">
        {layer.nodes.map((node) => (
          <div
            key={node.id}
            className={cn(
              "flex items-center gap-1 px-2 py-1 rounded text-[10px]",
              "bg-[#14141f] border border-[rgba(255,255,255,0.06)]"
            )}
          >
            <span>{NODE_ICONS[node.type] || "📦"}</span>
            <span className="text-[#e8e8f0]">{node.label}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
