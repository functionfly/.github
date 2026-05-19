/**
 * @functionfly/ui-graph
 * ResourceConsumptionOverlay - visualize resource consumption across nodes
 */

import * as React from "react";
import { cn } from "./utils";
import type { ResourceConsumptionOverlayProps, ResourceMetric } from "./types";

function formatBytes(mb: number): string {
  if (mb >= 1000) return `${(mb / 1000).toFixed(1)} GB`;
  if (mb >= 1) return `${mb.toFixed(1)} MB`;
  return `${(mb * 1024).toFixed(0)} KB`;
}

function formatPercent(value: number): string {
  return `${value.toFixed(1)}%`;
}

const RESOURCE_ICONS = {
  cpu: "⚡",
  memory: "💾",
  network: "🌐",
  storage: "💽",
  gpu: "🖥️",
};

export function ResourceConsumptionOverlay({
  metrics,
  aggregatedMetrics,
  showLabels = true,
  aggregationWindow = 1000,
  className,
}: ResourceConsumptionOverlayProps) {
  const [timeWindow, setTimeWindow] = React.useState(aggregationWindow);

  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <span className="text-xs text-[#6b7280]">Resource Consumption</span>
        <div className="flex items-center gap-2">
          <select
            value={timeWindow}
            onChange={(e) => setTimeWindow(Number(e.target.value))}
            className="px-2 py-1 rounded text-[10px] bg-[#14141f] border border-[rgba(255,255,255,0.08)] text-[#e8e8f0]"
          >
            <option value={1000}>1s window</option>
            <option value={5000}>5s window</option>
            <option value={10000}>10s window</option>
            <option value={60000}>1m window</option>
          </select>
        </div>
      </div>

      {/* Aggregated totals */}
      {aggregatedMetrics && (
        <div className="grid grid-cols-5 gap-2">
          <AggregatedMetricCard
            label="CPU"
            icon="⚡"
            value={formatPercent(aggregatedMetrics.totalCpu)}
            color="#f97316"
          />
          <AggregatedMetricCard
            label="Memory"
            icon="💾"
            value={formatBytes(aggregatedMetrics.totalMemory)}
            color="#3b82f6"
          />
          <AggregatedMetricCard
            label="Network"
            icon="🌐"
            value={formatBytes(aggregatedMetrics.totalNetwork)}
            color="#10b981"
          />
          <AggregatedMetricCard
            label="Storage"
            icon="💽"
            value={formatBytes(aggregatedMetrics.totalStorage)}
            color="#8b5cf6"
          />
          <AggregatedMetricCard
            label="GPU"
            icon="🖥️"
            value={formatPercent(aggregatedMetrics.totalGpu)}
            color="#ec4899"
          />
        </div>
      )}

      {/* Per-node breakdown */}
      {metrics.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-4 text-center">
          <div className="text-2xl opacity-30 mb-1">📊</div>
          <p className="text-xs text-[#6b7280]">No resource data</p>
        </div>
      ) : (
        <div className="space-y-1.5 max-h-48 overflow-y-auto">
          {metrics.map((metric) => (
            <NodeResourceRow
              key={metric.nodeId}
              metric={metric}
              showLabel={showLabels}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function AggregatedMetricCard({
  label,
  icon,
  value,
  color,
}: {
  label: string;
  icon: string;
  value: string;
  color: string;
}) {
  return (
    <div className="flex flex-col items-center p-2 rounded-lg bg-[#14141f] border border-[rgba(255,255,255,0.06)]">
      <span className="text-lg">{icon}</span>
      <span className="text-lg font-mono font-medium text-[#e8e8f0]">{value}</span>
      <span className="text-[10px] text-[#6b7280]">{label}</span>
    </div>
  );
}

function NodeResourceRow({ metric, showLabel }: { metric: ResourceMetric; showLabel: boolean }) {
  return (
    <div className="flex items-center gap-3 px-2 py-1.5 rounded-lg bg-[#14141f]">
      {/* Node ID */}
      <div className="w-16 truncate">
        <span className="text-[10px] text-[#6b7280] font-mono">{metric.nodeId.slice(0, 8)}</span>
      </div>

      {/* Resource bars */}
      <div className="flex-1 flex items-center gap-2">
        {metric.cpuPercent != null && (
          <ResourceBar label="CPU" value={metric.cpuPercent} max={100} color="#f97316" />
        )}
        {metric.memoryMB != null && (
          <ResourceBar label="MEM" value={metric.memoryMB} max={4096} color="#3b82f6" />
        )}
        {metric.networkMB != null && (
          <ResourceBar label="NET" value={metric.networkMB} max={100} color="#10b981" />
        )}
        {metric.gpuPercent != null && (
          <ResourceBar label="GPU" value={metric.gpuPercent} max={100} color="#ec4899" />
        )}
      </div>
    </div>
  );
}

function ResourceBar({
  label,
  value,
  max,
  color,
}: {
  label: string;
  value: number;
  max: number;
  color: string;
}) {
  const pct = Math.min(100, (value / max) * 100);

  return (
    <div className="flex items-center gap-1 flex-1">
      <span className="text-[9px] text-[#6b7280] w-6">{label}</span>
      <div className="flex-1 h-1.5 bg-[#1a1a28] rounded-full overflow-hidden">
        <div
          className="h-full rounded-full transition-all"
          style={{ width: `${pct}%`, backgroundColor: color }}
        />
      </div>
      <span className="text-[9px] text-[#6b7280] font-mono w-10 text-right">
        {value.toFixed(0)}
      </span>
    </div>
  );
}
