/**
 * @functionfly/ui-graph
 * ConditionalBranchRenderer - visualize conditional branching
 */

import * as React from "react";
import { cn } from "./utils";
import type { ConditionalBranchRendererProps, ConditionalBranch } from "./types";

export function ConditionalBranchRenderer({
  branches,
  activeBranchId,
  onBranchSelect,
  showCondition = true,
  className,
}: ConditionalBranchRendererProps) {
  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      <div className="flex items-center justify-between">
        <span className="text-xs text-[#6b7280]">Conditional Branches</span>
        <span className="text-[10px] text-[#f97316] font-mono">{branches.length} branches</span>
      </div>

      {branches.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-center">
          <div className="text-2xl opacity-30 mb-1">🔀</div>
          <p className="text-xs text-[#6b7280]">No branches defined</p>
          <p className="text-[10px] text-[#4b5563] mt-1">Conditional nodes will show branches here</p>
        </div>
      ) : (
        <div className="space-y-2">
          {branches.map((branch) => (
            <BranchCard
              key={branch.id}
              branch={branch}
              isActive={branch.id === activeBranchId}
              showCondition={showCondition}
              onSelect={onBranchSelect}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function BranchCard({
  branch,
  isActive,
  showCondition,
  onSelect,
}: {
  branch: ConditionalBranch;
  isActive: boolean;
  showCondition: boolean;
  onSelect?: ConditionalBranchRendererProps["onBranchSelect"];
}) {
  return (
    <div
      className={cn(
        "flex items-start gap-3 p-3 rounded-lg border transition-all cursor-pointer",
        isActive
          ? "border-[rgba(249,115,22,0.4)] bg-[rgba(249,115,22,0.1)] shadow-[0_0_16px_rgba(249,115,22,0.1)]"
          : "border-[rgba(255,255,255,0.06)] hover:border-[rgba(255,255,255,0.1)]"
      )}
      onClick={() => onSelect?.(branch.id)}
    >
      {/* Branch indicator */}
      <div
        className={cn(
          "size-6 rounded-md flex items-center justify-center text-xs font-bold shrink-0 mt-0.5",
          isActive ? "bg-[#f97316] text-white" : "bg-[#14141f] text-[#6b7280]"
        )}
      >
        {branch.isActive ? "✓" : branch.path.charAt(0)}
      </div>

      <div className="flex-1 min-w-0">
        {/* Branch label */}
        <div className="flex items-center gap-2">
          <span className={cn("text-sm", isActive ? "text-[#e8e8f0]" : "text-[#9ca3af]")}>
            {branch.label}
          </span>
          {branch.isActive && (
            <span className="text-[9px] px-1.5 py-0.5 rounded bg-[rgba(249,115,22,0.2)] text-[#f97316]">
              active
            </span>
          )}
        </div>

        {/* Condition */}
        {showCondition && branch.condition && (
          <div className="mt-1 px-2 py-1 rounded bg-[#14141f] text-[10px] text-[#6b7280] font-mono">
            if ({branch.condition})
          </div>
        )}

        {/* Nodes summary */}
        <div className="mt-2 flex flex-wrap gap-1">
          {branch.nodes.slice(0, 3).map((node) => (
            <div
              key={node.id}
              className="px-2 py-1 rounded text-[10px] bg-[#14141f] text-[#9ca3af]"
            >
              {node.label}
            </div>
          ))}
          {branch.nodes.length > 3 && (
            <div className="px-2 py-1 rounded text-[10px] bg-[#14141f] text-[#6b7280]">
              +{branch.nodes.length - 3} more
            </div>
          )}
        </div>
      </div>

      {/* Path indicator */}
      <div className="text-[10px] text-[#4b5563] font-mono shrink-0">{branch.path}</div>
    </div>
  );
}
