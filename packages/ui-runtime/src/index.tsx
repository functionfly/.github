/**
 * @functionfly/ui-runtime
 * Universal runtime selection and monitoring components
 */

import * as React from "react";
import { cn, Badge } from "@functionfly/ui-core";
import {
  Server,
  Globe,
  Cpu,
  Wrench,
  Cloud,
  LayoutDashboard,
  Eye,
  CheckCircle,
  XCircle,
  Loader2,
  RefreshCcw,
  ArrowRight,
  BarChart2,
  Clock,
  Zap,
  HardDrive,
  Network,
  AlertTriangle,
  Search,
} from "lucide-react";

// --- Types ---
export type RuntimeType =
  | "wasm"
  | "docker"
  | "serverless"
  | "browser"
  | "gpu"
  | "edge"
  | "hybrid"
  | "local";

export interface RuntimeDescriptor {
  id: string;
  name: string;
  type: RuntimeType;
  provider: string;
  region: string;
  status: "online" | "offline" | "degraded" | "pending";
  latency: number;
  costPerExecution: number;
  supportedLanguages: string[];
  maxMemory: number;
  maxTimeout: number;
  features: string[];
  currentLoad?: number;
  reliabilityScore?: number;
}

export interface RuntimeSelection {
  runtimeId: string;
  config: Record<string, unknown>;
  priority: "cost" | "speed" | "reliability" | "privacy";
}

export interface RuntimeTargetSelectorProps {
  runtimes: RuntimeDescriptor[];
  selectedId?: string;
  onSelect?: (id: string) => void;
  autoSuggest?: boolean;
  onAutoSuggest?: (criteria: { priority: string }) => void;
  className?: string;
}

export interface RuntimeCardProps extends RuntimeDescriptor {
  selected?: boolean;
  onClick?: (id: string) => void;
}

export interface RuntimeCapabilityMatrixProps {
  runtimes: RuntimeDescriptor[];
  features: string[];
  className?: string;
}

export interface WasmExecutionPanelProps {
  runtimeId: string;
  status: "idle" | "building" | "running" | "error";
  logs: Array<{ timestamp: string; level: "info" | "warn" | "error" | "debug"; message: string }>;
  onExecute?: () => void;
  onStop?: () => void;
  className?: string;
}

export interface ServerlessRuntimeViewerProps {
  deployments: Array<{
    id: string;
    name: string;
    runtime: string;
    region: string;
    status: "active" | "deploying" | "error" | "scaling";
    invocations: number;
    errorRate: number;
    latency: number;
  }>;
  className?: string;
}

export interface EdgeRuntimeMapProps {
  nodes: Array<{
    id: string;
    region: string;
    location: { lat: number; lng: number };
    status: "online" | "offline" | "degraded";
    load: number;
    latency: number;
  }>;
  className?: string;
}

// --- Helpers ---
const runtimeColors: Record<RuntimeType, string> = {
  wasm: "#f97316",
  docker: "#3b82f6",
  serverless: "#10b981",
  browser: "#8b5cf6",
  gpu: "#ec4899",
  edge: "#06b6d4",
  hybrid: "#a855f7",
  local: "#6366f1",
};

const runtimeIcons: Record<RuntimeType, React.ReactNode> = {
  wasm: <Cpu className="size-4" />,
  docker: <HardDrive className="size-4" />,
  serverless: <Cloud className="size-4" />,
  browser: <Globe className="size-4" />,
  gpu: <Zap className="size-4" />,
  edge: <Network className="size-4" />,
  hybrid: <LayoutDashboard className="size-4" />,
  local: <Server className="size-4" />,
};

function getStatusConfig(status: string) {
  return {
    online: { color: "#10b981", icon: <CheckCircle className="size-3" /> },
    offline: { color: "#6b7280", icon: <XCircle className="size-3" /> },
    degraded: { color: "#f59e0b", icon: <AlertTriangle className="size-3" /> },
    pending: { color: "#3b82f6", icon: <Loader2 className="size-3 animate-spin" /> },
  }[status] || { color: "#6b7280", icon: <XCircle className="size-3" /> };
}

