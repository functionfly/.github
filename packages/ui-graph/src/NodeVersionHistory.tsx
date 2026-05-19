/**
 * @functionfly/ui-graph
 * NodeVersionHistory - version history for a node
 */

import * as React from "react";
import { cn } from "./utils";
import type { NodeVersionHistoryProps, NodeVersion } from "./types";

function formatTimestamp(ts: number): string {
  const date = new Date(ts);
  const now = new Date();
  const diff = now.getTime() - date.getTime();

  if (diff < 60000) return "Just now";
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}d ago`;

  return date.toLocaleDateString();
}

export function NodeVersionHistory({
  nodeId,
  nodeLabel,
  versions,
  onVersionSelect,
  onVersionRestore,
  className,
}: NodeVersionHistoryProps) {
  const [selectedVersion, setSelectedVersion] = React.useState<string | null>(null);

  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-base">📜</span>
          <div>
            <div className="text-xs text-[#e8e8f0]">{nodeLabel}</div>
            <div className="text-[10px] text-[#6b7280] font-mono">{nodeId.slice(0, 8)}</div>
          </div>
        </div>
        <span className="text-[10px] text-[#f97316] font-mono">{versions.length} versions</span>
      </div>

      {/* Version list */}
      {versions.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-center">
          <div className="text-2xl opacity-30 mb-1">📋</div>
          <p className="text-xs text-[#6b7280]">No version history</p>
          <p className="text-[10px] text-[#4b5563] mt-1">Saved versions will appear here</p>
        </div>
      ) : (
        <div className="space-y-1 max-h-64 overflow-y-auto">
          {[...versions].reverse().map((version) => (
            <VersionCard
              key={version.version}
              version={version}
              isSelected={selectedVersion === version.version}
              onSelect={() => {
                setSelectedVersion(version.version);
                onVersionSelect?.(version);
              }}
            />
          ))}
        </div>
      )}

      {/* Restore button */}
      {selectedVersion && onVersionRestore && (
        <div className="pt-2 border-t border-[rgba(255,255,255,0.06)]">
          <button
            onClick={() => {
              const version = versions.find((v) => v.version === selectedVersion);
              if (version) onVersionRestore(version);
            }}
            className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-[#f97316]/20 text-[#f97316] hover:bg-[#f97316]/30 transition-colors text-xs font-medium"
          >
            <span>↩</span> Restore version {selectedVersion}
          </button>
        </div>
      )}
    </div>
  );
}

function VersionCard({
  version,
  isSelected,
  onSelect,
}: {
  version: NodeVersion;
  isSelected: boolean;
  onSelect: () => void;
}) {
  return (
    <div
      onClick={onSelect}
      className={cn(
        "flex items-center justify-between px-3 py-2 rounded-lg cursor-pointer transition-all border",
        isSelected
          ? "border-[rgba(249,115,22,0.4)] bg-[rgba(249,115,22,0.1)]"
          : "border-[rgba(255,255,255,0.06)] hover:border-[rgba(255,255,255,0.1)]"
      )}
    >
      <div className="flex items-center gap-3">
        {/* Version badge */}
        <div
          className={cn(
            "size-8 rounded-lg flex items-center justify-center text-xs font-bold",
            version.isActive
              ? "bg-[#10b981] text-white"
              : "bg-[#14141f] text-[#6b7280]"
          )}
        >
          v{version.version}
        </div>

        <div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-[#e8e8f0]">{formatTimestamp(version.timestamp)}</span>
            {version.isActive && (
              <span className="text-[9px] px-1.5 py-0.5 rounded bg-[rgba(16,185,129,0.2)] text-[#10b981]">
                active
              </span>
            )}
          </div>
          {version.author && (
            <div className="text-[10px] text-[#6b7280]">by {version.author}</div>
          )}
        </div>
      </div>

      {/* Changes indicator */}
      {version.changes && (
        <div
          className="max-w-[120px] truncate text-[10px] text-[#6b7280]"
          title={version.changes}
        >
          {version.changes}
        </div>
      )}
    </div>
  );
}
