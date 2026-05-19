/**
 * @functionfly/ui-graph
 * WorkflowForkManager - manage parallel execution branches
 */

import * as React from "react";
import { cn } from "./utils";
import type { WorkflowForkManagerProps, ForkPoint } from "./types";

export function WorkflowForkManager({
  forks,
  onForkSelect,
  onForkCreate,
  onForkDelete,
  className,
}: WorkflowForkManagerProps) {
  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      <div className="flex items-center justify-between">
        <span className="text-xs text-[#6b7280]">Fork Manager</span>
        <span className="text-[10px] text-[#f97316] font-mono">{forks.length} forks</span>
      </div>

      {forks.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-center">
          <div className="text-2xl opacity-30 mb-1">🔀</div>
          <p className="text-xs text-[#6b7280]">No fork points</p>
          <p className="text-[10px] text-[#4b5563] mt-1">Workflows with parallel branches will appear here</p>
        </div>
      ) : (
        <div className="space-y-2">
          {forks.map((fork) => (
            <ForkCard
              key={fork.id}
              fork={fork}
              onSelect={onForkSelect}
              onDelete={onForkDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function ForkCard({
  fork,
  onSelect,
  onDelete,
}: {
  fork: ForkPoint;
  onSelect?: WorkflowForkManagerProps["onForkSelect"];
  onDelete?: WorkflowForkManagerProps["onForkDelete"];
}) {
  return (
    <div className="flex items-center justify-between px-3 py-2.5 rounded-lg bg-[#14141f] border border-[rgba(255,255,255,0.06)] hover:border-[rgba(249,115,22,0.2)] transition-colors">
      <div className="flex items-center gap-3">
        <div className="size-8 flex items-center justify-center rounded-lg bg-[rgba(168,85,247,0.15)] text-[#a855f7]">
          🔀
        </div>
        <div>
          <div className="text-sm text-[#e8e8f0]">{fork.label || `Fork ${fork.id.slice(0, 6)}`}</div>
          <div className="text-[10px] text-[#6b7280]">{fork.branchCount} branches</div>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {/* Branch selector */}
        <select
          className="px-2 py-1 rounded text-[10px] bg-[#0d0d14] border border-[rgba(255,255,255,0.08)] text-[#e8e8f0]"
          value={fork.activeBranch || ""}
          onChange={(e) => onSelect?.(fork.id, e.target.value)}
        >
          <option value="">Select branch</option>
          {Array.from({ length: fork.branchCount }).map((_, i) => (
            <option key={i} value={`branch-${i}`}>
              Branch {i + 1}
            </option>
          ))}
        </select>

        <button
          onClick={() => onDelete?.(fork.id)}
          className="size-6 flex items-center justify-center rounded-md text-[#6b7280] hover:text-[#ef4444] hover:bg-[rgba(239,68,68,0.1)] transition-colors"
        >
          ✕
        </button>
      </div>
    </div>
  );
}