// --- RuntimeCard ---
export function RuntimeCard({
  id,
  name,
  type,
  provider,
  region,
  status,
  latency,
  costPerExecution,
  supportedLanguages,
  reliabilityScore,
  selected,
  onClick,
  currentLoad,
}: RuntimeCardProps) {
  const color = runtimeColors[type] || "#6b7280";
  const statusConfig = getStatusConfig(status);

  return (
    <div
      onClick={() => onClick?.(id)}
      className={cn(
        "relative p-3 bg-bg-primary border rounded-lg cursor-pointer transition-all duration-200",
        selected ? "border-brand-500 shadow-brand-500/10 shadow-md" : "border-border-subtle hover:border-border-default hover:shadow-sm",
      )}
      style={selected ? { boxShadow: `0 0 12px ${color}22` } : undefined}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2 min-w-0">
          <div className="size-8 rounded-md flex items-center justify-center shrink-0"
            style={{ backgroundColor: `${color}15`, color }}>
            {runtimeIcons[type]}
          </div>
          <div className="min-w-0">
            <h4 className="text-sm font-semibold text-text-primary truncate">{name}</h4>
            <p className="text-[10px] text-text-muted truncate">{provider} · {region}</p>
          </div>
        </div>
        <div className="flex items-center gap-1 text-[10px] shrink-0 ml-2">
          {statusConfig.icon}
          <span style={{ color: statusConfig.color }} className="hidden sm:inline">{status}</span>
        </div>
      </div>

      {/* Stats */}
      <div className="flex items-center gap-3 mb-2">
        <div className="flex items-center gap-1.5">
          <div className="text-[10px] text-text-muted">Late</div>
          <div className="text-xs font-bold text-text-primary">{latency}m</div>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="text-[10px] text-text-muted">Cost</div>
          <div className="text-xs font-bold text-brand-500">${costPerExecution.toFixed(4)}</div>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="text-[10px] text-text-muted">Reli</div>
          <div className="text-xs font-bold" style={{ color: reliabilityScore != null ? getStatusColor(reliabilityScore) : "#6b7280" }}>
            {reliabilityScore ?? "—"}%
          </div>
        </div>
      </div>

      {/* Load bar */}
      {currentLoad != null && (
        <div className="mb-2">
          <div className="flex justify-between text-[10px] text-text-muted mb-0.5">
            <span>Load</span>
            <span>{currentLoad}%</span>
          </div>
          <div className="h-1 bg-bg-tertiary rounded-full overflow-hidden">
            <div className="h-full rounded-full transition-all duration-500"
              style={{
                width: `${currentLoad}%`,
                backgroundColor: currentLoad > 80 ? "#ef4444" : currentLoad > 50 ? "#f59e0b" : color,
              }}
            />
          </div>
        </div>
      )}

      {/* Language badges */}
      <div className="flex flex-wrap gap-1">
        {supportedLanguages.map((lang) => (
          <span key={lang} className="px-1.5 py-0.5 text-[9px] bg-bg-tertiary text-text-muted rounded capitalize">
            {lang}
          </span>
        ))}
      </div>

      {/* Selected indicator */}
      {selected && (
        <div className="absolute top-1.5 right-1.5 size-2 rounded-full bg-brand-500 border-2 border-bg-primary" />
      )}
    </div>
  );
}

function getStatusColor(score: number): string {
  if (score >= 80) return "#10b981";
  if (score >= 50) return "#f59e0b";
  return "#ef4444";
}

