/**
 * @functionfly/ui-graph
 * NodeInspector - detailed node information panel
 */

import * as React from "react";
import { cn } from "./utils";
import type { NodeInspectorProps, NodeInspectorTab, NodeData } from "./types";

const DEFAULT_TABS: NodeInspectorTab[] = [
  { id: "properties", label: "Properties", icon: "📋" },
  { id: "metrics", label: "Metrics", icon: "📊" },
  { id: "inputs", label: "Inputs", icon: "⬇️" },
  { id: "outputs", label: "Outputs", icon: "⬆️" },
];

const STATUS_COLORS = {
  idle: "#6b7280",
  running: "#3b82f6",
  completed: "#10b981",
  error: "#ef4444",
  waiting: "#f59e0b",
};

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  if (ms < 3600000) return `${(ms / 60000).toFixed(1)}m`;
  return `${(ms / 3600000).toFixed(1)}h`;
}

export function NodeInspector({
  node,
  tabs = DEFAULT_TABS,
  activeTab = "properties",
  onTabChange,
  onClose,
  onUpdate,
  onDelete,
  className,
}: NodeInspectorProps) {
  const [localActiveTab, setLocalActiveTab] = React.useState(activeTab);
  const activeTab_ = onTabChange ? activeTab : localActiveTab;
  const setActiveTab = onTabChange ?? setLocalActiveTab;

  return (
    <div
      className={cn(
        "flex flex-col w-72 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]",
        "shadow-xl shadow-black/20 overflow-hidden",
        className
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2.5 bg-[#14141f] border-b border-[rgba(255,255,255,0.06)]">
        <div className="flex items-center gap-2">
          <span className="text-base">{node.icon ?? "⚡"}</span>
          <div>
            <div className="text-sm font-semibold text-[#e8e8f0] truncate max-w-[160px]">
              {node.label}
            </div>
            <div className="text-[10px] text-[#6b7280] capitalize">{node.type}</div>
          </div>
        </div>
        <button
          onClick={onClose}
          className="size-6 flex items-center justify-center rounded-md text-[#6b7280] hover:text-[#e8e8f0] hover:bg-[rgba(255,255,255,0.06)] transition-colors"
        >
          ✕
        </button>
      </div>

      {/* Status bar */}
      <div className="px-3 py-2 bg-[#0d0d14] border-b border-[rgba(255,255,255,0.06)]">
        <div className="flex items-center gap-2">
          <div
            className={cn("size-2 rounded-full", node.status === "running" && "animate-pulse")}
            style={{
              backgroundColor: STATUS_COLORS[node.status || "idle"],
              boxShadow: node.status === "running" ? `0 0 6px ${STATUS_COLORS.running}` : "none",
            }}
          />
          <span className="text-xs text-[#9ca3af] capitalize">{node.status || "idle"}</span>
          {node.executionTime != null && (
            <>
              <span className="text-[#4b5563]">•</span>
              <span className="text-xs text-[#9ca3af]">{formatDuration(node.executionTime)}</span>
            </>
          )}
          {node.version && (
            <>
              <span className="text-[#4b5563]">•</span>
              <span className="text-xs font-mono text-[#f97316]">v{node.version}</span>
            </>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex px-2 pt-2 bg-[#0d0d14] gap-1">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              "flex-1 flex items-center justify-center gap-1 px-2 py-1.5 rounded-lg text-[10px] font-medium transition-all",
              activeTab_ === tab.id
                ? "bg-[rgba(249,115,22,0.15)] text-[#f97316] border border-[rgba(249,115,22,0.3)]"
                : "text-[#6b7280] hover:text-[#9ca3af] hover:bg-[rgba(255,255,255,0.04)]"
            )}
          >
            <span>{tab.icon}</span>
            <span className="hidden sm:inline">{tab.label}</span>
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {activeTab_ === "properties" && (
          <PropertiesTab node={node} onUpdate={onUpdate} />
        )}
        {activeTab_ === "metrics" && <MetricsTab node={node} />}
        {activeTab_ === "inputs" && <PortsTab node={node} portType="inputs" />}
        {activeTab_ === "outputs" && <PortsTab node={node} portType="outputs" />}
      </div>

      {/* Footer actions */}
      <div className="flex items-center gap-2 px-3 py-2 bg-[#14141f] border-t border-[rgba(255,255,255,0.06)]">
        <button
          onClick={() => onDelete?.(node.id)}
          className="flex-1 flex items-center justify-center gap-1 px-2 py-1.5 rounded-lg text-[10px] font-medium text-[#ef4444] hover:bg-[rgba(239,68,68,0.1)] transition-colors"
        >
          <span>🗑️</span> Delete
        </button>
      </div>
    </div>
  );
}

function PropertiesTab({ node, onUpdate }: { node: NodeData; onUpdate?: NodeInspectorProps["onUpdate"] }) {
  return (
    <div className="space-y-2.5">
      {/* Description */}
      {node.description && (
        <div>
          <label className="text-[10px] text-[#6b7280] uppercase tracking-wide">Description</label>
          <p className="text-xs text-[#e8e8f0] mt-1">{node.description}</p>
        </div>
      )}

      {/* Metadata */}
      {node.metadata && Object.keys(node.metadata).length > 0 && (
        <div>
          <label className="text-[10px] text-[#6b7280] uppercase tracking-wide">Metadata</label>
          <div className="mt-1 space-y-1">
            {Object.entries(node.metadata).map(([key, value]) => (
              <div key={key} className="flex items-center justify-between">
                <span className="text-[10px] text-[#6b7280]">{key}</span>
                <span className="text-[10px] text-[#e8e8f0] font-mono">
                  {typeof value === "object" ? JSON.stringify(value) : String(value)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Cost */}
      {(node.tokenCost != null || node.costPerExecution != null) && (
        <div>
          <label className="text-[10px] text-[#6b7280] uppercase tracking-wide">Cost</label>
          <div className="mt-1 space-y-1">
            {node.tokenCost != null && (
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-[#6b7280]">Token cost</span>
                <span className="text-[10px] text-[#e8e8f0] font-mono">${node.tokenCost.toFixed(4)}</span>
              </div>
            )}
            {node.costPerExecution != null && (
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-[#6b7280]">Per execution</span>
                <span className="text-[10px] text-[#e8e8f0] font-mono">${node.costPerExecution.toFixed(4)}</span>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Last executed */}
      {node.lastExecutedAt && (
        <div>
          <label className="text-[10px] text-[#6b7280] uppercase tracking-wide">Last Executed</label>
          <p className="text-xs text-[#e8e8f0] mt-1 font-mono">{node.lastExecutedAt}</p>
        </div>
      )}
    </div>
  );
}

function MetricsTab({ node }: { node: NodeData }) {
  return (
    <div className="space-y-3">
      {/* Reliability score */}
      {node.reliabilityScore != null && (
        <div>
          <div className="flex items-center justify-between mb-1">
            <label className="text-[10px] text-[#6b7280] uppercase tracking-wide">Reliability</label>
            <span className="text-[10px] text-[#e8e8f0] font-mono">{node.reliabilityScore}%</span>
          </div>
          <div className="h-1.5 bg-[#14141f] rounded-full overflow-hidden">
            <div
              className="h-full rounded-full transition-all"
              style={{
                width: `${node.reliabilityScore}%`,
                backgroundColor:
                  node.reliabilityScore > 80 ? "#10b981" : node.reliabilityScore > 50 ? "#f59e0b" : "#ef4444",
              }}
            />
          </div>
        </div>
      )}

      {/* Error rate */}
      {node.errorRate != null && (
        <div>
          <div className="flex items-center justify-between mb-1">
            <label className="text-[10px] text-[#6b7280] uppercase tracking-wide">Error Rate</label>
            <span className="text-[10px] text-[#e8e8f0] font-mono">{node.errorRate.toFixed(2)}%</span>
          </div>
          <div className="h-1.5 bg-[#14141f] rounded-full overflow-hidden">
            <div
              className="h-full rounded-full transition-all bg-[#ef4444]"
              style={{ width: `${Math.min(node.errorRate, 100)}%` }}
            />
          </div>
        </div>
      )}

      {/* Execution time */}
      {node.executionTime != null && (
        <div>
          <label className="text-[10px] text-[#6b7280] uppercase tracking-wide">Execution Time</label>
          <p className="text-lg text-[#e8e8f0] font-mono mt-1">{formatDuration(node.executionTime)}</p>
        </div>
      )}

      {/* Token cost */}
      {node.tokenCost != null && (
        <div>
          <label className="text-[10px] text-[#6b7280] uppercase tracking-wide">Token Cost</label>
          <p className="text-lg text-[#e8e8f0] font-mono mt-1">${node.tokenCost.toFixed(4)}</p>
        </div>
      )}
    </div>
  );
}

function PortsTab({ node, portType }: { node: NodeData; portType: "inputs" | "outputs" }) {
  const ports = node[portType] || [];

  if (ports.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-6 text-center">
        <div className="text-2xl opacity-30 mb-1">{portType === "inputs" ? "⬇️" : "⬆️"}</div>
        <p className="text-xs text-[#6b7280]">No {portType} defined</p>
      </div>
    );
  }

  return (
    <div className="space-y-1.5">
      {ports.map((port) => (
        <div
          key={port.id}
          className="flex items-center justify-between px-2 py-1.5 rounded-lg bg-[#14141f] border border-[rgba(255,255,255,0.06)]"
        >
          <div className="flex items-center gap-2">
            <div
              className={cn(
                "size-2 rounded-full",
                port.type === "input" ? "bg-[#3b82f6]" : "bg-[#10b981]"
              )}
            />
            <span className="text-xs text-[#e8e8f0]">{port.label}</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-[10px] text-[#6b7280] font-mono">{port.dataType || "any"}</span>
            {port.multiple && (
              <span className="text-[8px] px-1 py-0.5 rounded bg-[rgba(139,92,246,0.2)] text-[#8b5cf6]">
                multi
              </span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
