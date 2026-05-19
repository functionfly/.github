/**
 * @functionfly/ui-runtime
 * Type definitions
 */

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

export interface GPUKernelInspectorProps {
  kernels: Array<{
    id: string;
    name: string;
    utilization: number;
    memoryUsed: number;
    memoryTotal: number;
    status: "running" | "idle" | "queued" | "error";
    executionTime: number;
  }>;
  onKernelClick?: (id: string) => void;
  className?: string;
}

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

export interface ModelRoute {
  modelId: string;
  modelName: string;
  provider: string;
  priority: number;
  avgLatency: number;
  costPer1K: number;
  dailyQuota: number;
  usedToday: number;
  successRate: number;
}

export interface ModelRoutingVisualizerProps {
  routes: ModelRoute[];
  activeRouteId?: string;
  onRouteSelect?: (id: string) => void;
  className?: string;
}

export interface InferenceProvider {
  id: string;
  name: string;
  status: "available" | "degraded" | "unavailable";
  avgLatency: number;
  throughput: number;
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
