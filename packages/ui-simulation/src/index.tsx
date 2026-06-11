/**
 * @functionfly/ui-simulation
 * Runtime Simulation Components - R-Sim system for predictive analysis
 */

import React, {
  useState,
  useCallback,
  useMemo,
  useRef,
  useEffect,
} from "react";
import { cn } from "@functionfly/ui-core";
import {
  Play,
  Pause,
  Square,
  RotateCcw,
  Gauge,
  TrendingUp,
  TrendingDown,
  Activity,
  AlertTriangle,
  AlertCircle,
  CheckCircle,
  XCircle,
  Brain,
  Zap,
  Clock,
  Target,
  BarChart3,
  PieChart,
  LineChart,
  AreaChart,
  Hexagon,
  Circle,
  Box,
  Database,
  Cpu,
  Server,
  Network,
  GitFork,
  GitBranch,
  GitMerge,
  Users,
  Bot,
  MessageSquare,
  FileText,
  Timer,
  FastForward,
  Rewind,
  SkipForward,
  SkipBack,
  ChevronRight,
  ChevronDown,
  ChevronLeft,
  ChevronUp,
  Plus,
  Minus,
  X,
  Check,
  Info,
  Eye,
  EyeOff,
  Filter,
  Search,
  RefreshCw,
  Download,
  Upload,
  Settings,
  MoreHorizontal,
  Layers,
  ArrowRight,
  ArrowUpRight,
  ArrowDownRight,
  ArrowLeft,
  Thermometer,
  Flame,
  Cloud,
  Droplet,
  Wind,
  Radio,
  Antenna,
  Workflow,
  GitCommit,
  GitPullRequest,
  LayersIcon,
  ScatterChart,
  BarChart,
  PieChartIcon,
  CircleDot,
  Triangle,
  SquareDot,
  Diamond,
  Hexagon as HexagonIcon,
} from "lucide-react";

// ============================================================================
// Simulation Control Center
// ============================================================================

interface SimulationConfig {
  id: string;
  name: string;
  type: "load" | "stress" | "chaos" | "regression" | "capacity";
  duration: number;
  warmupDuration?: number;
  cooldownDuration?: number;
  parallelism?: number;
  rampUpTime?: number;
}

interface SimulationMetrics {
  requestsTotal?: number;
  requestsSuccess?: number;
  requestsFailed?: number;
  avgLatency?: number;
  p50Latency?: number;
  p95Latency?: number;
  p99Latency?: number;
  maxLatency?: number;
  throughput?: number;
  errorRate?: number;
  timestamp?: number;
}

interface SimulationControlCenterProps {
  config: SimulationConfig | null;
  status: "idle" | "preparing" | "running" | "paused" | "completed" | "failed";
  metrics: SimulationMetrics | null;
  onStart?: () => void;
  onPause?: () => void;
  onStop?: () => void;
  onReset?: () => void;
  onConfigChange?: (config: SimulationConfig) => void;
  className?: string;
}

export const SimulationControlCenter: React.FC<
  SimulationControlCenterProps
