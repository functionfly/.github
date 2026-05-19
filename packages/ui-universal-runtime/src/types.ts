/**
 * @functionfly/ui-universal-runtime
 * Universal Runtime Components - Multi-runtime execution visualization
 */

// ============================================================================
// Shared Types
// ============================================================================

export type RuntimeType = 'wasm' | 'native' | 'serverless' | 'browser' | 'edge' | 'gpu';
export type ExecutionMode = 'synchronous' | 'asynchronous' | 'streaming';
export type RuntimeStatus = 'ready' | 'busy' | 'error' | 'offline';
export type CloudProvider = 'aws' | 'gcp' | 'azure' | 'cloudflare' | 'local';
export type KernelStatus = 'idle' | 'running' | 'queued' | 'completed';
export type InferenceProvider = 'openai' | 'anthropic' | 'local' | 'custom';

export interface RuntimeMetrics {
  cpuUsage: number;
  memoryUsage: number;
  gpuUsage?: number;
  executionCount: number;
  avgLatency: number;
  errorRate: number;
  timestamp: number;
}

export interface KernelInfo {
  id: string;
  name: string;
  status: KernelStatus;
  executionTime?: number;
  memoryFootprint?: number;
}

// ============================================================================
// Runtime Abstraction UI
// ============================================================================

export interface RuntimeInstance {
  id: string;
  name: string;
  type: RuntimeType;
  status: RuntimeStatus;
  region?: string;
  metrics: RuntimeMetrics;
}

export interface RuntimeAbstractionUIProps {
  instances: RuntimeInstance[];
  selectedInstanceId?: string | null;
  onInstanceSelect?: (instance: RuntimeInstance) => void;
  onInstanceHover?: (instance: RuntimeInstance | null) => void;
  className?: string;
}

// ============================================================================
// WebAssembly Execution Panel
// ============================================================================

export interface WasmExecution {
  id: string;
  moduleName: string;
  functionName: string;
  inputSize: number;
  outputSize: number;
  executionTime: number;
  memoryUsed: number;
  timestamp: number;
}

export interface WasmExecutionPanelProps {
  executions: WasmExecution[];
  selectedExecutionId?: string | null;
  onExecutionSelect?: (execution: WasmExecution) => void;
  onRefresh?: () => void;
  className?: string;
}

// ============================================================================
// GPU Kernel Inspector
// ============================================================================

export interface GPUKernel {
  id: string;
  name: string;
  gridSize: [number, number, number];
  blockSize: [number, number, number];
  status: KernelStatus;
  executionTime: number;
  memoryUsage: number;
  registersPerThread?: number;
  sharedMemoryUsage?: number;
}

export interface GPUKernelInspectorProps {
  kernels: GPUKernel[];
  selectedKernelId?: string | null;
  onKernelSelect?: (kernel: GPUKernel) => void;
  onKernelLaunch?: (kernelId: string) => void;
  className?: string;
}

// ============================================================================
// Serverless Runtime Viewer
// ============================================================================

export interface ServerlessFunction {
  id: string;
  name: string;
  runtime: string;
  memory: number;
  timeout: number;
  invocationCount: number;
  errorRate: number;
  avgDuration: number;
}

export interface ServerlessRuntimeViewerProps {
  functions: ServerlessFunction[];
  selectedFunctionId?: string | null;
  onFunctionSelect?: (fn: ServerlessFunction) => void;
  className?: string;
}

// ============================================================================
// Browser Agent Session
// ============================================================================

export interface AgentAction {
  type: string;
  timestamp: number;
  duration?: number;
  success: boolean;
  screenshot?: string;
}

export interface BrowserAgentSession {
  id: string;
  agentId: string;
  status: 'active' | 'paused' | 'completed' | 'failed';
  startedAt: number;
  actions: AgentAction[];
  currentUrl?: string;
}

export interface BrowserAgentSessionProps {
  session: BrowserAgentSession | null;
  onSessionPause?: () => void;
  onSessionResume?: () => void;
  onSessionStop?: () => void;
  className?: string;
}

// ============================================================================
// Edge Runtime Map
// ============================================================================

export interface EdgeNode {
  id: string;
  name: string;
  location: string;
  status: RuntimeStatus;
  latency: number;
  throughput: number;
  activeConnections: number;
}

export interface EdgeRuntimeMapProps {
  nodes: EdgeNode[];
  selectedNodeId?: string | null;
  onNodeSelect?: (node: EdgeNode) => void;
  onNodeHover?: (node: EdgeNode | null) => void;
  className?: string;
}

// ============================================================================
// Hybrid Execution Orchestrator
// ============================================================================

export interface HybridTask {
  id: string;
  name: string;
  stages: Array<{
    runtimeType: RuntimeType;
    status: KernelStatus;
    startTime?: number;
    endTime?: number;
  }>;
  currentStage: number;
  totalDuration: number;
}

export interface HybridExecutionOrchestratorProps {
  tasks: HybridTask[];
  selectedTaskId?: string | null;
  onTaskSelect?: (task: HybridTask) => void;
  onTaskCancel?: (taskId: string) => void;
  className?: string;
}

// ============================================================================
// Cross-Cloud Topology Map
// ============================================================================

export interface CloudNode {
  id: string;
  provider: CloudProvider;
  region: string;
  status: RuntimeStatus;
  metrics: RuntimeMetrics;
  connections: string[];
}

export interface CrossCloudTopologyMapProps {
  nodes: CloudNode[];
  selectedNodeId?: string | null;
  onNodeSelect?: (node: CloudNode) => void;
  className?: string;
}

// ============================================================================
// Model Routing Visualizer
// ============================================================================

export interface ModelRoute {
  id: string;
  modelName: string;
  provider: InferenceProvider;
  requestCount: number;
  avgLatency: number;
  successRate: number;
  fallbackProvider?: InferenceProvider;
}

export interface ModelRoutingVisualizerProps {
  routes: ModelRoute[];
  selectedRouteId?: string | null;
  onRouteSelect?: (route: ModelRoute) => void;
  className?: string;
}

// ============================================================================
// Inference Provider Selector
// ============================================================================

export interface InferenceOption {
  provider: InferenceProvider;
  name: string;
  models: string[];
  latency: number;
  costPer1kTokens: number;
  available: boolean;
  capabilities: string[];
}

export interface InferenceProviderSelectorProps {
  options: InferenceOption[];
  selectedProvider?: InferenceProvider | null;
  onProviderSelect?: (provider: InferenceProvider) => void;
  onCompare?: () => void;
  className?: string;
}

// ============================================================================
// Runtime Capability Matrix
// ============================================================================

export interface RuntimeCapability {
  runtimeType: RuntimeType;
  maxMemory: number;
  maxCompute: number;
  supportsStreaming: boolean;
  supportsConcurrency: boolean;
  supportsGpu: boolean;
  supportedLanguages: string[];
  estimatedCostPerHour: number;
}

export interface RuntimeCapabilityMatrixProps {
  capabilities: RuntimeCapability[];
  selectedRuntimeType?: RuntimeType | null;
  onRuntimeSelect?: (runtimeType: RuntimeType) => void;
  className?: string;
}
