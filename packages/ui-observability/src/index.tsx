/**
 * @functionfly/ui-observability
 * Live telemetry and monitoring components
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import {
  Activity,
  Clock,
  DollarSign,
  Zap,
  AlertTriangle,
  CheckCircle,
  XCircle,
  BarChart2,
  TrendingUp,
  TrendingDown,
  RefreshCcw,
  Gauge,
  Thermometer,
  Cpu,
  MemoryStick,
  Wifi,
  Server,
  HardDrive,
  Eye,
  Search,
  Filter,
  ArrowRight,
  ChevronRight,
  ChevronDown,
  Plus,
  Minus,
  Play,
  Pause,
  Square,
  RotateCcw,
  Mic,
  Volume2,
} from "lucide-react";

export interface TelemetryMetric {
  id: string;
  label: string;
  value: number;
  unit: string;
  timestamp: number;
  trend?: "up" | "down" | "stable";
  delta?: number;
  status?: "ok" | "warning" | "error";
  min?: number;
  max?: number;
  avg?: number;
  p50?: number;
  p95?: number;
  p99?: number;
}

export interface TokenUsage {
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  costUSD: number;
  models: Record<string, { calls: number; tokens: number; cost: number }>;
  timeRange: { start: string; end: string };
}

export interface LiveTelemetryPanelProps {
  metrics: TelemetryMetric[];
  tokenUsage?: TokenUsage;
  onMetricClick?: (metric: TelemetryMetric) => void;
  className?: string;
  showTokenUsage?: boolean;
  showAlerts?: boolean;
  timeRange?: "1h" | "6h" | "24h" | "7d" | "30d";
  onTimeRangeChange?: (range: "1h" | "6h" | "24h" | "7d" | "30d") => void;
}

export interface MetricCardProps {
  id?: string;
  label: string;
  value: number;
  unit: string;
  timestamp?: number;
  status?: "ok" | "warning" | "error";
  trend?: "up" | "down";
  delta?: number;
  size?: "sm" | "md" | "lg";
  sparkline?: number[];
  onMetricClick?: (metric: TelemetryMetric) => void;
}

export interface TokenUsageStreamProps {
  usage: TokenUsage;
  onTimeRangeChange?: (range: "1h" | "6h" | "24h" | "7d" | "30d") => void;
  className?: string;
}

export interface CostHeatmapProps {
  data: Array<{
    hour: number;
    day: string;
    cost: number;
    requests: number;
  }>;
  className?: string;
}

export interface LatencyGraphProps {
  data: Array<{
    timestamp: number;
    p50: number;
    p95: number;
    p99: number;
  }>;
  timeRange: "1h" | "6h" | "24h" | "7d" | "30d";
  height?: number;
  className?: string;
}

export interface ExecutionProfilerProps {
  executions: Array<{
    id: string;
    name: string;
    duration: number;
    tokens: number;
    cost: number;
    timestamp: string;
    status: "success" | "error" | "timeout";
    retries: number;
  }>;
  className?: string;
}

function getStatusColor(status: string): string {
  return (
    { ok: "#10b981", warning: "#f59e0b", error: "#ef4444" }[status] || "#6b7280"
  );
}

function getTrendIcon(trend?: string) {
  if (trend === "up") return <TrendingUp className="size-3 text-error" />;
  if (trend === "down") return <TrendingDown className="size-3 text-success" />;
  return <RefreshCcw className="size-3 text-text-muted" />;
}

function formatDelta(n?: number): string {
  if (n == null) return "";
  const sign = n >= 0 ? "+" : "";
  return `${sign}${n}%`;
}

export function MetricCard({
  label,
  value,
  unit,
  status = "ok",
  trend,
  delta,
  sparkline,
  size = "md",
  onMetricClick,
}: MetricCardProps) {
  const sizeClasses = {
    sm: "p-3 min-h-[80px]",
    md: "p-4 min-h-[100px]",
    lg: "p-5 min-h-[120px]",
  };
  return (
    <div
      className={cn(
        "bg-bg-secondary border border-border-subtle rounded-lg transition-all duration-200 hover:border-border-default cursor-pointer",
        sizeClasses[size],
        status === "error" && "border-error/30 bg-error/5",
        status === "warning" && "border-warning/30 bg-warning/5",
      )}
      onClick={() =>
        onMetricClick?.({
          label,
          value,
          unit,
          timestamp: Date.now(),
          id: "",
          status,
        } as TelemetryMetric)
      }
    >
      <div className="flex items-center justify-between mb-1">
        <span className="text-[11px] text-text-muted capitalize">{label}</span>
        <div className="flex items-center gap-1">
          {getTrendIcon(trend)}
          {delta != null && (
            <span
              className={`text-[10px] font-medium ${delta >= 0 ? "text-error" : "text-success"}`}
            >
              {formatDelta(delta)}
            </span>
          )}
          <div
            className={`size-2 rounded-full ${status === "error" ? "bg-error" : status === "warning" ? "bg-warning" : "bg-success"}`}
          />
        </div>
      </div>
      <div className="text-xl font-bold text-text-primary">
        {typeof value === "number" ? value.toLocaleString() : value}
        <span className="text-sm font-normal text-text-muted ml-1">{unit}</span>
      </div>
      {sparkline && sparkline.length > 0 && (
        <div className="mt-2 h-6 flex items-end gap-0.5">
          {sparkline.map((v, i) => (
            <div
              key={i}
              className="flex-1 rounded-t-[1px] transition-all duration-300"
              style={{
                height: `${Math.max(v * 100, 4)}%`,
                backgroundColor: getStatusColor(status),
                opacity: 0.4 + v * 0.6,
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function LiveTelemetryPanel({
  metrics,
  tokenUsage,
  onMetricClick,
  className,
  showTokenUsage = true,
  showAlerts = false,
  timeRange = "24h",
  onTimeRangeChange,
}: LiveTelemetryPanelProps) {
  const timeRanges: Array<"1h" | "6h" | "24h" | "7d" | "30d"> = [
    "1h",
    "6h",
    "24h",
    "7d",
    "30d",
  ];
  return (
    <div className={cn("space-y-4", className)}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Activity className="size-4 text-brand-400" />
          <span className="text-sm font-semibold text-text-primary">
            Live Telemetry
          </span>
          <div className="size-2 rounded-full bg-success animate-pulse" />
        </div>
        <div className="flex items-center gap-1">
          {timeRanges.map((r) => (
            <button
              key={r}
              onClick={() => onTimeRangeChange?.(r)}
              className={cn(
                "px-2 py-0.5 text-[10px] rounded transition-colors",
                timeRange === r
                  ? "bg-brand-500/20 text-brand-400 font-medium"
                  : "text-text-muted hover:text-text-secondary",
              )}
            >
              {r}
            </button>
          ))}
        </div>
      </div>
      {showAlerts && metrics.some((m) => m.status === "error") && (
        <div className="flex items-center gap-2 px-3 py-2 bg-error/10 border border-error/30 rounded-lg text-sm text-error">
          <AlertTriangle className="size-4 shrink-0" />
          <span>
            {metrics.filter((m) => m.status === "error").length} metrics in
            error state
          </span>
        </div>
      )}
      <div className="grid grid-cols-2 gap-3">
        <MetricCard
          label="Requests (24h)"
          value={12453}
          unit="req"
          sparkline={[0.3, 0.5, 0.8, 0.6, 0.9, 1, 0.7, 0.85, 0.6, 0.95]}
          size="md"
          onMetricClick={onMetricClick}
        />
        <MetricCard
          label="Avg Latency"
          value={42}
          unit="ms"
          status="ok"
          sparkline={[0.4, 0.3, 0.6, 0.5, 0.4, 0.7, 0.3, 0.5, 0.6, 0.4]}
          size="md"
          onMetricClick={onMetricClick}
        />
        <MetricCard
          label="Error Rate"
          value={0.2}
          unit="%"
          status="ok"
          sparkline={[0.1, 0.2, 0.1, 0.3, 0.2, 0.1, 0.4, 0.2, 0.1, 0.2]}
          size="md"
          onMetricClick={onMetricClick}
        />
        <MetricCard
          label="Active Agents"
          value={5}
          unit="agents"
          status="ok"
          sparkline={[0.2, 0.3, 0.5, 0.4, 0.6, 0.8, 0.7, 0.9, 0.8, 1]}
          size="md"
          onMetricClick={onMetricClick}
        />
      </div>
      {showTokenUsage && tokenUsage && (
        <TokenUsageStream
          usage={tokenUsage}
          onTimeRangeChange={onTimeRangeChange}
        />
      )}
    </div>
  );
}

export function TokenUsageStream({
  usage,
  onTimeRangeChange,
  className,
}: TokenUsageStreamProps) {
  const totalCost = Object.values(usage.models).reduce(
    (sum, m) => sum + m.cost,
    0,
  );
  return (
    <div
      className={cn(
        "bg-bg-secondary rounded-lg border border-border-subtle p-4",
        className,
      )}
    >
      <div className="flex items-center justify-between mb-3">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <Zap className="size-4 text-brand-400" /> Token Usage
        </h4>
        <div className="flex items-center gap-2">
          {Object.keys(usage.models).map((model) => (
            <button
              key={model}
              onClick={() => onTimeRangeChange?.("24h")}
              className="text-[10px] px-1.5 py-0.5 rounded bg-bg-primary text-text-muted border border-border-subtle hover:border-brand-500/30 transition-colors"
            >
              {model}
            </button>
          ))}
        </div>
      </div>
      <div className="grid grid-cols-2 gap-3 mb-3">
        <div className="bg-bg-primary rounded-lg p-3 min-w-0">
          <div className="text-[10px] text-text-muted mb-1">Total Tokens</div>
          <div className="text-lg font-bold text-text-primary">
            {usage.totalTokens.toLocaleString()}
          </div>
          <div className="text-[10px] text-text-muted mt-1">
            In: {usage.inputTokens.toLocaleString()} · Out:{" "}
            {usage.outputTokens.toLocaleString()}
          </div>
        </div>
        <div className="bg-bg-primary rounded-lg p-3 min-w-0">
          <div className="text-[10px] text-text-muted mb-1">Total Cost</div>
          <div className="text-lg font-bold text-brand-500">
            ${totalCost.toFixed(4)}
          </div>
          <div className="text-[10px] text-text-muted mt-1">
            {usage.totalTokens > 0
              ? `$${((totalCost / usage.totalTokens) * 1000).toFixed(4)}/1K tokens`
              : "$0.0000/1K tokens"}
          </div>
        </div>
        <div className="bg-bg-primary rounded-lg p-3 min-w-0">
          <div className="text-[10px] text-text-muted mb-1">Total Calls</div>
          <div className="text-lg font-bold text-text-primary">
            {Object.values(usage.models)
              .reduce((s, m) => s + m.calls, 0)
              .toLocaleString()}
          </div>
          <div className="text-[10px] text-text-muted mt-1">
            Models: {Object.keys(usage.models).length}
          </div>
        </div>
        <div className="bg-bg-primary rounded-lg p-3 min-w-0">
          <div className="text-[10px] text-text-muted mb-1">Cost Trend</div>
          <div className="text-lg font-bold text-success">↓ 12%</div>
          <div className="text-[10px] text-text-muted mt-1">
            vs. last 7 days
          </div>
        </div>
      </div>
      <div className="h-2 bg-bg-primary rounded-full overflow-hidden">
        {Object.entries(usage.models).map(([model, data], i, arr) => (
          <div
            key={model}
            className="h-full transition-all duration-500"
            style={{
              width:
                usage.totalTokens > 0
                  ? `${(data.tokens / usage.totalTokens) * 100}%`
                  : "0%",
              backgroundColor: `hsl(${i * 137.5}, 70%, 55%)`,
            }}
          />
        ))}
      </div>
      <div className="flex justify-between mt-1.5">
        {Object.entries(usage.models).map(([model, data]) => (
          <span key={model} className="text-[9px] text-text-muted font-mono">
            {model}: {data.tokens.toLocaleString()} tok
          </span>
        ))}
      </div>
    </div>
  );
}

export function CostHeatmap({ data, className }: CostHeatmapProps) {
  const maxCost = Math.max(...data.map((d) => d.cost), 1);
  return (
    <div
      className={cn(
        "bg-bg-secondary border border-border-subtle rounded-lg p-4",
        className,
      )}
    >
      <h4 className="text-sm font-medium text-text-primary mb-3">
        Cost Heatmap (Hours × Days)
      </h4>
      <div className="overflow-x-auto -mx-4 px-4">
        <div
          className="grid gap-[1px]"
          style={{
            gridTemplateColumns: "repeat(24, minmax(16px, 1fr))",
            minWidth: 400,
          }}
        >
          {Array.from({ length: 24 }, (_, h) => (
            <React.Fragment key={h}>
              <div className="text-[9px] text-text-muted text-center py-0.5 font-mono">
                {String(h).padStart(2, "0")}
              </div>
              {data
                .filter((d) => d.hour === h)
                .map((d, i) => (
                  <div
                    key={i}
                    className="aspect-square rounded-[1px]"
                    style={{
                      backgroundColor: `rgba(249, 115, 22, ${0.1 + (d.cost / maxCost) * 0.9})`,
                    }}
                    title={`$${d.cost.toFixed(4)} · ${d.requests} req`}
                  />
                ))}
            </React.Fragment>
          ))}
        </div>
      </div>
      <div className="flex items-center justify-between mt-2 text-[10px] text-text-muted">
        <span>Low cost</span>
        <div className="flex gap-0.5">
          {[0.1, 0.3, 0.5, 0.7, 0.9, 1].map((v) => (
            <div
              key={v}
              className="size-3 rounded-[1px]"
              style={{ backgroundColor: `rgba(249, 115, 22, ${v})` }}
            />
          ))}
        </div>
        <span>High cost</span>
      </div>
    </div>
  );
}

export function LatencyGraph({
  data,
  timeRange,
  height = 120,
  className,
}: LatencyGraphProps) {
  const maxP99 = Math.max(...data.map((d) => d.p99), 1);
  const polyline = (key: "p50" | "p95" | "p99", color: string) => {
    const points = data.map((d, i) => {
      const x = (i / (data.length - 1)) * 100;
      const y = 100 - (d[key] / maxP99) * 100;
      return `${x},${y}`;
    });
    return (
      <polyline
        key={key}
        points={points.join(" ")}
        fill="none"
        stroke={color}
        strokeWidth={key === "p99" ? 1.5 : 1}
        strokeDasharray={key === "p95" ? "4,3" : key === "p99" ? "2,3" : "none"}
        opacity={key === "p50" ? 1 : 0.7}
      />
    );
  };
  return (
    <div className={className}>
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-sm font-medium text-text-primary">
          <Clock className="size-4 inline mr-1.5 text-brand-400" />
          Latency (p50 / p95 / p99)
        </h4>
        <span className="text-[10px] text-text-muted">
          p99: {data[data.length - 1]?.p99 ?? 0}ms
        </span>
      </div>
      <svg
        width="100%"
        height={height}
        viewBox={`0 0 100 ${height}`}
        preserveAspectRatio="none"
        className="rounded-lg bg-bg-secondary border border-border-subtle"
      >
        <defs>
          <linearGradient id="latency-gradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#3b82f6" stopOpacity="0.2" />
            <stop offset="100%" stopColor="#3b82f6" stopOpacity="0" />
          </linearGradient>
        </defs>
        {[0.25, 0.5, 0.75].map((y) => (
          <line
            key={y}
            x1="0"
            y1={y * height}
            x2="100"
            y2={y * height}
            stroke="rgba(255,255,255,0.04)"
            strokeWidth="0.5"
          />
        ))}
        {data.length > 1 && (
          <polygon
            points={(() => {
              const pts = data.map(
                (d, i) =>
                  `${(i / (data.length - 1)) * 100},${100 - ((d.p50 || 0) / maxP99) * 100}`,
              );
              return `${pts.join(" ")} ${((data.length - 1) / (data.length - 1)) * 100},100 0,100`;
            })()}
            fill="url(#latency-gradient)"
          />
        )}
        {polyline("p50", "#3b82f6")}
        {polyline("p95", "#f59e0b")}
        {polyline("p99", "#ef4444")}
      </svg>
      <div className="flex items-center justify-between mt-1.5 text-[10px] text-text-muted">
        <span>{new Date(data[0]?.timestamp ?? 0).toLocaleTimeString()}</span>
        <div className="flex items-center gap-4">
          <span className="flex items-center gap-1">
            <span className="size-2 rounded-full bg-[#3b82f6]" />
            p50
          </span>
          <span className="flex items-center gap-1">
            <span className="size-2 rounded-full bg-[#f59e0b]" />
            p95
          </span>
          <span className="flex items-center gap-1">
            <span className="size-2 rounded-full bg-[#ef4444]" />
            p99
          </span>
        </div>
        <span>
          {new Date(data[data.length - 1]?.timestamp ?? 0).toLocaleTimeString()}
        </span>
      </div>
    </div>
  );
}

export function ExecutionProfiler({
  executions,
  className,
}: ExecutionProfilerProps) {
  const sorted = [...executions].sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
  );
  const statusColor = {
    success: "#10b981",
    error: "#ef4444",
    timeout: "#f59e0b",
  };
  return (
    <div className={cn("space-y-2", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <BarChart2 className="size-4 text-brand-400" /> Recent Executions
      </h4>
      <div className="space-y-1">
        {sorted.slice(0, 10).map((exec) => (
          <div
            key={exec.id}
            className="flex items-center justify-between p-2.5 bg-bg-primary rounded-lg border border-border-subtle hover:border-border-default transition-colors"
          >
            <div className="flex items-center gap-3 min-w-0 flex-1">
              <div
                className="size-2 rounded-full shrink-0"
                style={{ backgroundColor: statusColor[exec.status] }}
              />
              <span className="text-sm font-medium text-text-primary truncate">
                {exec.name}
              </span>
              <span className="text-[10px] text-text-muted truncate">
                {exec.timestamp}
              </span>
            </div>
            <div className="flex items-center gap-4 shrink-0">
              <span className="text-[10px] text-text-muted font-mono">
                {exec.duration}ms
              </span>
              <span className="text-[10px] text-text-muted font-mono">
                ${exec.cost.toFixed(4)}
              </span>
              <span className="text-[10px] text-text-muted">
                {exec.tokens.toLocaleString()} tok
              </span>
              {exec.retries > 0 && (
                <Badge variant="warning" size="sm">
                  {exec.retries} retry
                </Badge>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function Badge({ variant = "default", size = "md", className, children }: any) {
  const sizeClasses = {
    sm: "px-1.5 py-0.5 text-[10px]",
    md: "px-2 py-0.5 text-[11px]",
  };
  const variantClasses = {
    default: "bg-bg-tertiary text-text-secondary",
    brand: "bg-brand-500/20 text-brand-400",
    success: "bg-success/20 text-success",
    error: "bg-error/20 text-error",
    warning: "bg-warning/20 text-warning",
    ghost: "bg-transparent text-text-muted",
  };
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full font-medium ${sizeClasses[size]} ${variantClasses[variant]} ${className || ""}`}
    >
      {children}
    </span>
  );
}
// ============================================================================
// InferenceTraceViewer - View inference request traces
// ============================================================================

export interface InferenceSpan {
  id: string;
  parentId?: string;
  name: string;
  startTime: number;
  endTime?: number;
  duration?: number;
  status: "running" | "completed" | "failed";
  model?: string;
  inputTokens?: number;
  outputTokens?: number;
  error?: string;
  children?: InferenceSpan[];
}

export interface InferenceTraceViewerProps {
  traceId: string;
  spans: InferenceSpan[];
  onSpanClick?: (span: InferenceSpan) => void;
  selectedSpanId?: string;
  className?: string;
}

export function InferenceTraceViewer({
  traceId,
  spans,
  onSpanClick,
  selectedSpanId,
  className,
}: InferenceTraceViewerProps) {
  const [expandedSpans, setExpandedSpans] = React.useState<Set<string>>(
    new Set([spans[0]?.id]),
  );

  const toggleExpand = (id: string) => {
    setExpandedSpans((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const flattenSpans = (
    spanList: InferenceSpan[],
    depth = 0,
  ): Array<InferenceSpan & { depth: number }> => {
    const result: Array<InferenceSpan & { depth: number }> = [];
    spanList.forEach((span) => {
      result.push({ ...span, depth });
      if (expandedSpans.has(span.id) && span.children) {
        result.push(...flattenSpans(span.children, depth + 1));
      }
    });
    return result;
  };

  const flattenedSpans = flattenSpans(spans);

  return (
    <div className={cn("space-y-3", className)}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <Search className="size-4 text-brand-400" /> Inference Trace
        </h4>
        <div className="flex items-center gap-2 text-[10px] text-text-muted font-mono">
          <span>Trace:</span>
          <span className="text-brand-400">{traceId.slice(0, 8)}...</span>
        </div>
      </div>

      {/* Timeline header */}
      <div className="flex items-center gap-2 text-[10px] text-text-muted border-b border-border-subtle pb-2">
        <div className="w-48">Span</div>
        <div className="flex-1">Timeline</div>
        <div className="w-20 text-right">Duration</div>
        <div className="w-16 text-right">Tokens</div>
      </div>

      {/* Spans */}
      <div className="space-y-0.5 max-h-96 overflow-y-auto">
        {flattenedSpans.map((span) => {
          const isSelected = span.id === selectedSpanId;
          const hasChildren = span.children && span.children.length > 0;
          const statusColors = {
            running: "bg-blue-500",
            completed: "bg-success",
            failed: "bg-error",
          };

          return (
            <div
              key={span.id}
              onClick={() => onSpanClick?.(span)}
              className={cn(
                "flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer transition-colors text-[11px]",
                isSelected ? "bg-brand-500/20" : "hover:bg-bg-secondary",
              )}
              style={{ paddingLeft: `${span.depth * 16 + 8}px` }}
            >
              {/* Expand toggle */}
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  toggleExpand(span.id);
                }}
                className="size-4 flex items-center justify-center text-text-muted hover:text-text-primary"
              >
                {hasChildren ? (
                  expandedSpans.has(span.id) ? (
                    <ChevronDown className="size-3" />
                  ) : (
                    <ChevronRight className="size-3" />
                  )
                ) : null}
              </button>

              {/* Status dot */}
              <div
                className={cn(
                  "size-2 rounded-full shrink-0",
                  statusColors[span.status],
                )}
              />

              {/* Span name */}
              <div className="w-48 truncate text-text-primary">{span.name}</div>

              {/* Timeline bar */}
              <div className="flex-1 relative h-4">
                <div className="absolute inset-y-0 left-0 w-full bg-bg-tertiary rounded" />
                <div
                  className={cn(
                    "absolute inset-y-0 rounded transition-all",
                    statusColors[span.status],
                    span.status === "running" && "animate-pulse",
                  )}
                  style={{
                    left: "0%",
                    width: span.duration
                      ? `${Math.min((span.duration / (spans[0]?.duration || 1)) * 100, 100)}%`
                      : "20%",
                  }}
                />
              </div>

              {/* Duration */}
              <div className="w-20 text-right text-text-muted font-mono">
                {span.duration ? `${span.duration.toFixed(1)}ms` : "—"}
              </div>

              {/* Tokens */}
              <div className="w-16 text-right text-text-muted">
                {span.inputTokens && span.outputTokens
                  ? `${span.inputTokens + span.outputTokens}`
                  : "—"}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ============================================================================
// ErrorCascadeVisualizer - Visualize error propagation through the system
// ============================================================================

export interface ErrorNode {
  id: string;
  name: string;
  type: "function" | "agent" | "runtime" | "external";
  error?: string;
  children?: ErrorNode[];
}

export interface ErrorCascadeVisualizerProps {
  rootError: ErrorNode;
  onNodeClick?: (node: ErrorNode) => void;
  selectedNodeId?: string;
  className?: string;
}

export function ErrorCascadeVisualizer({
  rootError,
  onNodeClick,
  selectedNodeId,
  className,
}: ErrorCascadeVisualizerProps) {
  const renderNode = (node: ErrorNode, depth = 0): React.ReactElement => {
    const isSelected = node.id === selectedNodeId;
    const typeColors = {
      function: "border-l-brand-500",
      agent: "border-l-purple-500",
      runtime: "border-l-amber-500",
      external: "border-l-red-500",
    };

    return (
      <div key={node.id}>
        <button
          onClick={() => onNodeClick?.(node)}
          className={cn(
            "w-full flex items-center gap-3 p-3 rounded-lg border-l-4 bg-bg-secondary hover:bg-bg-hover transition-colors text-left",
            typeColors[node.type],
            isSelected && "bg-brand-500/10 border-brand-500",
          )}
        >
          <AlertTriangle
            className={cn(
              "size-4 shrink-0",
              node.error ? "text-error" : "text-text-muted",
            )}
          />
          <div className="flex-1 min-w-0">
            <div className="text-sm font-medium text-text-primary">
              {node.name}
            </div>
            {node.error && (
              <div className="text-[11px] text-error mt-0.5 line-clamp-1">
                {node.error}
              </div>
            )}
          </div>
          <span className="text-[10px] text-text-muted capitalize px-1.5 py-0.5 bg-bg-tertiary rounded">
            {node.type}
          </span>
        </button>
        {node.children && node.children.length > 0 && (
          <div className="ml-6 mt-1 space-y-1 border-l border-border-subtle pl-3">
            {node.children.map((child) => renderNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <AlertTriangle className="size-4 text-error" /> Error Cascade
      </h4>
      {renderNode(rootError)}
    </div>
  );
}

// ============================================================================
// MemoryPressureMonitor - Monitor memory pressure across runtimes
// ============================================================================

export interface MemoryPressureMonitorProps {
  data: Array<{
    runtime: string;
    usedMB: number;
    limitMB: number;
    pressure: "low" | "medium" | "high" | "critical";
  }>;
  className?: string;
}

export function MemoryPressureMonitor({
  data,
  className,
}: MemoryPressureMonitorProps) {
  const pressureColors = {
    low: "text-success bg-success/10",
    medium: "text-warning bg-warning/10",
    high: "text-orange-400 bg-orange-400/10",
    critical: "text-error bg-error/10",
  };

  const pressureBarColors = {
    low: "bg-success",
    medium: "bg-warning",
    high: "bg-orange-400",
    critical: "bg-error",
  };

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <MemoryStick className="size-4 text-brand-400" /> Memory Pressure
      </h4>
      <div className="space-y-2">
        {data.map((runtime) => {
          const usagePct = (runtime.usedMB / runtime.limitMB) * 100;
          return (
            <div
              key={runtime.runtime}
              className="p-3 bg-bg-secondary rounded-lg border border-border-subtle"
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Server className="size-4 text-text-muted" />
                  <span className="text-sm font-medium text-text-primary">
                    {runtime.runtime}
                  </span>
                </div>
                <span
                  className={cn(
                    "text-[10px] px-2 py-0.5 rounded-full capitalize",
                    pressureColors[runtime.pressure],
                  )}
                >
                  {runtime.pressure}
                </span>
              </div>
              <div className="h-2 bg-bg-tertiary rounded-full overflow-hidden mb-1">
                <div
                  className={cn(
                    "h-full rounded-full transition-all",
                    pressureBarColors[runtime.pressure],
                  )}
                  style={{ width: `${Math.min(usagePct, 100)}%` }}
                />
              </div>
              <div className="flex items-center justify-between text-[10px] text-text-muted">
                <span>{runtime.usedMB.toFixed(0)} MB used</span>
                <span>{runtime.limitMB} MB limit</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ============================================================================
// GPUUsageMonitor - Monitor GPU utilization
// ============================================================================

export interface GPUUsageMonitorProps {
  data: Array<{
    gpuId: string;
    name: string;
    utilizationPct: number;
    memoryUsedMB: number;
    memoryTotalMB: number;
    temperatureC?: number;
    powerW?: number;
  }>;
  className?: string;
}

export function GPUUsageMonitor({ data, className }: GPUUsageMonitorProps) {
  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Cpu className="size-4 text-brand-400" /> GPU Utilization
      </h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {data.map((gpu) => {
          const memPct = (gpu.memoryUsedMB / gpu.memoryTotalMB) * 100;
          return (
            <div
              key={gpu.gpuId}
              className="p-4 bg-bg-secondary rounded-lg border border-border-subtle"
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <div className="size-8 rounded bg-gradient-to-br from-brand-500/20 to-brand-500/5 flex items-center justify-center">
                    <Cpu className="size-4 text-brand-400" />
                  </div>
                  <div>
                    <div className="text-sm font-medium text-text-primary">
                      {gpu.name}
                    </div>
                    <div className="text-[10px] text-text-muted">
                      {gpu.gpuId}
                    </div>
                  </div>
                </div>
                {gpu.temperatureC && (
                  <div className="flex items-center gap-1 text-[10px]">
                    <Thermometer className="size-3" />
                    <span
                      className={
                        gpu.temperatureC > 80
                          ? "text-error"
                          : gpu.temperatureC > 60
                            ? "text-warning"
                            : "text-text-muted"
                      }
                    >
                      {gpu.temperatureC}°C
                    </span>
                  </div>
                )}
              </div>

              {/* Utilization */}
              <div className="mb-3">
                <div className="flex items-center justify-between text-[10px] mb-1">
                  <span className="text-text-muted">Utilization</span>
                  <span className="font-medium text-text-primary">
                    {gpu.utilizationPct}%
                  </span>
                </div>
                <div className="h-2 bg-bg-tertiary rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-brand-500 to-brand-400 rounded-full transition-all"
                    style={{ width: `${gpu.utilizationPct}%` }}
                  />
                </div>
              </div>

              {/* Memory */}
              <div className="mb-3">
                <div className="flex items-center justify-between text-[10px] mb-1">
                  <span className="text-text-muted">Memory</span>
                  <span className="font-medium text-text-primary">
                    {gpu.memoryUsedMB} / {gpu.memoryTotalMB} MB
                  </span>
                </div>
                <div className="h-2 bg-bg-tertiary rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-cyan-500 to-cyan-400 rounded-full transition-all"
                    style={{ width: `${memPct}%` }}
                  />
                </div>
              </div>

              {/* Power */}
              {gpu.powerW && (
                <div className="flex items-center justify-between text-[10px]">
                  <span className="text-text-muted">Power</span>
                  <span className="font-medium text-text-primary">
                    {gpu.powerW}W
                  </span>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ============================================================================
// BandwidthUsagePanel - Monitor bandwidth usage
// ============================================================================

export interface BandwidthUsagePanelProps {
  data: Array<{
    interface: string;
    rxMB: number;
    txMB: number;
    rxPackets: number;
    txPackets: number;
  }>;
  className?: string;
}

export function BandwidthUsagePanel({
  data,
  className,
}: BandwidthUsagePanelProps) {
  const totalRx = data.reduce((sum, d) => sum + d.rxMB, 0);
  const totalTx = data.reduce((sum, d) => sum + d.txMB, 0);

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Wifi className="size-4 text-brand-400" /> Bandwidth Usage
      </h4>

      {/* Summary */}
      <div className="grid grid-cols-2 gap-3">
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Download</div>
          <div className="text-lg font-bold text-success">
            ↓ {totalRx.toFixed(2)} MB
          </div>
        </div>
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Upload</div>
          <div className="text-lg font-bold text-brand-500">
            ↑ {totalTx.toFixed(2)} MB
          </div>
        </div>
      </div>

      {/* Interfaces */}
      <div className="space-y-2">
        {data.map((iface) => (
          <div
            key={iface.interface}
            className="p-3 bg-bg-secondary rounded-lg border border-border-subtle"
          >
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <HardDrive className="size-4 text-text-muted" />
                <span className="text-sm font-medium text-text-primary">
                  {iface.interface}
                </span>
              </div>
              <span className="text-[10px] text-text-muted">
                {iface.rxPackets.toLocaleString()} /{" "}
                {iface.txPackets.toLocaleString()} packets
              </span>
            </div>
            <div className="grid grid-cols-2 gap-2">
              <div>
                <div className="text-[9px] text-text-muted mb-0.5">
                  Download
                </div>
                <div className="h-1.5 bg-bg-tertiary rounded-full overflow-hidden">
                  <div
                    className="h-full bg-success rounded-full"
                    style={{
                      width: `${Math.min((iface.rxMB / Math.max(totalRx, 1)) * 100, 100)}%`,
                    }}
                  />
                </div>
                <span className="text-[10px] text-text-primary">
                  {iface.rxMB.toFixed(3)} MB
                </span>
              </div>
              <div>
                <div className="text-[9px] text-text-muted mb-0.5">Upload</div>
                <div className="h-1.5 bg-bg-tertiary rounded-full overflow-hidden">
                  <div
                    className="h-full bg-brand-500 rounded-full"
                    style={{
                      width: `${Math.min((iface.txMB / Math.max(totalTx, 1)) * 100, 100)}%`,
                    }}
                  />
                </div>
                <span className="text-[10px] text-text-primary">
                  {iface.txMB.toFixed(3)} MB
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// DistributedTracingViewer - View distributed traces
// ============================================================================

export interface TraceSpan {
  id: string;
  traceId: string;
  parentId?: string;
  serviceName: string;
  operationName: string;
  startTime: number;
  duration: number;
  status: "ok" | "error";
  tags?: Record<string, string>;
}

export interface DistributedTracingViewerProps {
  traceId: string;
  spans: TraceSpan[];
  onSpanClick?: (span: TraceSpan) => void;
  selectedSpanId?: string;
  className?: string;
}

export function DistributedTracingViewer({
  traceId,
  spans,
  onSpanClick,
  selectedSpanId,
  className,
}: DistributedTracingViewerProps) {
  // Group spans by service
  const services = React.useMemo(() => {
    const map = new Map<string, TraceSpan[]>();
    spans.forEach((span) => {
      if (!map.has(span.serviceName)) map.set(span.serviceName, []);
      map.get(span.serviceName)!.push(span);
    });
    return Array.from(map.entries());
  }, [spans]);

  const maxDuration = Math.max(...spans.map((s) => s.duration), 1);

  return (
    <div className={cn("space-y-3", className)}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <Server className="size-4 text-brand-400" /> Distributed Trace
        </h4>
        <span className="text-[10px] text-text-muted font-mono">
          {traceId.slice(0, 12)}...
        </span>
      </div>

      {/* Waterfall view */}
      <div className="space-y-1">
        {services.map(([service, serviceSpans]) => (
          <div key={service} className="space-y-0.5">
            <div className="text-[10px] font-medium text-text-muted px-2 py-1 bg-bg-tertiary rounded">
              {service} ({serviceSpans.length} spans)
            </div>
            {serviceSpans.map((span) => {
              const isSelected = span.id === selectedSpanId;
              return (
                <button
                  key={span.id}
                  onClick={() => onSpanClick?.(span)}
                  className={cn(
                    "w-full flex items-center gap-2 px-2 py-1 rounded text-[11px] transition-colors",
                    isSelected ? "bg-brand-500/20" : "hover:bg-bg-secondary",
                  )}
                >
                  <span
                    className={cn(
                      "size-2 rounded-full shrink-0",
                      span.status === "ok" ? "bg-success" : "bg-error",
                    )}
                  />
                  <span className="w-32 truncate text-left text-text-primary">
                    {span.operationName}
                  </span>
                  <div className="flex-1 h-3 bg-bg-tertiary rounded relative">
                    <div
                      className={cn(
                        "absolute inset-y-0 rounded transition-all",
                        span.status === "ok" ? "bg-brand-500" : "bg-error",
                      )}
                      style={{
                        left: "0%",
                        width: `${(span.duration / maxDuration) * 100}%`,
                      }}
                    />
                  </div>
                  <span className="w-16 text-right text-text-muted font-mono">
                    {span.duration.toFixed(1)}ms
                  </span>
                </button>
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// RequestReplayPanel - Replay historical requests
// ============================================================================

export interface ReplayRequest {
  id: string;
  timestamp: string;
  method: string;
  path: string;
  status: number;
  duration: number;
  requestSize: number;
  responseSize: number;
}

export interface RequestReplayPanelProps {
  requests: ReplayRequest[];
  onReplay?: (request: ReplayRequest) => void;
  className?: string;
}

export function RequestReplayPanel({
  requests,
  onReplay,
  className,
}: RequestReplayPanelProps) {
  const [selectedRequest, setSelectedRequest] =
    React.useState<ReplayRequest | null>(null);

  const statusColor = (status: number) => {
    if (status < 300) return "text-success";
    if (status < 400) return "text-warning";
    if (status < 500) return "text-orange-400";
    return "text-error";
  };

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <RotateCcw className="size-4 text-brand-400" /> Request Replay
      </h4>

      {/* Request list */}
      <div className="space-y-1 max-h-64 overflow-y-auto">
        {requests.map((req) => {
          const isSelected = selectedRequest?.id === req.id;
          return (
            <button
              key={req.id}
              onClick={() => setSelectedRequest(req)}
              className={cn(
                "w-full flex items-center gap-3 px-3 py-2 rounded-lg text-left transition-colors",
                isSelected
                  ? "bg-brand-500/20 border border-brand-500/30"
                  : "bg-bg-secondary border border-transparent hover:border-border-default",
              )}
            >
              <span className="text-[10px] font-mono text-text-muted w-32 truncate">
                {req.timestamp}
              </span>
              <span
                className={cn(
                  "text-[10px] font-bold w-8",
                  req.method === "GET"
                    ? "text-success"
                    : req.method === "POST"
                      ? "text-brand-400"
                      : "text-warning",
                )}
              >
                {req.method}
              </span>
              <span className="flex-1 truncate text-[11px] text-text-primary">
                {req.path}
              </span>
              <span className={statusColor(req.status)}>{req.status}</span>
              <span className="text-[10px] text-text-muted w-16 text-right">
                {req.duration}ms
              </span>
            </button>
          );
        })}
      </div>

      {/* Replay button */}
      {selectedRequest && onReplay && (
        <button
          onClick={() => onReplay(selectedRequest)}
          className="w-full py-2 text-xs bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors flex items-center justify-center gap-2"
        >
          <Play className="size-3" /> Replay Request
        </button>
      )}
    </div>
  );
}

// ============================================================================
// LiveLogConsole - Live log stream
// ============================================================================

export interface LogEntry {
  id: string;
  timestamp: number;
  level: "debug" | "info" | "warn" | "error";
  message: string;
  service?: string;
  fields?: Record<string, string>;
}

export interface LiveLogConsoleProps {
  logs: LogEntry[];
  onEntryClick?: (entry: LogEntry) => void;
  selectedEntryId?: string;
  autoScroll?: boolean;
  className?: string;
}

export function LiveLogConsole({
  logs,
  onEntryClick,
  selectedEntryId,
  autoScroll = true,
  className,
}: LiveLogConsoleProps) {
  const containerRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const levelColors = {
    debug: "text-text-dim",
    info: "text-text-secondary",
    warn: "text-warning",
    error: "text-error",
  };

  const levelIcons = {
    debug: <Minus className="size-3" />,
    info: <Mic className="size-3" />,
    warn: <AlertTriangle className="size-3" />,
    error: <XCircle className="size-3" />,
  };

  return (
    <div className={cn("space-y-2", className)}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <Volume2 className="size-4 text-brand-400" /> Live Logs
        </h4>
        <span className="text-[10px] text-text-muted">
          {logs.length} entries
        </span>
      </div>

      <div
        ref={containerRef}
        className="h-80 overflow-y-auto bg-bg-primary rounded-lg border border-border-subtle p-2 font-mono text-[11px] space-y-0.5"
      >
        {logs.map((log) => {
          const isSelected = log.id === selectedEntryId;
          return (
            <button
              key={log.id}
              onClick={() => onEntryClick?.(log)}
              className={cn(
                "w-full flex items-start gap-2 px-2 py-1 rounded text-left transition-colors",
                isSelected ? "bg-brand-500/20" : "hover:bg-bg-secondary",
              )}
            >
              <span className="text-text-dim shrink-0">
                {new Date(log.timestamp).toLocaleTimeString()}
              </span>
              <span className={cn("shrink-0", levelColors[log.level])}>
                {levelIcons[log.level]}
              </span>
              {log.service && (
                <span className="text-brand-400 shrink-0">[{log.service}]</span>
              )}
              <span className={levelColors[log.level]}>{log.message}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}

// ============================================================================
// RuntimeMetricsGrid - Grid of runtime metrics
// ============================================================================

export interface RuntimeMetricsGridProps {
  runtimes: Array<{
    id: string;
    name: string;
    status: "healthy" | "degraded" | "down";
    metrics: {
      requests: number;
      latencyP50: number;
      latencyP99: number;
      errorRate: number;
      cpu: number;
      memory: number;
    };
  }>;
  onRuntimeClick?: (runtimeId: string) => void;
  className?: string;
}

export function RuntimeMetricsGrid({
  runtimes,
  onRuntimeClick,
  className,
}: RuntimeMetricsGridProps) {
  const statusColors = {
    healthy: "border-success/30 bg-success/5",
    degraded: "border-warning/30 bg-warning/5",
    down: "border-error/30 bg-error/5",
  };

  const statusDot = {
    healthy: "bg-success",
    degraded: "bg-warning",
    down: "bg-error",
  };

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Gauge className="size-4 text-brand-400" /> Runtime Metrics
      </h4>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        {runtimes.map((runtime) => (
          <button
            key={runtime.id}
            onClick={() => onRuntimeClick?.(runtime.id)}
            className={cn(
              "p-4 rounded-lg border transition-all text-left hover:border-border-default",
              statusColors[runtime.status],
            )}
          >
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <div
                  className={cn(
                    "size-2 rounded-full",
                    statusDot[runtime.status],
                  )}
                />
                <span className="text-sm font-medium text-text-primary">
                  {runtime.name}
                </span>
              </div>
              <span className="text-[10px] text-text-muted capitalize">
                {runtime.status}
              </span>
            </div>
            <div className="grid grid-cols-2 gap-2 text-[10px]">
              <div>
                <span className="text-text-muted">Requests</span>
                <div className="text-sm font-bold text-text-primary">
                  {runtime.metrics.requests.toLocaleString()}
                </div>
              </div>
              <div>
                <span className="text-text-muted">Error Rate</span>
                <div
                  className={cn(
                    "text-sm font-bold",
                    runtime.metrics.errorRate < 1
                      ? "text-success"
                      : runtime.metrics.errorRate < 5
                        ? "text-warning"
                        : "text-error",
                  )}
                >
                  {runtime.metrics.errorRate.toFixed(2)}%
                </div>
              </div>
              <div>
                <span className="text-text-muted">p50 Latency</span>
                <div className="text-sm font-bold text-text-primary">
                  {runtime.metrics.latencyP50}ms
                </div>
              </div>
              <div>
                <span className="text-text-muted">p99 Latency</span>
                <div className="text-sm font-bold text-text-primary">
                  {runtime.metrics.latencyP99}ms
                </div>
              </div>
              <div>
                <span className="text-text-muted">CPU</span>
                <div className="text-sm font-bold text-text-primary">
                  {runtime.metrics.cpu.toFixed(0)}%
                </div>
              </div>
              <div>
                <span className="text-text-muted">Memory</span>
                <div className="text-sm font-bold text-text-primary">
                  {runtime.metrics.memory.toFixed(0)}%
                </div>
              </div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// ObservabilityRadar - Radar chart for system health
// ============================================================================

export interface ObservabilityRadarProps {
  metrics: Array<{
    label: string;
    value: number;
    max: number;
  }>;
  className?: string;
}

export function ObservabilityRadar({
  metrics,
  className,
}: ObservabilityRadarProps) {
  const size = 200;
  const center = size / 2;
  const maxRadius = size / 2 - 20;

  const getPoint = (
    value: number,
    max: number,
    index: number,
    total: number,
  ) => {
    const angle = (index / total) * 2 * Math.PI - Math.PI / 2;
    const radius = (value / max) * maxRadius;
    return {
      x: center + radius * Math.cos(angle),
      y: center + radius * Math.sin(angle),
    };
  };

  const getLabelPoint = (index: number, total: number) => {
    const angle = (index / total) * 2 * Math.PI - Math.PI / 2;
    const radius = maxRadius + 15;
    return {
      x: center + radius * Math.cos(angle),
      y: center + radius * Math.sin(angle),
    };
  };

  // Generate grid circles
  const gridLevels = [0.25, 0.5, 0.75, 1];

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Activity className="size-4 text-brand-400" /> Health Radar
      </h4>
      <div className="flex items-center justify-center">
        <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
          {/* Grid circles */}
          {gridLevels.map((level) => (
            <circle
              key={level}
              cx={center}
              cy={center}
              r={maxRadius * level}
              fill="none"
              stroke="rgba(255,255,255,0.1)"
              strokeWidth={1}
            />
          ))}

          {/* Axis lines */}
          {metrics.map((_, i) => {
            const labelPt = getLabelPoint(i, metrics.length);
            return (
              <line
                key={i}
                x1={center}
                y1={center}
                x2={labelPt.x}
                y2={labelPt.y}
                stroke="rgba(255,255,255,0.1)"
                strokeWidth={1}
              />
            );
          })}

          {/* Data polygon */}
          <polygon
            points={metrics
              .map((m, i) => {
                const pt = getPoint(m.value, m.max, i, metrics.length);
                return `${pt.x},${pt.y}`;
              })
              .join(" ")}
            fill="rgba(255, 107, 53, 0.2)"
            stroke="#ff6b35"
            strokeWidth={2}
          />

          {/* Labels */}
          {metrics.map((m, i) => {
            const labelPt = getLabelPoint(i, metrics.length);
            return (
              <text
                key={i}
                x={labelPt.x}
                y={labelPt.y}
                textAnchor="middle"
                dominantBaseline="middle"
                className="text-[9px] fill-text-muted"
              >
                {m.label}
              </text>
            );
          })}
        </svg>
      </div>
    </div>
  );
}

// ============================================================================
// IncidentTimeline - Timeline of incidents
// ============================================================================

export interface Incident {
  id: string;
  severity: "critical" | "high" | "medium" | "low";
  title: string;
  description?: string;
  startTime: number;
  endTime?: number;
  status: "active" | "resolved";
  affectedServices?: string[];
}

export interface IncidentTimelineProps {
  incidents: Incident[];
  onIncidentClick?: (incident: Incident) => void;
  selectedIncidentId?: string;
  className?: string;
}

export function IncidentTimeline({
  incidents,
  onIncidentClick,
  selectedIncidentId,
  className,
}: IncidentTimelineProps) {
  const severityColors = {
    critical: "bg-error text-white",
    high: "bg-orange-500 text-white",
    medium: "bg-warning text-black",
    low: "bg-blue-500 text-white",
  };

  const sortedIncidents = [...incidents].sort(
    (a, b) => b.startTime - a.startTime,
  );

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <AlertTriangle className="size-4 text-brand-400" /> Incident Timeline
      </h4>
      <div className="relative">
        {/* Timeline line */}
        <div className="absolute left-4 top-0 bottom-0 w-px bg-border-subtle" />

        {/* Incidents */}
        <div className="space-y-3">
          {sortedIncidents.map((incident) => {
            const isSelected = incident.id === selectedIncidentId;
            return (
              <button
                key={incident.id}
                onClick={() => onIncidentClick?.(incident)}
                className={cn(
                  "w-full flex items-start gap-3 p-3 rounded-lg border transition-all text-left",
                  isSelected
                    ? "bg-brand-500/10 border-brand-500/30"
                    : "bg-bg-secondary border-border-subtle hover:border-border-default",
                )}
              >
                {/* Severity dot */}
                <div
                  className={cn(
                    "size-8 rounded-full flex items-center justify-center shrink-0 z-10",
                    severityColors[incident.severity],
                  )}
                >
                  {incident.severity === "critical" ? (
                    <XCircle className="size-4" />
                  ) : (
                    <AlertTriangle className="size-4" />
                  )}
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm font-medium text-text-primary">
                      {incident.title}
                    </span>
                    <span
                      className={cn(
                        "text-[10px] px-1.5 py-0.5 rounded capitalize",
                        incident.status === "active"
                          ? "bg-error/20 text-error"
                          : "bg-success/20 text-success",
                      )}
                    >
                      {incident.status}
                    </span>
                  </div>
                  {incident.description && (
                    <p className="text-[11px] text-text-secondary line-clamp-2 mb-2">
                      {incident.description}
                    </p>
                  )}
                  <div className="flex items-center gap-3 text-[10px] text-text-muted">
                    <span>{new Date(incident.startTime).toLocaleString()}</span>
                    {incident.endTime && (
                      <span>
                        → {new Date(incident.endTime).toLocaleString()}
                      </span>
                    )}
                    {incident.affectedServices &&
                      incident.affectedServices.length > 0 && (
                        <span>{incident.affectedServices.join(", ")}</span>
                      )}
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
}

// ============================================================================
// AnomalyDetectionPanel - Anomaly detection alerts
// ============================================================================

export interface Anomaly {
  id: string;
  metric: string;
  description: string;
  detectedAt: number;
  severity: "critical" | "high" | "medium" | "low";
  value: number;
  expectedRange: [number, number];
  confidence: number;
}

export interface AnomalyDetectionPanelProps {
  anomalies: Anomaly[];
  onAnomalyClick?: (anomaly: Anomaly) => void;
  className?: string;
}

export function AnomalyDetectionPanel({
  anomalies,
  onAnomalyClick,
  className,
}: AnomalyDetectionPanelProps) {
  const severityColors = {
    critical: "border-error/30 bg-error/5",
    high: "border-orange-500/30 bg-orange-500/5",
    medium: "border-warning/30 bg-warning/5",
    low: "border-blue-500/30 bg-blue-500/5",
  };

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <AlertTriangle className="size-4 text-warning" /> Anomaly Detection
      </h4>
      <div className="space-y-2">
        {anomalies.map((anomaly) => (
          <button
            key={anomaly.id}
            onClick={() => onAnomalyClick?.(anomaly)}
            className={cn(
              "w-full p-3 rounded-lg border transition-all text-left",
              severityColors[anomaly.severity],
            )}
          >
            <div className="flex items-start justify-between mb-1">
              <div>
                <span className="text-sm font-medium text-text-primary">
                  {anomaly.metric}
                </span>
                <span className="text-[10px] text-text-muted ml-2">
                  #{anomaly.id.slice(0, 6)}
                </span>
              </div>
              <span className="text-[10px] text-text-muted">
                {new Date(anomaly.detectedAt).toLocaleString()}
              </span>
            </div>
            <p className="text-[11px] text-text-secondary mb-2">
              {anomaly.description}
            </p>
            <div className="flex items-center gap-4 text-[10px]">
              <span className="text-text-muted">
                Value:{" "}
                <span className="text-error font-medium">
                  {anomaly.value.toFixed(2)}
                </span>
              </span>
              <span className="text-text-muted">
                Expected: {anomaly.expectedRange[0].toFixed(2)} -{" "}
                {anomaly.expectedRange[1].toFixed(2)}
              </span>
              <span className="text-text-muted">
                Confidence: {(anomaly.confidence * 100).toFixed(0)}%
              </span>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// CostPredictionPanel - AI-powered cost prediction
// ============================================================================

export interface CostPredictionPanelProps {
  currentCost: number;
  predictedCost: number;
  trend: "increasing" | "stable" | "decreasing";
  predictionBasis: string;
  projections: Array<{
    period: string;
    cost: number;
    confidence: number;
  }>;
  className?: string;
}

export function CostPredictionPanel({
  currentCost,
  predictedCost,
  trend,
  predictionBasis,
  projections,
  className,
}: CostPredictionPanelProps) {
  const trendIcon = {
    increasing: <TrendingUp className="size-4 text-error" />,
    stable: <RefreshCcw className="size-4 text-text-muted" />,
    decreasing: <TrendingDown className="size-4 text-success" />,
  };

  const trendColor = {
    increasing: "text-error",
    stable: "text-text-muted",
    decreasing: "text-success",
  };

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <DollarSign className="size-4 text-brand-400" /> Cost Prediction
      </h4>

      {/* Current vs Predicted */}
      <div className="grid grid-cols-2 gap-3">
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">Current (30d)</div>
          <div className="text-lg font-bold text-text-primary">
            ${currentCost.toFixed(2)}
          </div>
        </div>
        <div className="p-3 bg-bg-secondary rounded-lg border border-border-subtle">
          <div className="text-[10px] text-text-muted mb-1">
            Predicted (30d)
          </div>
          <div className="text-lg font-bold text-brand-500 flex items-center gap-1">
            ${predictedCost.toFixed(2)}
            <span className={trendColor[trend]}>{trendIcon[trend]}</span>
          </div>
        </div>
      </div>

      {/* Basis */}
      <div className="text-[10px] text-text-muted p-2 bg-bg-secondary rounded">
        <span className="font-medium">Prediction basis:</span> {predictionBasis}
      </div>

      {/* Projections chart */}
      <div>
        <div className="text-[10px] text-text-muted mb-2">Projected Costs</div>
        <div className="flex items-end gap-2 h-24">
          {projections.map((proj, i) => {
            const maxCost = Math.max(...projections.map((p) => p.cost), 1);
            return (
              <div
                key={proj.period}
                className="flex-1 flex flex-col items-center gap-1"
              >
                <div
                  className="w-full rounded-t bg-brand-500/40 hover:bg-brand-500/60 transition-colors"
                  style={{ height: `${(proj.cost / maxCost) * 100}%` }}
                  title={`$${proj.cost.toFixed(2)} (${(proj.confidence * 100).toFixed(0)}% confidence)`}
                />
                <span className="text-[9px] text-text-muted">
                  {proj.period}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

// ============================================================================
// ResourceForecastGraph - Resource usage forecast
// ============================================================================

export interface ResourceForecastPoint {
  timestamp: number;
  actual?: number;
  predicted?: number;
  lower?: number;
  upper?: number;
}

export interface ResourceForecastGraphProps {
  metric: string;
  unit: string;
  data: ResourceForecastPoint[];
  className?: string;
}

export function ResourceForecastGraph({
  metric,
  unit,
  data,
  className,
}: ResourceForecastGraphProps) {
  const allValues = data.flatMap(
    (d) =>
      [d.actual, d.predicted, d.lower, d.upper].filter(
        (v) => v != null,
      ) as number[],
  );
  const maxValue = Math.max(...allValues, 1);
  const minValue = Math.min(...allValues, 0);

  const getY = (value: number) =>
    100 - ((value - minValue) / (maxValue - minValue)) * 100;

  const actualPoints = data
    .filter((d) => d.actual != null)
    .map((d, i) => ({
      x: (i / (data.length - 1)) * 100,
      y: getY(d.actual!),
    }));

  const predictedPoints = data
    .filter((d) => d.predicted != null)
    .map((d, i) => ({
      x: (i / (data.length - 1)) * 100,
      y: getY(d.predicted!),
    }));

  const upperBand = data
    .filter((d) => d.upper != null)
    .map((d, i) => `${(i / (data.length - 1)) * 100},${getY(d.upper!)}`);
  const lowerBand = data
    .filter((d) => d.lower != null)
    .map((d, i) => `${(i / (data.length - 1)) * 100},${getY(d.lower!)}`)
    .reverse();

  const separatorIndex = data.findIndex(
    (d) => d.predicted != null && d.actual == null,
  );

  return (
    <div className={cn("space-y-2", className)}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary">{metric}</h4>
        <span className="text-[10px] text-text-muted">{unit}</span>
      </div>
      <svg
        width="100%"
        height={120}
        viewBox="0 0 100 120"
        preserveAspectRatio="none"
        className="rounded-lg bg-bg-secondary border border-border-subtle"
      >
        <defs>
          <linearGradient id="forecast-gradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#ff6b35" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#ff6b35" stopOpacity="0" />
          </linearGradient>
        </defs>

        {/* Confidence band */}
        {upperBand.length > 0 && lowerBand.length > 0 && (
          <polygon
            points={[...upperBand, ...lowerBand].join(" ")}
            fill="url(#forecast-gradient)"
          />
        )}

        {/* Separator line */}
        {separatorIndex > 0 && (
          <line
            x1={(separatorIndex / (data.length - 1)) * 100}
            y1="0"
            x2={(separatorIndex / (data.length - 1)) * 100}
            y2="100"
            stroke="rgba(255,255,255,0.3)"
            strokeWidth={1}
            strokeDasharray="4,2"
          />
        )}

        {/* Actual line */}
        {actualPoints.length > 1 && (
          <polyline
            points={actualPoints.map((p) => `${p.x},${p.y}`).join(" ")}
            fill="none"
            stroke="#10b981"
            strokeWidth={1.5}
          />
        )}

        {/* Predicted line */}
        {predictedPoints.length > 1 && (
          <polyline
            points={predictedPoints.map((p) => `${p.x},${p.y}`).join(" ")}
            fill="none"
            stroke="#ff6b35"
            strokeWidth={1.5}
            strokeDasharray="4,2"
          />
        )}
      </svg>
      <div className="flex items-center justify-between text-[10px] text-text-muted">
        <span>Past</span>
        <div className="flex items-center gap-4">
          <span className="flex items-center gap-1">
            <span className="size-2 rounded-full bg-success" /> Actual
          </span>
          <span className="flex items-center gap-1">
            <span className="size-2 rounded-full bg-brand-500" /> Predicted
          </span>
        </div>
        <span>Future</span>
      </div>
    </div>
  );
}