> = ({
  config,
  status,
  metrics,
  onStart,
  onPause,
  onStop,
  onReset,
  onConfigChange,
  className,
}) => {
  const [elapsedTime, setElapsedTime] = useState(0);

  useEffect(() => {
    let interval: number;
    if (status === "running") {
      interval = setInterval(() => setElapsedTime((t) => t + 1), 1000);
    } else if (status === "idle") {
      setElapsedTime(0);
    }
    return () => clearInterval(interval);
  }, [status]);

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  };

  const getStatusColor = () => {
    switch (status) {
      case "running":
        return "text-green-400";
      case "preparing":
        return "text-amber-400";
      case "paused":
        return "text-aviation-amber";
      case "completed":
        return "text-aviation-cyan";
      case "failed":
        return "text-red-400";
      default:
        return "text-aviation-text-muted";
    }
  };

  const getProgressPercent = () => {
    if (!config || !metrics?.requestsTotal) return 0;
    return Math.min(
      100,
      (metrics.requestsTotal / (config.duration * 10)) * 100,
    );
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Gauge className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Simulation Control
            </h3>
          </div>
          <div
            className={cn(
              "flex items-center gap-1.5 text-xs font-medium uppercase",
              getStatusColor(),
            )}
          >
            <span
              className={cn("w-2 h-2 rounded-full", {
                "bg-green-400 animate-pulse": status === "running",
                "bg-amber-400": status === "preparing" || status === "paused",
                "bg-aviation-cyan": status === "completed",
                "bg-red-400": status === "failed",
                "bg-aviation-text-muted": status === "idle",
              })}
            />
            {status}
          </div>
        </div>
      </div>

      {/* Config Display */}
      {config && (
        <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="grid grid-cols-2 gap-4 text-xs">
            <div>
              <span className="text-aviation-text-dim">Name:</span>
              <span className="ml-2 text-aviation-text-primary font-medium">
                {config.name}
              </span>
            </div>
            <div>
              <span className="text-aviation-text-dim">Type:</span>
              <span className="ml-2 px-1.5 py-0.5 bg-aviation-cyan/20 text-aviation-cyan rounded uppercase text-[10px]">
                {config.type}
              </span>
            </div>
            <div>
              <span className="text-aviation-text-dim">Duration:</span>
              <span className="ml-2 text-aviation-text-primary">
                {config.duration}s
              </span>
            </div>
            <div>
              <span className="text-aviation-text-dim">Parallelism:</span>
              <span className="ml-2 text-aviation-text-primary">
                {config.parallelism || 1}
              </span>
            </div>
          </div>
        </div>
      )}

      {/* Timer Display */}
      <div className="flex items-center justify-center py-6 border-b border-aviation-border-panel">
        <div className="text-4xl font-mono font-bold text-aviation-text-primary">
          {formatTime(elapsedTime)}
        </div>
        <div className="ml-4 text-xs text-aviation-text-dim">
          / {formatTime(config?.duration || 0)}
        </div>
      </div>

      {/* Progress Bar */}
      <div className="px-4 py-2 border-b border-aviation-border-panel">
        <div className="h-2 bg-aviation-bg-instrument rounded-full overflow-hidden">
          <div
            className={cn(
              "h-full transition-all duration-300",
              status === "running"
                ? "bg-aviation-cyan"
                : status === "completed"
                  ? "bg-green-400"
                  : status === "failed"
                    ? "bg-red-400"
                    : "bg-aviation-text-dim",
            )}
            style={{ width: `${getProgressPercent()}%` }}
          />
        </div>
        <div className="flex justify-between mt-1 text-[10px] text-aviation-text-dim">
          <span>{metrics?.requestsTotal || 0} requests</span>
          <span>{getProgressPercent().toFixed(0)}%</span>
        </div>
      </div>

      {/* Metrics Grid */}
      {metrics && (
        <div className="flex-1 overflow-auto p-4">
          <div className="grid grid-cols-3 gap-3">
            <div className="p-3 bg-aviation-bg-secondary rounded-lg">
              <div className="flex items-center gap-1.5 text-[10px] text-aviation-text-dim mb-1">
                <TrendingUp className="w-3 h-3" />
                Throughput
              </div>
              <div className="text-lg font-bold text-aviation-text-primary">
                {metrics.throughput?.toFixed(1) || "0"}
              </div>
              <div className="text-[10px] text-aviation-text-dim">req/s</div>
            </div>

            <div className="p-3 bg-aviation-bg-secondary rounded-lg">
              <div className="flex items-center gap-1.5 text-[10px] text-aviation-text-dim mb-1">
                <Clock className="w-3 h-3" />
                Avg Latency
              </div>
              <div className="text-lg font-bold text-aviation-text-primary">
                {metrics.avgLatency?.toFixed(0) || "0"}
              </div>
              <div className="text-[10px] text-aviation-text-dim">ms</div>
            </div>

            <div className="p-3 bg-aviation-bg-secondary rounded-lg">
              <div className="flex items-center gap-1.5 text-[10px] text-aviation-text-dim mb-1">
                <AlertTriangle className="w-3 h-3" />
                Error Rate
              </div>
              <div
                className={cn(
                  "text-lg font-bold",
                  (metrics.errorRate || 0) > 5
                    ? "text-red-400"
                    : "text-aviation-text-primary",
                )}
              >
                {(metrics.errorRate || 0).toFixed(2)}%
              </div>
              <div className="text-[10px] text-aviation-text-dim">of total</div>
            </div>

            <div className="p-3 bg-aviation-bg-secondary rounded-lg">
              <div className="flex items-center gap-1.5 text-[10px] text-aviation-text-dim mb-1">
                <CheckCircle className="w-3 h-3 text-green-400" />
                Success
              </div>
              <div className="text-lg font-bold text-green-400">
                {metrics.requestsSuccess || 0}
              </div>
              <div className="text-[10px] text-aviation-text-dim">
                completed
              </div>
            </div>

            <div className="p-3 bg-aviation-bg-secondary rounded-lg">
              <div className="flex items-center gap-1.5 text-[10px] text-aviation-text-dim mb-1">
                <XCircle className="w-3 h-3 text-red-400" />
                Failed
              </div>
              <div className="text-lg font-bold text-red-400">
                {metrics.requestsFailed || 0}
              </div>
              <div className="text-[10px] text-aviation-text-dim">errors</div>
            </div>

            <div className="p-3 bg-aviation-bg-secondary rounded-lg">
              <div className="flex items-center gap-1.5 text-[10px] text-aviation-text-dim mb-1">
                <Activity className="w-3 h-3" />
                P95 Latency
              </div>
              <div className="text-lg font-bold text-aviation-text-primary">
                {metrics.p95Latency?.toFixed(0) || "0"}
              </div>
              <div className="text-[10px] text-aviation-text-dim">ms</div>
            </div>
          </div>
        </div>
      )}

      {/* Control Buttons */}
      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center justify-center gap-3">
          {status === "idle" && (
            <button
              onClick={onStart}
              className="flex items-center gap-2 px-4 py-2 bg-aviation-cyan text-aviation-bg-primary rounded-lg hover:bg-aviation-cyan/90 transition-colors"
            >
              <Play className="w-4 h-4" />
              Start Simulation
            </button>
          )}

          {status === "running" && (
            <>
              <button
                onClick={onPause}
                className="flex items-center gap-2 px-4 py-2 bg-aviation-amber text-aviation-bg-primary rounded-lg hover:bg-aviation-amber/90 transition-colors"
              >
                <Pause className="w-4 h-4" />
                Pause
              </button>
              <button
                onClick={onStop}
                className="flex items-center gap-2 px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
              >
                <Square className="w-4 h-4" />
                Stop
              </button>
            </>
          )}

          {status === "paused" && (
            <>
              <button
                onClick={onStart}
                className="flex items-center gap-2 px-4 py-2 bg-aviation-cyan text-aviation-bg-primary rounded-lg hover:bg-aviation-cyan/90 transition-colors"
              >
                <Play className="w-4 h-4" />
                Resume
              </button>
              <button
                onClick={onStop}
                className="flex items-center gap-2 px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
              >
                <Square className="w-4 h-4" />
                Stop
              </button>
            </>
          )}

          {(status === "completed" || status === "failed") && (
            <button
              onClick={onReset}
              className="flex items-center gap-2 px-4 py-2 bg-aviation-bg-instrument text-aviation-text-primary rounded-lg hover:bg-aviation-bg-panel transition-colors"
            >
              <RotateCcw className="w-4 h-4" />
              Reset
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Execution Forecast Panel
// ============================================================================

interface ForecastDataPoint {
  timestamp: number;
  expectedLatency: number;
  confidence: number;
  upperBound: number;
  lowerBound: number;
}

interface ExecutionForecastPanelProps {
  forecasts: ForecastDataPoint[];
  horizon?: number;
  selectedPointId?: string | null;
  onPointSelect?: (point: ForecastDataPoint) => void;
  onRefresh?: () => void;
  className?: string;
}

export const ExecutionForecastPanel: React.FC<ExecutionForecastPanelProps> = ({
  forecasts,
  horizon = 60,
  selectedPointId,
  onPointSelect,
  onRefresh,
  className,
}) => {
  const [hoveredPoint, setHoveredPoint] = useState<number | null>(null);

  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const maxLatency = useMemo(() => {
    return Math.max(...forecasts.map((f) => f.upperBound), 100);
  }, [forecasts]);

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <TrendingUp className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Execution Forecast
            </h3>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-aviation-text-dim">
              {forecasts.length} points
            </span>
            <button
              onClick={onRefresh}
              className="p-1.5 hover:bg-aviation-bg-instrument rounded transition-colors"
            >
              <RefreshCw className="w-4 h-4 text-aviation-text-muted" />
            </button>
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center gap-4 text-[10px]">
          <div className="flex items-center gap-1.5">
            <div className="w-3 h-0.5 bg-aviation-cyan" />
            <span className="text-aviation-text-muted">Expected</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-3 h-0.5 bg-aviation-cyan/30" />
            <span className="text-aviation-text-muted">Confidence Band</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-2 h-2 rounded-full bg-amber-400" />
            <span className="text-aviation-text-muted">Selected</span>
          </div>
        </div>
      </div>

      {/* Chart Area */}
      <div className="flex-1 overflow-hidden p-4">
        <div className="relative h-full">
          {/* Y-axis labels */}
          <div className="absolute left-0 top-0 bottom-0 w-12 flex flex-col justify-between text-[10px] text-aviation-text-dim">
            <span>{maxLatency.toFixed(0)}ms</span>
            <span>{(maxLatency * 0.75).toFixed(0)}ms</span>
            <span>{(maxLatency * 0.5).toFixed(0)}ms</span>
            <span>{(maxLatency * 0.25).toFixed(0)}ms</span>
            <span>0ms</span>
          </div>

          {/* Chart SVG */}
          <svg
            className="absolute left-14 right-0 top-0 bottom-4"
            viewBox="0 0 100 100"
            preserveAspectRatio="none"
          >
            {/* Grid lines */}
            {[0, 25, 50, 75, 100].map((y) => (
              <line
                key={y}
                x1="0"
                y1={y}
                x2="100"
                y2={y}
                className="stroke-aviation-border-panel"
                strokeWidth="0.5"
                strokeDasharray="2 2"
              />
            ))}

            {/* Confidence band */}
            <path
              d={
                forecasts
                  .map((f, i) => {
                    const x = (i / (forecasts.length - 1)) * 100;
                    const upperY = 100 - (f.upperBound / maxLatency) * 100;
                    return `${i === 0 ? "M" : "L"} ${x} ${upperY}`;
                  })
                  .join(" ") +
                " " +
                forecasts
                  .map((f, i) => {
                    const x = (i / (forecasts.length - 1)) * 100;
                    const lowerY = 100 - (f.lowerBound / maxLatency) * 100;
                    return `L ${x} ${lowerY}`;
                  })
                  .reverse()
                  .join(" ")
              }
              className="fill-aviation-cyan/10"
            />

            {/* Expected line */}
            <path
              d={forecasts
                .map((f, i) => {
                  const x = (i / (forecasts.length - 1)) * 100;
                  const y = 100 - (f.expectedLatency / maxLatency) * 100;
                  return `${i === 0 ? "M" : "L"} ${x} ${y}`;
                })
                .join(" ")}
              className="fill-none stroke-aviation-cyan"
              strokeWidth="2"
            />

            {/* Confidence dots */}
            {forecasts
              .filter((_, i) => i % 5 === 0)
              .map((f, i) => {
                const x = ((i * 5) / (forecasts.length - 1)) * 100;
                const y = 100 - (f.expectedLatency / maxLatency) * 100;
                return (
                  <g key={i}>
                    <circle
                      cx={x}
                      cy={y}
                      r="2"
                      className="fill-aviation-cyan"
                    />
                    <text
                      x={x}
                      y={y + 8}
                      textAnchor="middle"
                      className="text-[6px] fill-aviation-text-dim"
                    >
                      {Math.round(f.confidence * 100)}%
                    </text>
                  </g>
                );
              })}
          </svg>

          {/* X-axis labels */}
          <div className="absolute left-14 right-0 bottom-0 h-4 flex justify-between text-[10px] text-aviation-text-dim">
            <span>{formatTime(forecasts[0]?.timestamp || Date.now())}</span>
            <span>
              {formatTime(
                forecasts[Math.floor(forecasts.length / 2)]?.timestamp ||
                  Date.now(),
              )}
            </span>
            <span>
              {formatTime(
                forecasts[forecasts.length - 1]?.timestamp || Date.now(),
              )}
            </span>
          </div>
        </div>
      </div>

      {/* Selected Point Details */}
      {hoveredPoint !== null && forecasts[hoveredPoint] && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between">
            <div>
              <span className="text-xs text-aviation-text-muted">
                Expected:{" "}
              </span>
              <span className="text-sm font-medium text-aviation-text-primary">
                {forecasts[hoveredPoint].expectedLatency.toFixed(1)}ms
              </span>
            </div>
            <div className="flex items-center gap-3 text-[10px] text-aviation-text-dim">
              <span>
                Range: {forecasts[hoveredPoint].lowerBound.toFixed(0)}-
                {forecasts[hoveredPoint].upperBound.toFixed(0)}ms
              </span>
              <span>
                Confidence:{" "}
                {Math.round(forecasts[hoveredPoint].confidence * 100)}%
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Failure Probability Map
// ============================================================================

interface FailureNode {
  id: string;
  name: string;
  type: "service" | "endpoint" | "database" | "cache" | "queue" | "worker";
  failureProbability: number;
  historicalRate?: number;
  affectedRequests?: number;
  correlationId?: string;
}

interface FailureProbabilityMapProps {
  nodes: FailureNode[];
  selectedNodeId?: string | null;
  threshold?: number;
  onNodeSelect?: (node: FailureNode) => void;
  onNodeHover?: (node: FailureNode | null) => void;
  className?: string;
}

export const FailureProbabilityMap: React.FC<FailureProbabilityMapProps> = ({
  nodes,
  selectedNodeId = null,
  threshold = 0.3,
  onNodeSelect,
  onNodeHover,
  className,
}) => {
  const [hoveredNode, setHoveredNode] = useState<string | null>(null);

  const getNodeColor = (probability: number) => {
    if (probability > 0.7)
      return {
        bg: "bg-red-500/20",
        border: "border-red-500",
        text: "text-red-400",
      };
    if (probability > 0.5)
      return {
        bg: "bg-amber-500/20",
        border: "border-amber-500",
        text: "text-amber-400",
      };
    if (probability > threshold)
      return {
        bg: "bg-yellow-500/20",
        border: "border-yellow-500",
        text: "text-yellow-400",
      };
    return {
      bg: "bg-green-500/20",
      border: "border-green-500",
      text: "text-green-400",
    };
  };

  const getNodeIcon = (type: string) => {
    switch (type) {
      case "service":
        return <Server className="w-4 h-4" />;
      case "endpoint":
        return <Network className="w-4 h-4" />;
      case "database":
        return <Database className="w-4 h-4" />;
      case "cache":
        return <Cpu className="w-4 h-4" />;
      case "queue":
        return <Layers className="w-4 h-4" />;
      case "worker":
        return <Bot className="w-4 h-4" />;
      default:
        return <Box className="w-4 h-4" />;
    }
  };

  // Layout nodes in a force-directed-like arrangement
  const nodePositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {};
    const centerX = 200;
    const centerY = 150;

    nodes.forEach((node, index) => {
      const angle = (2 * Math.PI * index) / nodes.length;
      const radius = 80 + node.failureProbability * 60;
      positions[node.id] = {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle),
      };
    });
    return positions;
  }, [nodes]);

  return (
    <div
      className={cn(
        "relative h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="absolute top-3 left-3 flex items-center gap-2 z-10">
        <div className="flex items-center gap-1.5 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel">
          <AlertTriangle className="w-4 h-4 text-red-400" />
          <span className="text-xs text-aviation-text-primary font-medium">
            Failure Probability
          </span>
        </div>
      </div>

      {/* Threshold indicator */}
      <div className="absolute top-3 right-3 flex items-center gap-2 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel z-10">
        <span className="text-[10px] text-aviation-text-dim">Threshold:</span>
        <span className="text-xs text-amber-400">
          {(threshold * 100).toFixed(0)}%
        </span>
      </div>

      <svg className="w-full h-full" viewBox="0 0 400 300">
        {/* Connection lines to show correlations */}
        {nodes
          .filter((n) => n.correlationId)
          .map((node) => {
            const related = nodes.find((n) => n.id === node.correlationId);
            if (!related) return null;
            const sourcePos = nodePositions[node.id];
            const targetPos = nodePositions[related.id];
            if (!sourcePos || !targetPos) return null;

            return (
              <line
                x1={sourcePos.x}
                y1={sourcePos.y}
                x2={targetPos.x}
                y2={targetPos.y}
                className="stroke-red-500/30"
                strokeWidth={1}
                strokeDasharray="4 4"
              />
            );
          })}

        {/* Nodes */}
        {nodes.map((node) => {
          const pos = nodePositions[node.id];
          if (!pos) return null;

          const isSelected = selectedNodeId === node.id;
          const isHovered = hoveredNode === node.id;
          const colors = getNodeColor(node.failureProbability);
          const size = 30 + node.failureProbability * 20;

          return (
            <g
              key={node.id}
              onClick={() => onNodeSelect?.(node)}
              onMouseEnter={() => {
                setHoveredNode(node.id);
                onNodeHover?.(node);
              }}
              onMouseLeave={() => {
                setHoveredNode(null);
                onNodeHover?.(null);
              }}
              className="cursor-pointer"
            >
              {/* Outer ring for critical nodes */}
              {node.failureProbability > 0.7 && (
                <circle
                  cx={pos.x}
                  cy={pos.y}
                  r={size + 8}
                  className="fill-none stroke-red-500 animate-pulse"
                  strokeWidth={2}
                />
              )}

              {/* Main node */}
              <circle
                cx={pos.x}
                cy={pos.y}
                r={size}
                className={cn(
                  "fill-aviation-bg-panel stroke-2 transition-all",
                  colors.border,
                )}
                strokeWidth={isSelected ? 3 : 1}
              />

              {/* Icon */}
              <g transform={`translate(${pos.x - 10}, ${pos.y - 10})`}>
                <span className={colors.text}>{getNodeIcon(node.type)}</span>
              </g>

              {/* Probability badge */}
              <g transform={`translate(${pos.x - 12}, ${pos.y + size + 4})`}>
                <rect
                  x="0"
                  y="0"
                  width="24"
                  height="12"
                  rx="2"
                  className={colors.bg}
                />
                <text
                  x="12"
                  y="9"
                  textAnchor="middle"
                  className={cn("text-[9px] font-bold", colors.text)}
                >
                  {(node.failureProbability * 100).toFixed(0)}%
                </text>
              </g>

              {/* Label */}
              <text
                x={pos.x}
                y={pos.y - size - 8}
                textAnchor="middle"
                className="text-[10px] fill-aviation-text-primary"
              >
                {node.name.length > 12
                  ? node.name.slice(0, 12) + "..."
                  : node.name}
              </text>
            </g>
          );
        })}
      </svg>

      {/* Hovered Node Details */}
      {hoveredNode && nodes.find((n) => n.id === hoveredNode) && (
        <div className="absolute bottom-3 left-3 right-3 p-3 bg-aviation-bg-secondary/90 rounded-lg border border-aviation-border-panel backdrop-blur-sm z-10">
          <div className="flex items-start justify-between">
            <div>
              <h4 className="text-sm font-medium text-aviation-text-primary">
                {nodes.find((n) => n.id === hoveredNode)?.name}
              </h4>
              <div className="flex items-center gap-2 mt-1 text-[10px] text-aviation-text-dim">
                <span className="uppercase">
                  {nodes.find((n) => n.id === hoveredNode)?.type}
                </span>
                {nodes.find((n) => n.id === hoveredNode)?.affectedRequests && (
                  <span>
                    •{" "}
                    {nodes.find((n) => n.id === hoveredNode)?.affectedRequests}{" "}
                    req affected
                  </span>
                )}
              </div>
            </div>
            <div
              className={cn(
                "text-lg font-bold",
                getNodeColor(
                  nodes.find((n) => n.id === hoveredNode)!.failureProbability,
                ).text,
              )}
            >
              {(
                nodes.find((n) => n.id === hoveredNode)!.failureProbability *
                100
              ).toFixed(1)}
              %
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Latency Prediction Graph
// ============================================================================

interface LatencyDataPoint {
  timestamp: number;
  predicted: number;
  actual?: number;
  p50: number;
  p95: number;
  p99: number;
}

interface LatencyPredictionGraphProps {
  data: LatencyDataPoint[];
  selectedTimestamp?: number | null;
  showActual?: boolean;
  showConfidence?: boolean;
  onDataPointSelect?: (point: LatencyDataPoint) => void;
  className?: string;
}

export const LatencyPredictionGraph: React.FC<LatencyPredictionGraphProps> = ({
  data,
  selectedTimestamp,
  showActual = true,
  showConfidence = true,
  onDataPointSelect,
  className,
}) => {
  const [hoveredPoint, setHoveredPoint] = useState<number | null>(null);

  const maxLatency = useMemo(() => {
    return Math.max(
      ...data.map((d) => Math.max(d.predicted, d.p99, d.actual || 0)),
      100,
    );
  }, [data]);

  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Timer className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Latency Prediction
            </h3>
          </div>
          <div className="flex items-center gap-3 text-[10px]">
            <span className="flex items-center gap-1">
              <div className="w-3 h-0.5 bg-aviation-cyan" />
              <span className="text-aviation-text-muted">Predicted</span>
            </span>
            {showActual && (
              <span className="flex items-center gap-1">
                <div className="w-3 h-0.5 bg-green-400" />
                <span className="text-aviation-text-muted">Actual</span>
              </span>
            )}
            {showConfidence && (
              <span className="flex items-center gap-1">
                <div className="w-3 h-0.5 bg-purple-500/50" />
                <span className="text-aviation-text-muted">P95-P99</span>
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Chart */}
      <div className="flex-1 overflow-hidden p-4">
        <svg
          className="w-full h-full"
          viewBox="0 0 100 100"
          preserveAspectRatio="none"
        >
          {/* Grid */}
          {[0, 25, 50, 75, 100].map((y) => (
            <line
              key={y}
              x1="0"
              y1={y}
              x2="100"
              y2={y}
              className="stroke-aviation-border-panel"
              strokeWidth="0.5"
              strokeDasharray="2 2"
            />
          ))}

          {/* Confidence band (P95-P99) */}
          {showConfidence && (
            <path
              d={data
                .map((d, i) => {
                  const x = (i / (data.length - 1)) * 100;
                  const y95 = 100 - (d.p95 / maxLatency) * 100;
                  const y99 = 100 - (d.p99 / maxLatency) * 100;
                  return `${i === 0 ? "M" : "L"} ${x} ${y95} ${i === 0 ? "" : ""} ${x} ${y99}`;
                })
                .join(" ")}
              className="fill-purple-500/10"
            />
          )}

          {/* Predicted line */}
          <path
            d={data
              .map((d, i) => {
                const x = (i / (data.length - 1)) * 100;
                const y = 100 - (d.predicted / maxLatency) * 100;
                return `${i === 0 ? "M" : "L"} ${x} ${y}`;
              })
              .join(" ")}
            className="fill-none stroke-aviation-cyan"
            strokeWidth="2"
          />

          {/* Actual line */}
          {showActual && (
            <path
              d={data
                .filter((d) => d.actual !== undefined)
                .map((d, i) => {
                  const actualIndex = data.indexOf(d);
                  const x = (actualIndex / (data.length - 1)) * 100;
                  const y = 100 - ((d.actual || 0) / maxLatency) * 100;
                  return `${i === 0 ? "M" : "L"} ${x} ${y}`;
                })
                .join(" ")}
              className="fill-none stroke-green-400"
              strokeWidth="2"
              strokeDasharray="4 2"
            />
          )}

          {/* P50 markers */}
          {data
            .filter((_, i) => i % 10 === 0)
            .map((d, i) => {
              const x = ((i * 10) / (data.length - 1)) * 100;
              const y = 100 - (d.p50 / maxLatency) * 100;
              return (
                <circle
                  key={i}
                  cx={x}
                  cy={y}
                  r="1.5"
                  className="fill-aviation-amber"
                />
              );
            })}
        </svg>
      </div>

      {/* Stats */}
      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="grid grid-cols-4 gap-3 text-xs">
          <div>
            <span className="text-aviation-text-dim">Current: </span>
            <span className="font-medium text-aviation-text-primary">
              {data[data.length - 1]?.predicted.toFixed(1)}ms
            </span>
          </div>
          <div>
            <span className="text-aviation-text-dim">P50: </span>
            <span className="font-medium text-aviation-text-primary">
              {data[data.length - 1]?.p50.toFixed(1)}ms
            </span>
          </div>
          <div>
            <span className="text-aviation-text-dim">P95: </span>
            <span className="font-medium text-aviation-text-primary">
              {data[data.length - 1]?.p95.toFixed(1)}ms
            </span>
          </div>
          <div>
            <span className="text-aviation-text-dim">P99: </span>
            <span className="font-medium text-aviation-text-primary">
              {data[data.length - 1]?.p99.toFixed(1)}ms
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Hallucination Risk Analyzer
// ============================================================================

interface HallucinationRisk {
  id: string;
  source: string;
  type: "context" | "training" | "reasoning" | "retrieval";
  severity: "critical" | "high" | "medium" | "low";
  confidence: number;
  description: string;
  indicators?: string[];
  mitigationSuggestion?: string;
}

interface HallucinationRiskAnalyzerProps {
  risks: HallucinationRisk[];
  selectedRiskId?: string | null;
  threshold?: number;
  onRiskSelect?: (risk: HallucinationRisk) => void;
  onRiskDismiss?: (riskId: string) => void;
  className?: string;
}

export const HallucinationRiskAnalyzer: React.FC<
  HallucinationRiskAnalyzerProps
> = ({
  risks,
  selectedRiskId = null,
  threshold = 0.5,
  onRiskSelect,
  onRiskDismiss,
  className,
}) => {
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return {
          bg: "bg-red-500/20",
          border: "border-red-500",
          text: "text-red-400",
          badge: "bg-red-500",
        };
      case "high":
        return {
          bg: "bg-amber-500/20",
          border: "border-amber-500",
          text: "text-amber-400",
          badge: "bg-amber-500",
        };
      case "medium":
        return {
          bg: "bg-yellow-500/20",
          border: "border-yellow-500",
          text: "text-yellow-400",
          badge: "bg-yellow-500",
        };
      default:
        return {
          bg: "bg-green-500/20",
          border: "border-green-500",
          text: "text-green-400",
          badge: "bg-green-500",
        };
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case "context":
        return <Brain className="w-4 h-4" />;
      case "training":
        return <FileText className="w-4 h-4" />;
      case "reasoning":
        return <GitFork className="w-4 h-4" />;
      case "retrieval":
        return <Database className="w-4 h-4" />;
      default:
        return <AlertTriangle className="w-4 h-4" />;
    }
  };

  const filteredRisks = risks.filter((r) => r.confidence >= threshold);
  const criticalCount = filteredRisks.filter(
    (r) => r.severity === "critical",
  ).length;

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-5 h-5 text-red-400" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Hallucination Risk
            </h3>
          </div>
          {criticalCount > 0 && (
            <div className="flex items-center gap-1.5 px-2 py-1 bg-red-500/20 rounded border border-red-500/50">
              <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
              <span className="text-xs text-red-400 font-medium">
                {criticalCount} Critical
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Risk List */}
      <div className="flex-1 overflow-y-auto">
        {filteredRisks.map((risk) => {
          const colors = getSeverityColor(risk.severity);
          const isSelected = selectedRiskId === risk.id;

          return (
            <div
              key={risk.id}
              onClick={() => onRiskSelect?.(risk)}
              className={cn(
                "p-4 border-b border-aviation-border-panel cursor-pointer transition-colors",
                isSelected
                  ? "bg-aviation-bg-instrument"
                  : "hover:bg-aviation-bg-secondary",
              )}
            >
              <div className="flex items-start justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className={colors.text}>{getTypeIcon(risk.type)}</span>
                  <span className="text-xs text-aviation-text-muted uppercase">
                    {risk.type}
                  </span>
                  <span
                    className={cn(
                      "px-1.5 py-0.5 text-[10px] rounded uppercase text-white",
                      colors.badge,
                    )}
                  >
                    {risk.severity}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span className={cn("text-sm font-bold", colors.text)}>
                    {Math.round(risk.confidence * 100)}%
                  </span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onRiskDismiss?.(risk.id);
                    }}
                    className="p-1 hover:bg-aviation-bg-panel rounded"
                  >
                    <X className="w-3 h-3 text-aviation-text-muted" />
                  </button>
                </div>
              </div>

              <p className="text-sm text-aviation-text-primary mb-2">
                {risk.description}
              </p>

              <div className="flex items-center justify-between">
                <span className="text-[10px] text-aviation-text-dim">
                  Source: {risk.source}
                </span>
                {risk.mitigationSuggestion && (
                  <span className="text-[10px] text-aviation-cyan">
                    Click for mitigation
                  </span>
                )}
              </div>

              {risk.indicators && risk.indicators.length > 0 && (
                <div className="flex items-center gap-1 mt-2">
                  {risk.indicators.slice(0, 3).map((indicator, i) => (
                    <span
                      key={i}
                      className="px-1.5 py-0.5 bg-aviation-bg-secondary rounded text-[10px] text-aviation-text-dim"
                    >
                      {indicator}
                    </span>
                  ))}
                </div>
              )}
            </div>
          );
        })}

        {filteredRisks.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <CheckCircle className="w-8 h-8 mb-2 text-green-400" />
            <p className="text-sm">No hallucination risks detected</p>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Cost Simulation Chart
// ============================================================================

interface CostProjection {
  timestamp: number;
  computeCost: number;
  memoryCost: number;
  networkCost: number;
  storageCost: number;
  totalCost: number;
  cumulativeCost: number;
}

interface CostSimulationChartProps {
  projections: CostProjection[];
  selectedPointId?: string | null;
  showBreakdown?: boolean;
  comparisonBaseline?: CostProjection[];
  onPointSelect?: (point: CostProjection) => void;
  className?: string;
}

export const CostSimulationChart: React.FC<CostSimulationChartProps> = ({
  projections,
  selectedPointId,
  showBreakdown = false,
  comparisonBaseline,
  onPointSelect,
  className,
}) => {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  const maxCost = useMemo(() => {
    return Math.max(...projections.map((p) => p.totalCost), 10);
  }, [projections]);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(value);
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const latestPoint = projections[projections.length - 1];

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <DollarSign className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Cost Simulation
            </h3>
          </div>
          <div className="flex items-center gap-3 text-xs">
            <span className="flex items-center gap-1">
              <div className="w-3 h-3 rounded bg-aviation-cyan" />
              <span className="text-aviation-text-muted">Compute</span>
            </span>
            {showBreakdown && (
              <>
                <span className="flex items-center gap-1">
                  <div className="w-3 h-3 rounded bg-purple-500" />
                  <span className="text-aviation-text-muted">Memory</span>
                </span>
                <span className="flex items-center gap-1">
                  <div className="w-3 h-3 rounded bg-amber-500" />
                  <span className="text-aviation-text-muted">Network</span>
                </span>
                <span className="flex items-center gap-1">
                  <div className="w-3 h-3 rounded bg-green-500" />
                  <span className="text-aviation-text-muted">Storage</span>
                </span>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Chart */}
      <div className="flex-1 overflow-hidden p-4">
        <svg
          className="w-full h-full"
          viewBox="0 0 100 100"
          preserveAspectRatio="none"
        >
          {/* Grid */}
          {[0, 25, 50, 75, 100].map((y) => (
            <line
              key={y}
              x1="0"
              y1={y}
              x2="100"
              y2={y}
              className="stroke-aviation-border-panel"
              strokeWidth="0.5"
              strokeDasharray="2 2"
            />
          ))}

          {/* Cost area */}
          <path
            d={
              projections
                .map((p, i) => {
                  const x = (i / (projections.length - 1)) * 100;
                  const y = 100 - (p.totalCost / maxCost) * 100;
                  return `${i === 0 ? "M" : "L"} ${x} ${y}`;
                })
                .join(" ") + " L 100 100 L 0 100 Z"
            }
            className="fill-aviation-cyan/20"
          />

          {/* Cost line */}
          <path
            d={projections
              .map((p, i) => {
                const x = (i / (projections.length - 1)) * 100;
                const y = 100 - (p.totalCost / maxCost) * 100;
                return `${i === 0 ? "M" : "L"} ${x} ${y}`;
              })
              .join(" ")}
            className="fill-none stroke-aviation-cyan"
            strokeWidth="2"
          />

          {/* Breakdown layers */}
          {showBreakdown &&
            projections.map((p, i) => {
              const x = (i / (projections.length - 1)) * 100;
              const computeY = 100 - (p.computeCost / maxCost) * 100;
              const memoryY = computeY + (p.memoryCost / maxCost) * 100;
              const networkY = memoryY + (p.networkCost / maxCost) * 100;

              return (
                <g key={i}>
                  <line
                    x1={x}
                    y1={computeY}
                    x2={x}
                    y2={memoryY}
                    className="stroke-purple-500"
                    strokeWidth="2"
                  />
                  <line
                    x1={x}
                    y1={memoryY}
                    x2={x}
                    y2={networkY}
                    className="stroke-amber-500"
                    strokeWidth="2"
                  />
                </g>
              );
            })}
        </svg>
      </div>

      {/* Stats */}
      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="grid grid-cols-3 gap-4 text-xs">
          <div>
            <span className="text-aviation-text-dim">Current: </span>
            <span className="font-medium text-aviation-text-primary">
              {formatCurrency(latestPoint?.totalCost || 0)}/hr
            </span>
          </div>
          <div>
            <span className="text-aviation-text-dim">Compute: </span>
            <span className="font-medium text-aviation-text-primary">
              {formatCurrency(latestPoint?.computeCost || 0)}
            </span>
          </div>
          <div>
            <span className="text-aviation-text-dim">Cumulative: </span>
            <span className="font-medium text-aviation-cyan">
              {formatCurrency(latestPoint?.cumulativeCost || 0)}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Stress Test Runner
// ============================================================================

interface StressTestResult {
  id: string;
  timestamp: number;
  duration: number;
  peakLoad: number;
  steadyStateLoad?: number;
  successRate: number;
  avgResponseTime: number;
  maxResponseTime: number;
  errors: Array<{ type: string; count: number }>;
  bottlenecks?: string[];
}

interface StressTestRunnerProps {
  results: StressTestResult[];
  currentTestId?: string | null;
  isRunning?: boolean;
  onTestStart?: (config: SimulationConfig) => void;
  onTestStop?: () => void;
  onTestSelect?: (result: StressTestResult) => void;
  className?: string;
}

export const StressTestRunner: React.FC<StressTestRunnerProps> = ({
  results,
  currentTestId,
  isRunning = false,
  onTestStart,
  onTestStop,
  onTestSelect,
  className,
}) => {
  const [selectedConfig, setSelectedConfig] = useState<SimulationConfig>({
    id: "default",
    name: "Load Test",
    type: "load",
    duration: 60,
    parallelism: 10,
  });

  const getSuccessRateColor = (rate: number) => {
    if (rate > 95) return "text-green-400";
    if (rate > 80) return "text-amber-400";
    return "text-red-400";
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}m ${secs}s`;
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Flame className="w-5 h-5 text-red-400" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Stress Test Runner
            </h3>
          </div>
          {isRunning && (
            <div className="flex items-center gap-1.5 px-2 py-1 bg-red-500/20 rounded border border-red-500/50">
              <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
              <span className="text-xs text-red-400">Running...</span>
            </div>
          )}
        </div>
      </div>

      {/* Config Panel */}
      <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="grid grid-cols-4 gap-3 text-xs">
          <div>
            <label className="text-aviation-text-dim block mb-1">Type</label>
            <select
              value={selectedConfig.type}
              onChange={(e) =>
                setSelectedConfig({
                  ...selectedConfig,
                  type: e.target.value as any,
                })
              }
              disabled={isRunning}
              className="w-full px-2 py-1 bg-aviation-bg-instrument border border-aviation-border-panel rounded text-aviation-text-primary"
            >
              <option value="load">Load</option>
              <option value="stress">Stress</option>
              <option value="chaos">Chaos</option>
              <option value="capacity">Capacity</option>
            </select>
          </div>
          <div>
            <label className="text-aviation-text-dim block mb-1">
              Duration (s)
            </label>
            <input
              type="number"
              value={selectedConfig.duration}
              onChange={(e) =>
                setSelectedConfig({
                  ...selectedConfig,
                  duration: parseInt(e.target.value) || 60,
                })
              }
              disabled={isRunning}
              className="w-full px-2 py-1 bg-aviation-bg-instrument border border-aviation-border-panel rounded text-aviation-text-primary"
            />
          </div>
          <div>
            <label className="text-aviation-text-dim block mb-1">
              Parallelism
            </label>
            <input
              type="number"
              value={selectedConfig.parallelism || 1}
              onChange={(e) =>
                setSelectedConfig({
                  ...selectedConfig,
                  parallelism: parseInt(e.target.value) || 1,
                })
              }
              disabled={isRunning}
              className="w-full px-2 py-1 bg-aviation-bg-instrument border border-aviation-border-panel rounded text-aviation-text-primary"
            />
          </div>
          <div>
            <label className="text-aviation-text-dim block mb-1">
              Ramp Up (s)
            </label>
            <input
              type="number"
              value={selectedConfig.rampUpTime || 0}
              onChange={(e) =>
                setSelectedConfig({
                  ...selectedConfig,
                  rampUpTime: parseInt(e.target.value) || 0,
                })
              }
              disabled={isRunning}
              className="w-full px-2 py-1 bg-aviation-bg-instrument border border-aviation-border-panel rounded text-aviation-text-primary"
            />
          </div>
        </div>
      </div>

      {/* Results List */}
      <div className="flex-1 overflow-y-auto">
        {results.map((result) => {
          const isSelected = currentTestId === result.id;

          return (
            <div
              key={result.id}
              onClick={() => onTestSelect?.(result)}
              className={cn(
                "p-4 border-b border-aviation-border-panel cursor-pointer transition-colors",
                isSelected
                  ? "bg-aviation-bg-instrument border-l-2 border-l-aviation-cyan"
                  : "hover:bg-aviation-bg-secondary",
              )}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-aviation-text-primary">
                    {formatDuration(result.duration)}
                  </span>
                  <span className="text-[10px] text-aviation-text-dim">
                    Peak: {result.peakLoad} req/s
                  </span>
                </div>
                <div
                  className={cn(
                    "text-sm font-bold",
                    getSuccessRateColor(result.successRate),
                  )}
                >
                  {result.successRate.toFixed(1)}%
                </div>
              </div>

              <div className="grid grid-cols-2 gap-2 text-[10px] text-aviation-text-dim">
                <span>Avg: {result.avgResponseTime.toFixed(0)}ms</span>
                <span>Max: {result.maxResponseTime.toFixed(0)}ms</span>
              </div>

              {result.errors.length > 0 && (
                <div className="flex items-center gap-1 mt-2">
                  <AlertTriangle className="w-3 h-3 text-red-400" />
                  <span className="text-[10px] text-red-400">
                    {result.errors.length} error types
                  </span>
                </div>
              )}

              {result.bottlenecks && result.bottlenecks.length > 0 && (
                <div className="flex items-center gap-1 mt-1">
                  <Gauge className="w-3 h-3 text-amber-400" />
                  <span className="text-[10px] text-amber-400">
                    {result.bottlenecks.length} bottlenecks
                  </span>
                </div>
              )}
            </div>
          );
        })}

        {results.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <Flame className="w-8 h-8 mb-2 opacity-50" />
            <p className="text-sm">No stress tests run yet</p>
          </div>
        )}
      </div>

      {/* Control Button */}
      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <button
          onClick={() =>
            isRunning ? onTestStop?.() : onTestStart?.(selectedConfig)
          }
          className={cn(
            "w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg transition-colors",
            isRunning
              ? "bg-red-500 text-white hover:bg-red-600"
              : "bg-aviation-cyan text-aviation-bg-primary hover:bg-aviation-cyan/90",
          )}
        >
          {isRunning ? (
            <>
              <Square className="w-4 h-4" />
              Stop Test
            </>
          ) : (
            <>
              <Play className="w-4 h-4" />
              Run Stress Test
            </>
          )}
        </button>
      </div>
    </div>
  );
};

// ============================================================================
// Scaling Forecast Map
// ============================================================================

interface ScalingProjection {
  timestamp: number;
  currentReplicas: number;
  predictedReplicas: number;
  confidence: number;
  trigger: "cpu" | "memory" | "requests" | "queue_depth" | "custom";
  estimatedCostPerHour: number;
}

interface ScalingForecastMapProps {
  projections: ScalingProjection[];
  selectedTimestamp?: number | null;
  onProjectionSelect?: (projection: ScalingProjection) => void;
  onProjectionHover?: (projection: ScalingProjection | null) => void;
  className?: string;
}

export const ScalingForecastMap: React.FC<ScalingForecastMapProps> = ({
  projections,
  selectedTimestamp,
  onProjectionSelect,
  onProjectionHover,
  className,
}) => {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  const maxReplicas = useMemo(() => {
    return Math.max(
      ...projections.map((p) =>
        Math.max(p.currentReplicas, p.predictedReplicas),
      ),
      10,
    );
  }, [projections]);

  const getTriggerIcon = (trigger: string) => {
    switch (trigger) {
      case "cpu":
        return <Cpu className="w-3 h-3" />;
      case "memory":
        return <Database className="w-3 h-3" />;
      case "requests":
        return <Activity className="w-3 h-3" />;
      case "queue_depth":
        return <Layers className="w-3 h-3" />;
      default:
        return <Target className="w-3 h-3" />;
    }
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <TrendingUp className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Scaling Forecast
            </h3>
          </div>
          <div className="flex items-center gap-3 text-[10px]">
            <span className="flex items-center gap-1">
              <div className="w-3 h-0.5 bg-aviation-text-muted" />
              <span className="text-aviation-text-muted">Current</span>
            </span>
            <span className="flex items-center gap-1">
              <div className="w-3 h-0.5 bg-aviation-cyan" />
              <span className="text-aviation-text-muted">Predicted</span>
            </span>
          </div>
        </div>
      </div>

      {/* Chart */}
      <div className="flex-1 overflow-hidden p-4">
        <svg
          className="w-full h-full"
          viewBox="0 0 100 100"
          preserveAspectRatio="none"
        >
          {/* Grid */}
          {[0, 25, 50, 75, 100].map((y) => (
            <line
              key={y}
              x1="0"
              y1={y}
              x2="100"
              y2={y}
              className="stroke-aviation-border-panel"
              strokeWidth="0.5"
              strokeDasharray="2 2"
            />
          ))}

          {/* Current replicas line */}
          <path
            d={projections
              .map((p, i) => {
                const x = (i / (projections.length - 1)) * 100;
                const y = 100 - (p.currentReplicas / maxReplicas) * 100;
                return `${i === 0 ? "M" : "L"} ${x} ${y}`;
              })
              .join(" ")}
            className="fill-none stroke-aviation-text-muted"
            strokeWidth="1"
            strokeDasharray="4 2"
          />

          {/* Predicted replicas line */}
          <path
            d={projections
              .map((p, i) => {
                const x = (i / (projections.length - 1)) * 100;
                const y = 100 - (p.predictedReplicas / maxReplicas) * 100;
                return `${i === 0 ? "M" : "L"} ${x} ${y}`;
              })
              .join(" ")}
            className="fill-none stroke-aviation-cyan"
            strokeWidth="2"
          />

          {/* Trigger markers */}
          {projections
            .filter((_, i) => i % 5 === 0)
            .map((p, i) => {
              const x = ((i * 5) / (projections.length - 1)) * 100;
              const y = 100 - (p.predictedReplicas / maxReplicas) * 100;
              return (
                <g key={i}>
                  <circle cx={x} cy={y} r="2" className="fill-amber-400" />
                  <text
                    x={x}
                    y={y - 4}
                    textAnchor="middle"
                    className="text-[6px] fill-amber-400"
                  >
                    {getTriggerIcon(p.trigger)}
                  </text>
                </g>
              );
            })}
        </svg>
      </div>

      {/* Projection Details */}
      {hoveredIndex !== null && projections[hoveredIndex] && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3 text-xs">
              <span className="text-aviation-text-dim">
                {formatTime(projections[hoveredIndex].timestamp)}
              </span>
              <span className="flex items-center gap-1">
                {getTriggerIcon(projections[hoveredIndex].trigger)}
                <span className="text-aviation-text-muted uppercase">
                  {projections[hoveredIndex].trigger}
                </span>
              </span>
            </div>
            <div className="flex items-center gap-3 text-xs">
              <span>
                <span className="text-aviation-text-dim">Replicas: </span>
                <span className="font-medium text-aviation-text-primary">
                  {projections[hoveredIndex].currentReplicas} →{" "}
                  {projections[hoveredIndex].predictedReplicas}
                </span>
              </span>
              <span>
                <span className="text-aviation-text-dim">Cost: </span>
                <span className="font-medium text-aviation-cyan">
                  ${projections[hoveredIndex].estimatedCostPerHour.toFixed(2)}
                  /hr
                </span>
              </span>
              <span className="text-aviation-text-dim">
                {Math.round(projections[hoveredIndex].confidence * 100)}% conf
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Agent Behavior Predictor
// ============================================================================

interface BehaviorPrediction {
  agentId: string;
  agentName: string;
  predictedActions: Array<{
    action: string;
    probability: number;
    expectedOutcome: string;
    confidence: number;
    timestamp: number;
  }>;
  riskScore?: number;
  recommendedInterventions?: string[];
}

interface AgentBehaviorPredictorProps {
  predictions: BehaviorPrediction[];
  selectedAgentId?: string | null;
  timeHorizon?: number;
  onAgentSelect?: (prediction: BehaviorPrediction) => void;
  onInterventionApply?: (agentId: string, intervention: string) => void;
  className?: string;
}

export const AgentBehaviorPredictor: React.FC<AgentBehaviorPredictorProps> = ({
  predictions,
  selectedAgentId,
  timeHorizon = 60,
  onAgentSelect,
  onInterventionApply,
  className,
}) => {
  const selectedPrediction = predictions.find(
    (p) => p.agentId === selectedAgentId,
  );

  const getRiskColor = (score?: number) => {
    if (!score) return "text-aviation-text-muted";
    if (score > 0.8) return "text-red-400";
    if (score > 0.5) return "text-amber-400";
    return "text-green-400";
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bot className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Agent Behavior
            </h3>
          </div>
          <span className="text-xs text-aviation-text-dim">
            {timeHorizon}s horizon
          </span>
        </div>
      </div>

      {/* Agent List */}
      <div className="flex-1 flex">
        {/* Agent List */}
        <div className="w-1/3 border-r border-aviation-border-panel overflow-y-auto">
          {predictions.map((prediction) => {
            const isSelected = selectedAgentId === prediction.agentId;

            return (
              <div
                key={prediction.agentId}
                onClick={() => onAgentSelect?.(prediction)}
                className={cn(
                  "p-3 border-b border-aviation-border-panel cursor-pointer transition-colors",
                  isSelected
                    ? "bg-aviation-bg-instrument"
                    : "hover:bg-aviation-bg-secondary",
                )}
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium text-aviation-text-primary">
                    {prediction.agentName}
                  </span>
                  {prediction.riskScore !== undefined && (
                    <span
                      className={cn(
                        "text-xs font-bold",
                        getRiskColor(prediction.riskScore),
                      )}
                    >
                      {(prediction.riskScore * 100).toFixed(0)}%
                    </span>
                  )}
                </div>
                <span className="text-[10px] text-aviation-text-dim">
                  {prediction.predictedActions.length} predicted actions
                </span>
              </div>
            );
          })}
        </div>

        {/* Prediction Details */}
        <div className="flex-1 overflow-y-auto p-4">
          {selectedPrediction ? (
            <div className="space-y-3">
              {selectedPrediction.predictedActions.map((action, i) => (
                <div
                  key={i}
                  className="p-3 bg-aviation-bg-secondary rounded-lg"
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium text-aviation-text-primary">
                      {action.action}
                    </span>
                    <span className="text-xs text-aviation-cyan">
                      {Math.round(action.probability * 100)}%
                    </span>
                  </div>
                  <p className="text-xs text-aviation-text-dim mb-2">
                    Expected: {action.expectedOutcome}
                  </p>
                  <div className="flex items-center justify-between">
                    <span className="text-[10px] text-aviation-text-dim">
                      Confidence: {Math.round(action.confidence * 100)}%
                    </span>
                    <div
                      className="h-1 bg-aviation-bg-instrument rounded-full overflow-hidden"
                      style={{ width: 60 }}
                    >
                      <div
                        className="h-full bg-aviation-cyan"
                        style={{ width: `${action.confidence * 100}%` }}
                      />
                    </div>
                  </div>
                </div>
              ))}

              {selectedPrediction.recommendedInterventions &&
                selectedPrediction.recommendedInterventions.length > 0 && (
                  <div className="mt-4">
                    <h4 className="text-xs font-medium text-aviation-text-muted mb-2">
                      Recommended Interventions
                    </h4>
                    {selectedPrediction.recommendedInterventions.map(
                      (intervention, i) => (
                        <button
                          key={i}
                          onClick={() =>
                            onInterventionApply?.(
                              selectedPrediction.agentId,
                              intervention,
                            )
                          }
                          className="w-full flex items-center gap-2 p-2 mt-1 bg-amber-500/10 border border-amber-500/30 rounded text-xs text-amber-400 hover:bg-amber-500/20 transition-colors"
                        >
                          <Zap className="w-3 h-3" />
                          {intervention}
                        </button>
                      ),
                    )}
                  </div>
                )}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center h-full text-aviation-text-muted">
              <Bot className="w-8 h-8 mb-2 opacity-50" />
              <p className="text-sm">Select an agent to view predictions</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Workflow Outcome Simulator
// ============================================================================

interface WorkflowPath {
  id: string;
  name: string;
  probability: number;
  steps: Array<{
    name: string;
    duration: number;
    successProbability: number;
    alternativePath?: string;
  }>;
  totalDuration: number;
  expectedCost: number;
}

interface WorkflowOutcomeSimulatorProps {
  workflowId: string;
  paths: WorkflowPath[];
  selectedPathId?: string | null;
  onPathSelect?: (path: WorkflowPath) => void;
  onSimulationRun?: () => void;
  className?: string;
}

export const WorkflowOutcomeSimulator: React.FC<
  WorkflowOutcomeSimulatorProps
> = ({
  workflowId,
  paths,
  selectedPathId,
  onPathSelect,
  onSimulationRun,
  className,
}) => {
  const selectedPath = paths.find((p) => p.id === selectedPathId);

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  };

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Workflow className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Workflow Outcomes
            </h3>
          </div>
          <button
            onClick={onSimulationRun}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-aviation-cyan/10 text-aviation-cyan text-xs rounded hover:bg-aviation-cyan/20 transition-colors"
          >
            <Play className="w-3 h-3" />
            Simulate
          </button>
        </div>
      </div>

      {/* Path List */}
      <div className="flex-1 flex">
        {/* Paths */}
        <div className="w-2/5 border-r border-aviation-border-panel overflow-y-auto">
          {paths.map((path) => {
            const isSelected = selectedPathId === path.id;

            return (
              <div
                key={path.id}
                onClick={() => onPathSelect?.(path)}
                className={cn(
                  "p-4 border-b border-aviation-border-panel cursor-pointer transition-colors",
                  isSelected
                    ? "bg-aviation-bg-instrument"
                    : "hover:bg-aviation-bg-secondary",
                )}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-aviation-text-primary">
                    {path.name}
                  </span>
                  <span className="text-xs text-aviation-cyan">
                    {Math.round(path.probability * 100)}%
                  </span>
                </div>

                {/* Probability bar */}
                <div className="h-1.5 bg-aviation-bg-instrument rounded-full overflow-hidden mb-2">
                  <div
                    className="h-full bg-aviation-cyan transition-all"
                    style={{ width: `${path.probability * 100}%` }}
                  />
                </div>

                <div className="flex items-center justify-between text-[10px] text-aviation-text-dim">
                  <span>{path.steps.length} steps</span>
                  <span>{formatDuration(path.totalDuration)}</span>
                  <span>${path.expectedCost.toFixed(2)}</span>
                </div>
              </div>
            );
          })}
        </div>

        {/* Path Details */}
        <div className="flex-1 overflow-y-auto p-4">
          {selectedPath ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h4 className="text-sm font-medium text-aviation-text-primary">
                  {selectedPath.name}
                </h4>
                <div className="flex items-center gap-2 text-xs">
                  <span className="text-aviation-text-dim">Duration:</span>
                  <span className="text-aviation-text-primary">
                    {formatDuration(selectedPath.totalDuration)}
                  </span>
                </div>
              </div>

              {/* Step timeline */}
              <div className="relative">
                <div className="absolute left-4 top-0 bottom-0 w-px bg-aviation-border-panel" />
                <div className="space-y-4">
                  {selectedPath.steps.map((step, i) => (
                    <div key={i} className="relative pl-10">
                      {/* Timeline dot */}
                      <div
                        className={cn(
                          "absolute left-2.5 top-2 w-3 h-3 rounded-full border-2",
                          step.successProbability > 0.9
                            ? "bg-green-400 border-green-400"
                            : step.successProbability > 0.7
                              ? "bg-amber-400 border-amber-400"
                              : "bg-red-400 border-red-400",
                        )}
                      />

                      {/* Step card */}
                      <div className="p-3 bg-aviation-bg-secondary rounded-lg">
                        <div className="flex items-center justify-between mb-1">
                          <span className="text-sm font-medium text-aviation-text-primary">
                            {step.name}
                          </span>
                          <span className="text-xs text-aviation-text-dim">
                            {formatDuration(step.duration)}
                          </span>
                        </div>

                        <div className="flex items-center justify-between">
                          <span className="text-[10px] text-aviation-text-dim">
                            Success: {Math.round(step.successProbability * 100)}
                            %
                          </span>
                          {step.alternativePath && (
                            <span className="text-[10px] text-amber-400">
                              Alt: {step.alternativePath}
                            </span>
                          )}
                        </div>

                        {/* Success probability bar */}
                        <div className="mt-2 h-1 bg-aviation-bg-instrument rounded-full overflow-hidden">
                          <div
                            className={cn(
                              "h-full",
                              step.successProbability > 0.9
                                ? "bg-green-400"
                                : step.successProbability > 0.7
                                  ? "bg-amber-400"
                                  : "bg-red-400",
                            )}
                            style={{
                              width: `${step.successProbability * 100}%`,
                            }}
                          />
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Summary */}
              <div className="p-3 bg-aviation-bg-secondary rounded-lg border border-aviation-border-panel">
                <div className="grid grid-cols-2 gap-3 text-xs">
                  <div>
                    <span className="text-aviation-text-dim">
                      Total Duration:{" "}
                    </span>
                    <span className="font-medium text-aviation-text-primary">
                      {formatDuration(selectedPath.totalDuration)}
                    </span>
                  </div>
                  <div>
                    <span className="text-aviation-text-dim">
                      Expected Cost:{" "}
                    </span>
                    <span className="font-medium text-aviation-cyan">
                      ${selectedPath.expectedCost.toFixed(2)}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center h-full text-aviation-text-muted">
              <GitFork className="w-8 h-8 mb-2 opacity-50" />
              <p className="text-sm">Select a path to view details</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Resource Collision Detector
// ============================================================================

interface ResourceCollision {
  id: string;
  resourceA: string;
  resourceB: string;
  type: "cpu" | "memory" | "io" | "network" | "disk";
  severity: "critical" | "high" | "medium" | "low";
  probability: number;
  impact: string;
  mitigation?: string;
}

interface ResourceCollisionDetectorProps {
  collisions: ResourceCollision[];
  selectedCollisionId?: string | null;
  threshold?: number;
  onCollisionSelect?: (collision: ResourceCollision) => void;
  onCollisionResolve?: (collisionId: string) => void;
  className?: string;
}

export const ResourceCollisionDetector: React.FC<
  ResourceCollisionDetectorProps
> = ({
  collisions,
  selectedCollisionId,
  threshold = 0.5,
  onCollisionSelect,
  onCollisionResolve,
  className,
}) => {
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return {
          bg: "bg-red-500/20",
          border: "border-red-500",
          text: "text-red-400",
          badge: "bg-red-500",
        };
      case "high":
        return {
          bg: "bg-amber-500/20",
          border: "border-amber-500",
          text: "text-amber-400",
          badge: "bg-amber-500",
        };
      case "medium":
        return {
          bg: "bg-yellow-500/20",
          border: "border-yellow-500",
          text: "text-yellow-400",
          badge: "bg-yellow-500",
        };
      default:
        return {
          bg: "bg-green-500/20",
          border: "border-green-500",
          text: "text-green-400",
          badge: "bg-green-500",
        };
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case "cpu":
        return <Cpu className="w-4 h-4" />;
      case "memory":
        return <Database className="w-4 h-4" />;
      case "io":
        return <Activity className="w-4 h-4" />;
      case "network":
        return <Network className="w-4 h-4" />;
      case "disk":
        return <Server className="w-4 h-4" />;
      default:
        return <Box className="w-4 h-4" />;
    }
  };

  const filteredCollisions = collisions.filter(
    (c) => c.probability >= threshold,
  );
  const criticalCount = filteredCollisions.filter(
    (c) => c.severity === "critical",
  ).length;

  return (
    <div
      className={cn(
        "flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden",
        className,
      )}
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-5 h-5 text-amber-400" />
            <h3 className="text-sm font-medium text-aviation-text-primary">
              Resource Collisions
            </h3>
          </div>
          <div className="flex items-center gap-2">
            {criticalCount > 0 && (
              <div className="flex items-center gap-1.5 px-2 py-1 bg-red-500/20 rounded border border-red-500/50">
                <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
                <span className="text-xs text-red-400 font-medium">
                  {criticalCount} Critical
                </span>
              </div>
            )}
            <span className="text-xs text-aviation-text-dim">
              {filteredCollisions.length} detected
            </span>
          </div>
        </div>
      </div>

      {/* Collision List */}
      <div className="flex-1 overflow-y-auto">
        {filteredCollisions.map((collision) => {
          const colors = getSeverityColor(collision.severity);
          const isSelected = selectedCollisionId === collision.id;

          return (
            <div
              key={collision.id}
              onClick={() => onCollisionSelect?.(collision)}
              className={cn(
                "p-4 border-b border-aviation-border-panel cursor-pointer transition-colors",
                isSelected
                  ? "bg-aviation-bg-instrument"
                  : "hover:bg-aviation-bg-secondary",
              )}
            >
              <div className="flex items-start justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className={colors.text}>
                    {getTypeIcon(collision.type)}
                  </span>
                  <span className="text-sm font-medium text-aviation-text-primary">
                    {collision.resourceA} ↔ {collision.resourceB}
                  </span>
                </div>
                <span
                  className={cn(
                    "px-1.5 py-0.5 text-[10px] rounded uppercase text-white",
                    colors.badge,
                  )}
                >
                  {collision.severity}
                </span>
              </div>

              <p className="text-xs text-aviation-text-dim mb-2">
                {collision.impact}
              </p>

              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-[10px] text-aviation-text-dim">
                    Probability:
                  </span>
                  <span className={cn("text-sm font-bold", colors.text)}>
                    {Math.round(collision.probability * 100)}%
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  {collision.mitigation && (
                    <span className="text-[10px] text-aviation-cyan">
                      Mitigation available
                    </span>
                  )}
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onCollisionResolve?.(collision.id);
                    }}
                    className="px-2 py-1 text-xs text-aviation-cyan hover:bg-aviation-cyan/10 rounded transition-colors"
                  >
                    Resolve
                  </button>
                </div>
              </div>
            </div>
          );
        })}

        {filteredCollisions.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <CheckCircle className="w-8 h-8 mb-2 text-green-400" />
            <p className="text-sm">No resource collisions detected</p>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// DollarSign icon helper (missing from lucide)
// ============================================================================
const DollarSign: React.FC<{ className?: string }> = ({ className }) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <line x1="12" y1="1" x2="12" y2="23" />
    <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
  </svg>
);
