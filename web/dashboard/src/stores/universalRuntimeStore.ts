import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

export type RuntimeStatus = 'ready' | 'active' | 'inactive' | 'error';
export type ExecutionMode = 'wasm' | 'gpu' | 'serverless' | 'browser' | 'edge' | 'hybrid';
export type ActiveView = 'overview' | 'wasm' | 'gpu' | 'serverless' | 'browser' | 'edge' | 'orchestrator' | 'topology' | 'routing' | 'inference';

export interface Runtime {
  id: string;
  name: string;
  type: ExecutionMode;
  status: RuntimeStatus;
  version: string;
  capabilities: string[];
  region?: string;
  endpoint?: string;
}

export interface CloudProvider {
  id: string;
  name: string;
  region: string;
  status: 'connected' | 'disconnected' | 'pending';
  capabilities: string[];
}

export interface InferenceProvider {
  id: string;
  name: string;
  model: string;
  status: 'active' | 'inactive' | 'degraded';
  latency?: number;
  throughput?: number;
}

export interface RuntimeCapability {
  id: string;
  name: string;
  supported: boolean;
  performance?: string;
}

export interface RuntimeMetrics {
  totalRequests: number;
  successfulRequests: number;
  failedRequests: number;
  averageLatency: number;
  throughput: number;
  activeConnections: number;
}

export interface RuntimeAlert {
  id: string;
  severity: 'info' | 'warning' | 'critical';
  message: string;
  timestamp: Date;
  dismissed: boolean;
}

interface UniversalRuntimeState {
  runtimes: Runtime[];
  selectedRuntimeId: string | null;
  activeView: ActiveView;
  executionMode: ExecutionMode;
  cloudProviders: CloudProvider[];
  inferenceProviders: InferenceProvider[];
  capabilities: RuntimeCapability[];
  metrics: RuntimeMetrics;
  alerts: RuntimeAlert[];
}

interface UniversalRuntimeActions {
  selectRuntime: (runtimeId: string | null) => void;
  setActiveView: (view: ActiveView) => void;
  setExecutionMode: (mode: ExecutionMode) => void;
  addCloudProvider: (provider: CloudProvider) => void;
  removeCloudProvider: (providerId: string) => void;
  selectInferenceProvider: (providerId: string) => void;
  updateMetrics: (metrics: Partial<RuntimeMetrics>) => void;
  dismissAlert: (alertId: string) => void;
  addRuntime: (runtime: Runtime) => void;
  removeRuntime: (runtimeId: string) => void;
  updateRuntimeStatus: (runtimeId: string, status: RuntimeStatus) => void;
}

const initialMetrics: RuntimeMetrics = {
  totalRequests: 0,
  successfulRequests: 0,
  failedRequests: 0,
  averageLatency: 0,
  throughput: 0,
  activeConnections: 0,
};

export const useUniversalRuntimeStore = create<UniversalRuntimeState & UniversalRuntimeActions>()(
  immer((set) => ({
    runtimes: [],
    selectedRuntimeId: null,
    activeView: 'overview',
    executionMode: 'wasm',
    cloudProviders: [],
    inferenceProviders: [],
    capabilities: [],
    metrics: initialMetrics,
    alerts: [],

    selectRuntime: (runtimeId) =>
      set((state) => {
        state.selectedRuntimeId = runtimeId;
      }),

    setActiveView: (view) =>
      set((state) => {
        state.activeView = view;
      }),

    setExecutionMode: (mode) =>
      set((state) => {
        state.executionMode = mode;
      }),

    addCloudProvider: (provider) =>
      set((state) => {
        state.cloudProviders.push(provider);
      }),

    removeCloudProvider: (providerId) =>
      set((state) => {
        state.cloudProviders = state.cloudProviders.filter((p) => p.id !== providerId);
      }),

    selectInferenceProvider: (providerId) =>
      set((state) => {
        const provider = state.inferenceProviders.find((p) => p.id === providerId);
        if (provider) {
          provider.status = 'active';
        }
      }),

    updateMetrics: (metrics) =>
      set((state) => {
        Object.assign(state.metrics, metrics);
      }),

    dismissAlert: (alertId) =>
      set((state) => {
        const alert = state.alerts.find((a) => a.id === alertId);
        if (alert) {
          alert.dismissed = true;
        }
      }),

    addRuntime: (runtime) =>
      set((state) => {
        state.runtimes.push(runtime);
      }),

    removeRuntime: (runtimeId) =>
      set((state) => {
        state.runtimes = state.runtimes.filter((r) => r.id !== runtimeId);
      }),

    updateRuntimeStatus: (runtimeId, status) =>
      set((state) => {
        const runtime = state.runtimes.find((r) => r.id === runtimeId);
        if (runtime) {
          runtime.status = status;
        }
      }),
  }))
);

