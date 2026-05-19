/**
 * @functionfly/ui-graph
 * DependencyExplorer - visualize node dependencies
 */

import * as React from "react";
import { cn } from "./utils";
import type { DependencyExplorerProps, DependencyNode, NodeData, EdgeData, NodeType } from "./types";

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

function buildDependencyTree(
  nodes: NodeData[],
  edges: EdgeData[],
  rootId: string,
  direction: "downstream" | "upstream" | "both"
): DependencyNode[] {
  const result: DependencyNode[] = [];
  const visited = new Set<string>();
  const nodeMap = new Map(nodes.map((n) => [n.id, n]));

  const traverse = (nodeId: string, depth: number, path: string[]) => {
    if (visited.has(nodeId)) return;
    visited.add(nodeId);

    const node = nodeMap.get(nodeId);
    if (!node) return;

    const currentPath = [...path, nodeId];
    const isCyclic = currentPath.filter((id) => id === nodeId).length > 1;

    result.push({
      id: nodeId,
      label: node.label,
      type: node.type,
      depth,
      path: currentPath,
      isCyclic,
    });

    if (direction === "downstream" || direction === "both") {
      const outgoing = edges.filter((e) => e.source === nodeId);
      for (const edge of outgoing) {
        traverse(edge.target, depth + 1, currentPath);
      }
    }

    if (direction === "upstream" || direction === "both") {
      const incoming = edges.filter((e) => e.target === nodeId);
      for (const edge of incoming) {
        traverse(edge.source, depth + 1, currentPath);
      }
    }
  };

  traverse(rootId, 0, []);
  return result;
}

export function DependencyExplorer({
  nodes,
  edges,
  rootNodeId,
  direction = "downstream",
  onNodeClick,
  className,
}: DependencyExplorerProps) {
  const [selectedNodeId, setSelectedNodeId] = React.useState<string | null>(rootNodeId || null);

  const tree = React.useMemo(() => {
    if (!rootNodeId && !selectedNodeId) return [];
    return buildDependencyTree(nodes, edges, selectedNodeId!, direction);
  }, [nodes, edges, selectedNodeId, direction]);

  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <span className="text-xs text-[#6b7280]">Dependency Explorer</span>
        <select
          value={selectedNodeId || ""}
          onChange={(e) => setSelectedNodeId(e.target.value || null)}
          className="px-2 py-1 rounded text-[10px] bg-[#14141f] border border-[rgba(255,255,255,0.08)] text-[#e8e8f0]"
        >
          <option value="">Select root node</option>
          {nodes.map((n) => (
            <option key={n.id} value={n.id}>
              {n.label}
            </option>
          ))}
        </select>
      </div>

      {/* Tree view */}
      {tree.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-center">
          <div className="text-2xl opacity-30 mb-1">🌳</div>
          <p className="text-xs text-[#6b7280]">Select a root node</p>
          <p className="text-[10px] text-[#4b5563] mt-1">View dependencies starting from a node</p>
        </div>
      ) : (
        <div className="space-y-0.5 max-h-80 overflow-y-auto">
          {tree.map((node) => (
            <div
              key={node.id}
              onClick={() => onNodeClick?.(node.id)}
              className={cn(
                "flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer transition-colors",
                "hover:bg-[rgba(255,255,255,0.04)]",
                node.isCyclic && "bg-[rgba(239,68,68,0.1)]"
              )}
              style={{ paddingLeft: `${node.depth * 16 + 8}px` }}
            >
              {/* Connector line */}
              {node.depth > 0 && (
                <div
                  className="absolute left-0 w-4 h-px bg-[rgba(255,255,255,0.1)]"
                  style={{ left: `${(node.depth - 1) * 16 + 8}px` }}
                />
              )}

              {/* Node type indicator */}
              <div
                className="size-3 rounded-sm"
                style={{ backgroundColor: NODE_COLORS[node.type] || "#6b7280" }}
              />

              {/* Node info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-[#e8e8f0] truncate">{node.label}</span>
                  {node.isCyclic && (
                    <span className="text-[9px] px-1 py-0.5 rounded bg-[rgba(239,68,68,0.2)] text-[#ef4444]">
                      cyclic
                    </span>
                  )}
                </div>
                <div className="text-[10px] text-[#6b7280] font-mono">{node.path.join(" → ")}</div>
              </div>

              {/* Depth indicator */}
              <span className="text-[10px] text-[#4b5563] font-mono">#{node.depth}</span>
            </div>
          ))}
        </div>
      )}

      {/* Direction selector */}
      <div className="flex items-center gap-2 pt-2 border-t border-[rgba(255,255,255,0.06)]">
        <span className="text-[10px] text-[#6b7280]">Direction:</span>
        <div className="flex gap-1">
          {(["downstream", "upstream", "both"] as const).map((dir) => (
            <button
              key={dir}
              onClick={() => {}}
              className={cn(
                "px-2 py-1 rounded text-[10px] transition-colors",
                direction === dir
                  ? "bg-[#f97316]/20 text-[#f97316]"
                  : "bg-[#14141f] text-[#6b7280] hover:text-[#9ca3af]"
              )}
            >
              {dir === "downstream" ? "↓" : dir === "upstream" ? "↑" : "↕"} {dir}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