// --- RuntimeTargetSelector ---
export function RuntimeTargetSelector({
  runtimes,
  selectedId,
  onSelect,
  autoSuggest,
  onAutoSuggest,
  className,
}: RuntimeTargetSelectorProps) {
  const [filter, setFilter] = React.useState("");
  const [sortBy, setSortBy] = React.useState<"latency" | "cost" | "reliability">("latency");

  const filtered = React.useMemo(() => {
    let result = runtimes;
    if (filter) {
      result = result.filter(
        (r) =>
          r.name.toLowerCase().includes(filter.toLowerCase()) ||
          r.type.toLowerCase().includes(filter.toLowerCase()) ||
          r.provider.toLowerCase().includes(filter.toLowerCase())
      );
    }
    return [...result].sort((a, b) => {
      if (sortBy === "latency") return a.latency - b.latency;
      if (sortBy === "cost") return a.costPerExecution - b.costPerExecution;
      if (sortBy === "reliability") return (b.reliabilityScore ?? 0) - (a.reliabilityScore ?? 0);
      return 0;
    });
  }, [runtimes, filter, sortBy]);

  // Auto-select best
  const bestRuntime = filtered.length > 0 ? filtered[0] : undefined;

  return (
    <div className={cn("space-y-3", className)}>
      {/* Toolbar */}
      <div className="flex flex-col gap-2">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-text-muted" />
          <input
            type="text"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder="Search runtimes..."
            className="w-full pl-8 pr-3 py-1.5 text-xs bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500 transition-colors"
          />
        </div>
        <select
          value={sortBy}
          onChange={(e) => setSortBy(e.target.value as any)}
          className="px-2.5 py-1.5 text-xs bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500"
        >
          <option value="latency">Lowest Latency</option>
          <option value="cost">Lowest Cost</option>
          <option value="reliability">Best Reliability</option>
        </select>
      </div>

      {/* Auto-suggest */}
      {autoSuggest && bestRuntime && (
        <div
          className="flex items-center gap-3 p-3 bg-brand-500/10 border border-brand-500/20 rounded-lg cursor-pointer hover:bg-brand-500/20 transition-colors"
          onClick={() => {
            onSelect?.(bestRuntime.id);
            onAutoSuggest?.({ priority: sortBy });
          }}
        >
          <Zap className="size-5 text-brand-400" />
          <div className="flex-1">
            <div className="text-sm font-medium text-brand-400">Auto-Suggested: {bestRuntime.name}</div>
            <div className="text-[11px] text-text-muted">Best {sortBy} · {bestRuntime.latency}ms · ${bestRuntime.costPerExecution.toFixed(4)}/call</div>
          </div>
          <ArrowRight className="size-4 text-brand-400" />
        </div>
      )}

      {/* Grid */}
      <div className="grid grid-cols-1 gap-2">
        {filtered.map((runtime) => (
          <RuntimeCard
            key={runtime.id}
            {...runtime}
            selected={runtime.id === selectedId}
            onClick={(id) => onSelect?.(id)}
          />
        ))}
      </div>
    </div>
  );
}