export const selectReadyRuntimes = (state: UniversalRuntimeState) =>
  state.runtimes.filter((r) => r.status === 'ready');

export const selectActiveRuntimes = (state: UniversalRuntimeState) =>
  state.runtimes.filter((r) => r.status === 'active');

export const selectRuntimeByType = (type: ExecutionMode) => (state: UniversalRuntimeState) =>
  state.runtimes.filter((r) => r.type === type);

export const selectTotalThroughput = (state: UniversalRuntimeState) =>
  state.metrics.throughput;

export const useUniversalRuntime = () => useUniversalRuntimeStore();

export const useWasmExecution = () =>
  useUniversalRuntimeStore((state) => ({
    runtimes: state.runtimes.filter((r) => r.type === 'wasm'),
    activeRuntime: state.runtimes.find((r) => r.type === 'wasm' && r.status === 'active'),
    capabilities: state.capabilities.filter((c) => c.supported),
  }));

export const useGPUKernel = () =>
  useUniversalRuntimeStore((state) => ({
    runtimes: state.runtimes.filter((r) => r.type === 'gpu'),
    activeRuntime: state.runtimes.find((r) => r.type === 'gpu' && r.status === 'active'),
    metrics: state.metrics,
  }));

export const useServerlessRuntime = () =>
  useUniversalRuntimeStore((state) => ({
    runtimes: state.runtimes.filter((r) => r.type === 'serverless'),
    activeRuntime: state.runtimes.find((r) => r.type === 'serverless' && r.status === 'active'),
    cloudProviders: state.cloudProviders,
  }));

export const useBrowserAgent = () =>
  useUniversalRuntimeStore((state) => ({
    runtimes: state.runtimes.filter((r) => r.type === 'browser'),
    activeRuntime: state.runtimes.find((r) => r.type === 'browser' && r.status === 'active'),
    metrics: state.metrics,
  }));

export const useEdgeRuntime = () =>
  useUniversalRuntimeStore((state) => ({
    runtimes: state.runtimes.filter((r) => r.type === 'edge'),
    activeRuntime: state.runtimes.find((r) => r.type === 'edge' && r.status === 'active'),
    cloudProviders: state.cloudProviders,
  }));

export const useHybridOrchestrator = () =>
  useUniversalRuntimeStore((state) => ({
    runtimes: state.runtimes,
    selectedRuntimeId: state.selectedRuntimeId,
    executionMode: state.executionMode,
    metrics: state.metrics,
    alerts: state.alerts.filter((a) => !a.dismissed),
  }));

export const useCrossCloudTopology = () =>
  useUniversalRuntimeStore((state) => ({
    cloudProviders: state.cloudProviders,
    activeView: state.activeView,
    metrics: state.metrics,
  }));

export const useModelRouting = () =>
  useUniversalRuntimeStore((state) => ({
    inferenceProviders: state.inferenceProviders,
    selectedRuntimeId: state.selectedRuntimeId,
    capabilities: state.capabilities,
  }));

export const useInferenceSelector = () =>
  useUniversalRuntimeStore((state) => ({
    providers: state.inferenceProviders,
    selectedProviderId: state.selectedRuntimeId,
    selectProvider: state.selectInferenceProvider,
  }));

export const useCapabilityMatrix = () =>
  useUniversalRuntimeStore((state) => ({
    capabilities: state.capabilities,
    runtimes: state.runtimes,
    metrics: state.metrics,
  }));
