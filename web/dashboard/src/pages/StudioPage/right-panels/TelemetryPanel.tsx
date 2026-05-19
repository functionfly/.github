import React from "react";
import {
  LiveTelemetryPanel,
  CostHeatmap,
  LatencyGraph,
} from "@functionfly/ui-observability";
import type { TelemetryMetric, TokenUsage } from "@functionfly/ui-observability";
import { BarChart2, Zap, AlertTriangle, CheckCircle } from "lucide-react";

interface TelemetryPanelProps {
  metrics: TelemetryMetric[];
  tokenUsage: TokenUsage;
  onMetricClick?: (metric: TelemetryMetric) => void;
}

export function TelemetryPanel({ metrics, tokenUsage, onMetricClick }: TelemetryPanelProps) {
  return (
    <div className="p-3 space-y-4">
      <LiveTelemetryPanel
        metrics={metrics}
        tokenUsage={tokenUsage}
        showTokenUsage={true}
        showAlerts={true}
        timeRange="24h"
        onMetricClick={onMetricClick}
      />

      <div className="border-t border-border-subtle pt-4">
        <h4 className="text-xs font-medium mb-2 flex items-center gap-2">
          <BarChart2 className="size-3 text-brand-400" />
          Cost Heatmap
        </h4>
        <CostHeatmap data={[]} />
      </div>

      <div className="border-t border-border-subtle pt-4">
        <h4 className="text-xs font-medium mb-2 flex items-center gap-2">
          <Zap className="size-3 text-warning" />
          Latency Distribution
        </h4>
        <LatencyGraph
          data={metrics.map((m) => ({
            timestamp: m.timestamp,
            p50: m.value * 0.5,
            p95: m.value * 0.9,
            p99: m.value,
          }))}
          timeRange="24h"
        />
      </div>
    </div>
  );
}