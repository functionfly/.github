/**
 * @functionfly/ui-graph
 * FailurePropagationMap - visualize how failures propagate through the graph
 */

import * as React from "react";
import { cn } from "./utils";
import type { FailurePropagationMapProps, FailureNode, NodeType } from "./types";

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

export function FailurePropagationMap({
  rootFailure,
  propagation,
  onNodeClick,
  className,
}: FailurePropagationMapProps) {
  const rootNode = propagation.find((n) => n.isRootCause);

  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(239,68,68,0.2)]", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-[#ef4444] text-lg">⚠️</span>
          <span className="text-xs text-[#6b7280]">Failure Propagation</span>
        </div>
        <span className="text-[10px] text-[#ef4444] font-mono">{propagation.length} affected</span>
      </div>

      {/* Root failure */}
      <div className="px-3 py-2 rounded-lg bg-[rgba(239,68,68,0.1)] border border-[rgba(239,68,68,0.2)]">
        <div className="text-[10px] text-[#ef4444] uppercase tracking-wide mb-1">Root Cause</div>
        <div className="text-sm text-[#e8e8f0]">{rootFailure.label}</div>
        <div className="text-xs text-[#6b7280] mt-1">{rootFailure.reason}</div>
      </div>

      {/* Propagation tree */}
      {propagation.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-4 text-center">
          <p className="text-xs text-[#6b7280]">No propagation data</p>
        </div>
      ) : (
        <div className="space-y-1 max-h-64 overflow-y-auto">
          {propagation.map((node) => (
            <div
              key={node.id}
              onClick={() => onNodeClick?.(node.nodeId)}
              className={cn(
                "flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer transition-colors",
                "border",
                node.isRootCause
                  ? "border-[rgba(239,68,68,0.4)] bg-[rgba(239,68,68,0.05)]"
                  : "border-[rgba(255,255,255,0.06)] hover:border-[rgba(255,255,255,0.1)]"
              )}
              style={{ paddingLeft: `${node.depth * 16 + 12}px` }}
            >
              {/* Propagation arrow */}
              {node.depth > 0 && (
                <div
                  className="absolute left-0 h-px bg-[rgba(239,68,68,0.2)]"
                  style={{ left: `${(node.depth - 1) * 16 + 12}px`, width: "16px" }}
                />
              )}

              {/* Node indicator */}
              <div
                className="size-3 rounded-sm shrink-0"
                style={{ backgroundColor: NODE_COLORS[node.type] || "#6b7280" }}
              />

              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-[#e8e8f0] truncate">{node.label}</span>
                  {node.isRootCause && (
                    <span className="text-[9px] px-1 py-0.5 rounded bg-[rgba(239,68,68,0.2)] text-[#ef4444]">
                      root
                    </span>
                  )}
                </div>
                {node.failureReason && (
                  <p className="text-[10px] text-[#6b7280] truncate">{node.failureReason}</p>
                )}
              </div>

              {/* Affected count */}
              <span className="text-[10px] text-[#ef4444] font-mono shrink-0">
                {node.affectedNodes.length} affected
              </span>
            </div>
          ))}
        </div>
      )}

      {/* Propagation path */}
      {rootNode && (
        <div className="pt-2 border-t border-[rgba(255,255,255,0.06)]">
          <div className="text-[10px] text-[#6b7280] uppercase tracking-wide mb-1">Path</div>
          <div className="flex items-center gap-1 text-[10px] font-mono text-[#6b7280] flex-wrap">
            {rootNode.propagationPath.map((id, i) => (
              <React.Fragment key={i}>
                <span className="text-[#e8e8f0]">{id.slice(0, 6)}</span>
                {i < rootNode.propagationPath.length - 1 && <span>→</span>}
              </React.Fragment>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