// --- WasmExecutionPanel ---
export function WasmExecutionPanel({
  runtimeId,
  status,
  logs,
  onExecute,
  onStop,
  className,
}: WasmExecutionPanelProps) {
  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (ref.current) {
      ref.current.scrollTop = ref.current.scrollHeight;
    }
  }, [logs]);

  return (
    <div className={cn("space-y-3", className)}>
      {/* Status bar */}
      <div className="flex items-center justify-between p-3 bg-bg-secondary rounded-lg border border-border-subtle">
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "size-2.5 rounded-full",
              status === "running" && "bg-success animate-pulse",
              status === "error" && "bg-error",
              status === "building" && "bg-warning animate-pulse",
              status === "idle" && "bg-text-muted"
            )}
          />
          <span className="text-sm font-medium text-text-primary capitalize">{status}</span>
        </div>
        <div className="flex gap-2">
          {onExecute && (status === "idle" || status === "error") && (
            <button onClick={onExecute} className="px-3 py-1.5 text-[11px] bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors">
              Run
            </button>
          )}
          {onStop && status === "running" && (
            <button onClick={onStop} className="px-3 py-1.5 text-[11px] bg-error/10 hover:bg-error/20 text-error rounded-lg transition-colors">
              Stop
            </button>
          )}
        </div>
      </div>

      {/* Logs */}
      <div
        ref={ref}
        className="h-64 bg-bg-primary rounded-lg border border-border-subtle p-3 font-mono text-[11px] overflow-y-auto"
      >
        {logs.length === 0 ? (
          <div className="h-full flex items-center justify-center text-text-muted">
            No logs yet. Execute to see output.
          </div>
        ) : (
          <div className="space-y-0.5">
            {logs.map((log, i) => (
              <div
                key={i}
                className="flex gap-2 py-0.5"
                style={{
                  color:
                    log.level === "error"
                      ? "#ef4444"
                      : log.level === "warn"
                      ? "#f59e0b"
                      : log.level === "debug"
                      ? "#6b7280"
                      : "#a0a0b0",
                }}
              >
                <span className="text-text-muted shrink-0">{log.timestamp}</span>
                <span className="uppercase font-bold shrink-0" style={{ width: 36 }}>
                  [{log.level}]
                </span>
                <span>{log.message}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}



// --- RuntimeCapabilityMatrix ---
export function RuntimeCapabilityMatrix({
  runtimes,
  features,
  className,
}: RuntimeCapabilityMatrixProps) {
  return (
    <div className={cn("space-y-4", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Server className="size-4 text-brand-400" /> Capability Matrix
      </h4>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border-subtle">
              <th className="text-left py-2 px-3 text-text-muted font-medium">Runtime</th>
              {features.map((f) => (
                <th key={f} className="text-center py-2 px-3 text-text-muted font-medium capitalize">
                  {f.replace('-', ' ')}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {runtimes.map((runtime) => (
              <tr key={runtime.id} className="border-b border-border-subtle hover:bg-bg-secondary/50">
                <td className="py-2 px-3 text-text-primary font-medium">{runtime.name}</td>
                {features.map((f) => {
                  const hasFeature = runtime.features?.includes(f) ?? false;
                  return (
                    <td key={f} className="text-center py-2 px-3">
                      {hasFeature ? (
                        <span className="text-success text-lg">✓</span>
                      ) : (
                        <span className="text-text-muted text-lg">—</span>
                      )}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// --- ServerlessRuntimeViewer ---
export function ServerlessRuntimeViewer({ deployments, className }: ServerlessRuntimeViewerProps) {
  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Server className="size-4 text-brand-400" /> Deployments
      </h4>
      <div className="space-y-2">
        {deployments.map((dep) => {
          const statusColor = dep.status === "active" ? "#10b981" : dep.status === "deploying" ? "#3b82f6" : dep.status === "error" ? "#ef4444" : "#f59e0b";
          return (
            <div key={dep.id} className="flex items-center justify-between p-3 bg-bg-primary rounded-lg border border-border-subtle hover:border-border-default transition-colors">
              <div className="flex items-center gap-3">
                <div className="size-2 rounded-full" style={{ backgroundColor: statusColor }} />
                <div>
                  <div className="text-sm font-medium text-text-primary">{dep.name}</div>
                  <div className="text-[11px] text-text-muted">
                    {dep.runtime} · {dep.region}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-4 text-[10px] text-text-muted">
                <span>{dep.invocations.toLocaleString()} calls</span>
                <span>{dep.latency}ms</span>
                <span style={{ color: dep.errorRate > 1 ? "#ef4444" : "#10b981" }}>{dep.errorRate}% errors</span>
                <Badge variant={dep.status === "active" ? "success" : dep.status === "deploying" ? "default" : "error"} size="sm">
                  {dep.status}
                </Badge>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// --- EdgeRuntimeMap (simplified 2D D3-style) ---
export function EdgeRuntimeMap({ nodes, className }: EdgeRuntimeMapProps) {
  return (
    <div className={cn("relative bg-[#08080f] rounded-xl overflow-hidden border border-border-subtle", className)}>
      <div className="p-3 absolute top-0 left-0 right-0 z-10 flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <Network className="size-4 text-brand-400" /> Edge Nodes
        </h4>
        <Badge variant="brand" size="sm">Live</Badge>
      </div>
      <svg className="w-full h-64" viewBox="0 0 800 256" preserveAspectRatio="xMidYMid meet">
        {/* Simplified world line */}
        <path d="M 50 160 Q 200 100 400 140 T 750 120" fill="none" stroke="rgba(255,255,255,0.08)" strokeWidth="2" />
        {nodes.map((node, i) => {
          const x = 50 + (i / (nodes.length - 1 || 1)) * 700;
          const y = 140 + Math.sin(i * 1.5) * 30;
          const color = node.status === "online" ? "#10b981" : node.status === "degraded" ? "#f59e0b" : "#ef4444";
          return (
            <g key={node.id}>
              <circle cx={x} cy={y} r={12 + (node.load / 100) * 8} fill={color} opacity={0.2} />
              <circle cx={x} cy={y} r={8} fill={color} opacity={0.6} />
              <circle cx={x} cy={y} r={4} fill="#fff" />
              <text x={x} y={y - 18} textAnchor="middle" fill="#a0a0b0" fontSize={10}>
                {node.region}
              </text>
              <text x={x} y={y + 22} textAnchor="middle" fill={color} fontSize={9}>
                {node.latency}ms
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
}


// --- GPUKernelInspector ---
export interface GPUKernelInspectorProps {
  kernels: Array<{
    id: string;
    name: string;
    utilization: number; // 0-100
    memoryUsed: number; // MB
    memoryTotal: number; // MB
    status: "running" | "idle" | "queued" | "error";
    executionTime: number; // ms
  }>;
  onKernelClick?: (id: string) => void;
  className?: string;
}

export function GPUKernelInspector({ kernels, onKernelClick, className }: GPUKernelInspectorProps) {
  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Cpu className="size-4 text-pink-400" /> GPU Kernel Inspector
      </h4>

      <div className="space-y-2">
        {kernels.map((kernel) => {
          const utilColor = kernel.utilization > 80 ? "#ef4444" : kernel.utilization > 50 ? "#f59e0b" : "#10b981";
          const memPct = (kernel.memoryUsed / kernel.memoryTotal) * 100;

          return (
            <div
              key={kernel.id}
              onClick={() => onKernelClick?.(kernel.id)}
              className="p-3 bg-bg-primary rounded-lg border border-border-subtle hover:border-border-default cursor-pointer transition-colors"
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-text-primary">{kernel.name}</span>
                  <span className={cn(
                    "text-[9px] px-1.5 py-0.5 rounded capitalize",
                    kernel.status === "running" ? "bg-emerald-500/20 text-emerald-400" :
                    kernel.status === "idle" ? "bg-gray-500/20 text-gray-400" :
                    kernel.status === "queued" ? "bg-blue-500/20 text-blue-400" :
                    "bg-red-500/20 text-red-400"
                  )}>
                    {kernel.status}
                  </span>
                </div>
                <span className="text-[10px] text-text-muted font-mono">{kernel.executionTime}ms</span>
              </div>

              {/* Utilization bar */}
              <div className="mb-1.5">
                <div className="flex justify-between text-[9px] text-text-muted mb-0.5">
                  <span>Utilization</span>
                  <span style={{ color: utilColor }}>{kernel.utilization}%</span>
                </div>
                <div className="h-1.5 bg-bg-tertiary rounded-full overflow-hidden">
                  <div className="h-full rounded-full" style={{ width: `${kernel.utilization}%`, backgroundColor: utilColor }} />
                </div>
              </div>

              {/* Memory bar */}
              <div>
                <div className="flex justify-between text-[9px] text-text-muted mb-0.5">
                  <span>Memory</span>
                  <span>{kernel.memoryUsed}MB / {kernel.memoryTotal}MB</span>
                </div>
                <div className="h-1 bg-bg-tertiary rounded-full overflow-hidden">
                  <div className="h-full rounded-full bg-purple-500" style={{ width: `${memPct}%` }} />
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// --- CrossCloudTopologyMap ---
export interface CloudNode {
  id: string;
  name: string;
  provider: "aws" | "gcp" | "azure" | "cloudflare" | "self-hosted";
  region: string;
  type: "compute" | "storage" | "network" | "database" | "ai";
  status: "online" | "degraded" | "offline";
  latency: number;
  costPerHour: number;
  connections: string[];
}

export interface CrossCloudTopologyMapProps {
  nodes: CloudNode[];
  className?: string;
}

const providerColors: Record<string, string> = {
  aws: "#ff9900",
  gcp: "#4285f4",
  azure: "#0078d4",
  cloudflare: "#f38020",
  "self-hosted": "#6b7280",
};

const typeIcons: Record<string, React.ReactNode> = {
  compute: <Cpu className="size-3" />,
  storage: <HardDrive className="size-3" />,
  network: <Network className="size-3" />,
  database: <Server className="size-3" />,
  ai: <Brain className="size-3" />,
};

export function CrossCloudTopologyMap({ nodes, className }: CrossCloudTopologyMapProps) {
  const [selected, setSelected] = React.useState<string | null>(null);

  // Simple force-directed layout approximation
  const positions = React.useMemo(() => {
    const pos: Record<string, { x: number; y: number }> = {};
    const centerX = 200;
    const centerY = 150;
    const radius = 140;

    nodes.forEach((node, i) => {
      if (nodes.length === 1) {
        pos[node.id] = { x: centerX, y: centerY };
      } else {
        const angle = (2 * Math.PI * i) / nodes.length - Math.PI / 2;
        pos[node.id] = {
          x: centerX + radius * Math.cos(angle),
          y: centerY + radius * Math.sin(angle),
        };
      }
    });
    return pos;
  }, [nodes]);

  return (
    <div className={cn("relative", className)}>
      <div className="p-3 absolute top-0 left-0 right-0 z-10 flex items-center justify-between">
        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
          <Globe className="size-4 text-brand-400" /> Multi-Cloud Topology
        </h4>
        <div className="flex gap-2">
          {Object.entries(providerColors).map(([provider, color]) => (
            <div key={provider} className="flex items-center gap-1 text-[9px] text-text-muted">
              <div className="size-2 rounded-full" style={{ backgroundColor: color }} />
              <span className="capitalize">{provider}</span>
            </div>
          ))}
        </div>
      </div>

      <svg width="400" height="300" className="w-full h-auto">
        <g>
          {nodes.flatMap((node) =>
            node.connections.map((targetId) => {
              const from = positions[node.id];
              const to = positions[targetId];
              if (!from || !to) return null;

              const midX = (from.x + to.x) / 2;
              const midY = (from.y + to.y) / 2;

              return (
                <g key={`${node.id}-${targetId}`}>
                  <line
                    x1={from.x}
                    y1={from.y}
                    x2={to.x}
                    y2={to.y}
                    stroke="rgba(255,255,255,0.1)"
                    strokeWidth="1"
                    strokeDasharray="4,4"
                  />
                  <circle cx={midX} cy={midY} r={2} fill="rgba(255,255,255,0.3)" />
                </g>
              );
            })
          )}
        </g>
      </svg>
    </div>
  );
}

// --- InferenceProviderSelector ---
export interface InferenceProvider {
  id: string;
  name: string;
  status: "available" | "degraded" | "unavailable";
  avgLatency: number;
  throughput: number; // tokens/sec
  costPer1M: number;
  supportsStreaming: boolean;
  models: string[];
  regions: string[];
}

export interface InferenceProviderSelectorProps {
  providers: InferenceProvider[];
  selectedId?: string;
  onSelect?: (id: string) => void;
  className?: string;
}

export function InferenceProviderSelector({
  providers,
  selectedId,
  onSelect,
  className,
}: InferenceProviderSelectorProps) {
  const [filterStatus, setFilterStatus] = React.useState<InferenceProvider["status"] | "all">("all");

  const filtered = filterStatus === "all"
    ? providers
    : providers.filter((p) => p.status === filterStatus);

  const bestProvider = [...filtered].sort((a, b) => a.avgLatency - b.avgLatency)[0];

  return (
    <div className={cn("space-y-3", className)}>
      <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
        <Zap className="size-4 text-brand-400" /> Inference Provider
      </h4>

      {/* Filter tabs */}
      <div className="flex gap-1">
        {(["all", "available", "degraded", "unavailable"] as const).map((status) => (
          <button
            key={status}
            onClick={() => setFilterStatus(status)}
            className={cn(
              "px-2 py-0.5 text-[10px] rounded capitalize transition-colors",
              filterStatus === status ? "bg-brand-500/20 text-brand-400 font-medium" : "text-text-muted hover:text-text-secondary"
            )}
          >
            {status}
          </button>
        ))}
      </div>

      {/* Auto-suggest */}
      {bestProvider && (
        <div
          className="flex items-center gap-3 p-3 bg-brand-500/10 border border-brand-500/20 rounded-lg cursor-pointer hover:bg-brand-500/20"
          onClick={() => onSelect?.(bestProvider.id)}
        >
          <Zap className="size-4 text-brand-400" />
          <div className="flex-1">
            <div className="text-sm font-medium text-brand-400">Recommended: {bestProvider.name}</div>
            <div className="text-[11px] text-text-muted">
              {bestProvider.avgLatency}ms latency · {bestProvider.throughput} tok/sec · ${bestProvider.costPer1M.toFixed(2)}/1M
            </div>
          </div>
          <ArrowRight className="size-4 text-brand-400" />
        </div>
      )}

      {/* Provider list */}
      <div className="space-y-2">
        {filtered.map((provider) => {
          const statusColors = {
            available: "border-emerald-500/30",
            degraded: "border-amber-500/30",
            unavailable: "border-red-500/30",
          };
          const dotColors = {
            available: "bg-emerald-400",
            degraded: "bg-amber-400",
            unavailable: "bg-red-400",
          };

          return (
            <div
              key={provider.id}
              onClick={() => onSelect?.(provider.id)}
              className={cn(
                "p-3 rounded-lg border cursor-pointer transition-all",
                selectedId === provider.id ? "border-brand-500" : `border-border-subtle ${statusColors[provider.status]}`,
                provider.status === "unavailable" && "opacity-50"
              )}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <div className={cn("size-2 rounded-full", dotColors[provider.status])} />
                  <span className="text-sm font-medium text-text-primary">{provider.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  {provider.supportsStreaming && (
                    <span className="text-[9px] px-1.5 py-0.5 bg-blue-500/20 text-blue-400 rounded">Stream</span>
                  )}
                  <span className="text-[10px] text-text-muted capitalize">{provider.status}</span>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-2 text-[10px]">
                <div>
                  <span className="text-text-muted">Latency</span>
                  <div className="text-text-primary font-mono">{provider.avgLatency}ms</div>
                </div>
                <div>
                  <span className="text-text-muted">Throughput</span>
                  <div className="text-text-primary font-mono">{provider.throughput} tok/s</div>
                </div>
                <div>
                  <span className="text-text-muted">Cost</span>
                  <div className="text-brand-400 font-mono">${provider.costPer1M.toFixed(2)}</div>
                </div>
              </div>

              {/* Supported models */}
              <div className="flex flex-wrap gap-1 mt-2">
                {provider.models.slice(0, 4).map((model) => (
                  <span key={model} className="px-1.5 py-0.5 text-[9px] bg-bg-tertiary text-text-muted rounded">
                    {model}
                  </span>
                ))}
                {provider.models.length > 4 && (
                  <span className="text-[9px] text-text-muted">+{provider.models.length - 4}</span>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function Brain({ className }: { className?: string }) {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
      <path d="M12 5a3 3 0 1 0-5.997.125 4 4 0 0 0-2.526 5.77 4 4 0 0 0 .556 6.588A4 4 0 1 0 12 18Z" />
      <path d="M12 5a3 3 0 1 1 5.997.125 4 4 0 0 1 2.526 5.77 4 4 0 0 1-.556 6.588A4 4 0 1 1 12 18Z" />
      <path d="M15 13a4.5 4.5 0 0 1-3-4 4.5 4.5 0 0 1-3 4" />
      <path d="M17.599 6.5a3 3 0 0 0 .399-1.375" />
      <path d="M6.003 5.125A3 3 0 0 0 6.401 6.5" />
      <path d="M3.477 10.896a4 4 0 0 1 .585-.396" />
      <path d="M19.938 10.5a4 4 0 0 1 .585.396" />
      <path d="M6 18a4 4 0 0 1-1.967-.516" />
      <path d="M19.967 17.484A4 4 0 0 1 18 18" />
    </svg>
  );
}

// Note: Types and components exported inline

// Note: Components exported above, no self-reference needed