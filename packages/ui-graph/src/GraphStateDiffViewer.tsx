/**
 * @functionfly/ui-graph
 * GraphStateDiffViewer - compare two graph states
 */

import * as React from "react";
import { cn } from "./utils";
import type { GraphStateDiffViewerProps, StateDiff, CanvasState } from "./types";

function deepDiff(left: unknown, right: unknown, path = ""): StateDiff[] {
  const diffs: StateDiff[] = [];

  if (left === right) return diffs;

  if (typeof left !== typeof right) {
    diffs.push({ type: "modified", path, oldValue: left, newValue: right });
    return diffs;
  }

  if (typeof left !== "object" || left === null || right === null) {
    diffs.push({ type: "modified", path, oldValue: left, newValue: right });
    return diffs;
  }

  const leftObj = left as Record<string, unknown>;
  const rightObj = right as Record<string, unknown>;
  const allKeys = new Set([...Object.keys(leftObj), ...Object.keys(rightObj)]);

  for (const key of allKeys) {
    const newPath = path ? `${path}.${key}` : key;
    if (!(key in leftObj)) {
      diffs.push({ type: "added", path: newPath, newValue: rightObj[key] });
    } else if (!(key in rightObj)) {
      diffs.push({ type: "removed", path: newPath, oldValue: leftObj[key] });
    } else {
      diffs.push(...deepDiff(leftObj[key], rightObj[key], newPath));
    }
  }

  return diffs;
}

function NodeListDiff({ label, nodes, side }: { label: string; nodes: CanvasState["nodes"]; side: "left" | "right" }) {
  return (
    <div className="flex-1 min-w-0">
      <div className="text-[10px] text-[#6b7280] uppercase tracking-wide mb-2">{label}</div>
      <div className="space-y-1 max-h-48 overflow-y-auto">
        {nodes.length === 0 ? (
          <div className="text-xs text-[#4b5563] italic">No nodes</div>
        ) : (
          nodes.map((node) => (
            <div
              key={node.id}
              className={cn(
                "flex items-center gap-2 px-2 py-1 rounded text-xs",
                side === "left" ? "bg-[rgba(239,68,68,0.1)]" : "bg-[rgba(16,185,129,0.1)]"
              )}
            >
              <span className="text-[#6b7280]">{node.type}</span>
              <span className="text-[#e8e8f0] truncate">{node.label}</span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

export function GraphStateDiffViewer({
  leftState,
  rightState,
  diffType = "full",
  onRestore,
  className,
}: GraphStateDiffViewerProps) {
  const [selectedDiff, setSelectedDiff] = React.useState<StateDiff | null>(null);

  const diffs = React.useMemo(() => {
    if (diffType === "nodes") {
      return deepDiff(leftState.nodes, rightState.nodes, "nodes");
    }
    if (diffType === "edges") {
      return deepDiff(leftState.edges, rightState.edges, "edges");
    }
    if (diffType === "viewport") {
      return deepDiff(leftState.viewport, rightState.viewport, "viewport");
    }
    return deepDiff(leftState, rightState);
  }, [leftState, rightState, diffType]);

  const addedDiffs = diffs.filter((d) => d.type === "added");
  const removedDiffs = diffs.filter((d) => d.type === "removed");
  const modifiedDiffs = diffs.filter((d) => d.type === "modified");

  return (
    <div className={cn("flex flex-col gap-4 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <span className="text-xs text-[#6b7280]">State Diff</span>
        <div className="flex items-center gap-3 text-[10px]">
          <span className="text-[#ef4444]">−{removedDiffs.length}</span>
          <span className="text-[#f97316]">±{modifiedDiffs.length}</span>
          <span className="text-[#10b981]">+{addedDiffs.length}</span>
        </div>
      </div>

      {/* Side by side comparison */}
      {diffType === "full" && (
        <div className="flex gap-4">
          <NodeListDiff label="Left State" nodes={leftState.nodes} side="left" />
          <NodeListDiff label="Right State" nodes={rightState.nodes} side="right" />
        </div>
      )}

      {/* Diff list */}
      <div className="space-y-1 max-h-64 overflow-y-auto">
        {diffs.length === 0 ? (
          <div className="text-center py-4 text-xs text-[#6b7280]">No differences found</div>
        ) : (
          diffs.map((diff, index) => (
            <div
              key={index}
              onClick={() => setSelectedDiff(diff)}
              className={cn(
                "flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer text-xs transition-colors",
                diff.type === "added" && "bg-[rgba(16,185,129,0.1)] hover:bg-[rgba(16,185,129,0.2)]",
                diff.type === "removed" && "bg-[rgba(239,68,68,0.1)] hover:bg-[rgba(239,68,68,0.2)]",
                diff.type === "modified" && "bg-[rgba(245,158,11,0.1)] hover:bg-[rgba(245,158,11,0.2)]"
              )}
            >
              <span
                className={cn(
                  "size-4 rounded text-[9px] font-bold flex items-center justify-center",
                  diff.type === "added" && "bg-[#10b981] text-white",
                  diff.type === "removed" && "bg-[#ef4444] text-white",
                  diff.type === "modified" && "bg-[#f59e0b] text-white"
                )}
              >
                {diff.type === "added" ? "+" : diff.type === "removed" ? "−" : "±"}
              </span>
              <span className="text-[#e8e8f0] font-mono truncate">{diff.path}</span>
            </div>
          ))
        )}
      </div>

      {/* Selected diff detail */}
      {selectedDiff && (
        <div className="flex flex-col gap-2 px-3 py-2 rounded-lg bg-[#14141f] border border-[rgba(255,255,255,0.06)]">
          <div className="text-[10px] text-[#6b7280] font-mono">{selectedDiff.path}</div>
          <div className="grid grid-cols-2 gap-2 text-[10px]">
            {selectedDiff.oldValue !== undefined && (
              <div>
                <span className="text-[#6b7280]">Old: </span>
                <span className="text-[#ef4444] font-mono">
                  {typeof selectedDiff.oldValue === "object"
                    ? JSON.stringify(selectedDiff.oldValue)
                    : String(selectedDiff.oldValue)}
                </span>
              </div>
            )}
            {selectedDiff.newValue !== undefined && (
              <div>
                <span className="text-[#6b7280]">New: </span>
                <span className="text-[#10b981] font-mono">
                  {typeof selectedDiff.newValue === "object"
                    ? JSON.stringify(selectedDiff.newValue)
                    : String(selectedDiff.newValue)}
                </span>
              </div>
            )}
          </div>
          {onRestore && (
            <button
              onClick={() => onRestore(selectedDiff)}
              className="mt-1 px-2 py-1 rounded text-[10px] bg-[#f97316]/20 text-[#f97316] hover:bg-[#f97316]/30 transition-colors"
            >
              Restore this value
            </button>
          )}
        </div>
      )}
    </div>
  );
}
