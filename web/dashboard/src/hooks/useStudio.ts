import { API_BASE_URL } from '@/lib/constants';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useMemo, useState } from 'react';
import { toast } from 'sonner';

// ============================================================================
// Types
// ============================================================================

export type ViewMode = 'design' | 'execute' | 'debug' | 'monitor' | 'simulate';
export type ComplexityLevel = 'simple' | 'moderate' | 'complex' | 'extreme';

export interface StudioState {
  selectedPanel: string | null;
  selectedTab: string | null;
  complexityLevel: ComplexityLevel;
  zoomLevel: number;
  viewMode: ViewMode;
  isPanelCollapsed: boolean;
  showMinimap: boolean;
  gridEnabled: boolean;
  snapToGrid: boolean;
}

export interface AgentStatus {
  id: string;
  agentId: string;
  name: string;
  status: 'pending' | 'active' | 'paused' | 'terminating' | 'terminated' | 'error';
  health?: number;
  lastActivity?: string;
  parentAgentId?: string;
}

export interface AgentListResponse {
  ok: boolean;
  agents: AgentStatus[];
  total: number;
  limit: number;
  offset: number;
}

export interface SpawnAgentRequest {
  name: string;
  description?: string;
  swarmRole?: 'worker' | 'manager' | 'infrastructure';
  capabilities?: Record<string, unknown>;
  initialBudgetUsd?: number;
}

export interface SpawnAgentResponse {
  ok: boolean;
  agent: AgentStatus;
  apiKey: string;
}

export interface SimulationConfig {
  id?: string;
  name: string;
  iterations: number;
  duration?: number;
  failureMode?: 'none' | 'random' | 'targeted' | 'cascade';
  failureRate?: number;
  stressLevel?: 'low' | 'medium' | 'high' | 'extreme';
  latencyMs?: number;
  errorRate?: number;
  enableMonteCarlo?: boolean;
  saveResults?: boolean;
}

export interface SimulationResult {
  id: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'aborted';
  config: SimulationConfig;
  startedAt: string;
  completedAt?: string;
  metrics?: {
    totalExecutions: number;
    successfulExecutions: number;
    failedExecutions: number;
    averageLatencyMs: number;
    p50LatencyMs: number;
    p95LatencyMs: number;
    p99LatencyMs: number;
    throughput: number;
    costUsd: number;
  };
  monteCarloResults?: {
    mean: number;
    stdDev: number;
    confidenceInterval: [number, number];
    distribution: Record<string, number>;
  };
  error?: string;
}

export interface SimulationStatus {
  activeSimulation: SimulationResult | null;
  recentSimulations: SimulationResult[];
}

export interface GhostTask {
  id: string;
  title: string;
  description?: string;
  status: 'pending' | 'in_progress' | 'awaiting_approval' | 'approved' | 'rejected' | 'completed' | 'failed';
  phase?: string;
  started_at?: string;
  updated_at?: string;
  completed_at?: string;
  duration_ms?: number;
  logs?: GhostLogEntry[];
  artifacts?: GhostArtifact[];
  agent_id?: string;
  confidence?: number;
  dependencies?: string[];
  llm_output?: string;
}

export interface GhostLogEntry {
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
}

export interface GhostArtifact {
  name: string;
  type: string;
  path: string;
  size?: number;
}

export interface GhostLogEntry {
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  metadata?: Record<string, unknown>;
}

export interface GhostBuild {
  id: string;
  name: string;
  status: 'creating' | 'building' | 'ready' | 'failed';
  goal?: string;
  description?: string;
  phase?: 'planning' | 'provisioning' | 'building' | 'deploying' | 'monitoring' | 'complete' | 'error' | 'paused';
  progress?: number;
  tasks?: GhostTask[];
  taskId?: string;
  current_task_id?: string;
  human_approval_required?: boolean;
  approval_type?: 'schema' | 'deployment' | 'pr' | 'infra';
  error?: string;
  started_at?: string;
  updated_at?: string;
  createdAt: string;
  updatedAt: string;
}

export interface GhostCreateRequest {
  name: string;
  description?: string;
  autoApprobeThreshold?: 'low' | 'medium' | 'high';
}

export interface TelemetryMetrics {
  timestamp: string;
  requests: number;
  successfulRequests: number;
  failedRequests: number;
  averageLatencyMs: number;
  p50LatencyMs: number;
  p95LatencyMs: number;
  p99LatencyMs: number;
  tokenUsage?: {
    promptTokens: number;
    completionTokens: number;
    totalTokens: number;
    costUsd: number;
  };
  errorRate: number;
  throughput: number;
}

export interface TelemetryResponse {
  ok: boolean;
  metrics: TelemetryMetrics[];
  period: {
    start: string;
    end: string;
  };
  summary?: {
    totalRequests: number;
    averageLatencyMs: number;
    averageErrorRate: number;
    totalTokenUsage: number;
    totalCostUsd: number;
  };
}

export interface WorkflowNode {
  id: string;
  type: string;
  name: string;
  config: Record<string, unknown>;
  position: { x: number; y: number };
}

export interface WorkflowEdge {
  id: string;
  source: string;
  target: string;
  condition?: string;
}

export interface WorkflowGraph {
  id?: string;
  name: string;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  metadata?: Record<string, unknown>;
}

export interface WorkflowExecution {
  id: string;
  graphId: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  startedAt: string;
  completedAt?: string;
  result?: unknown;
  error?: string;
  nodeResults?: Array<{
    nodeId: string;
    status: 'success' | 'error' | 'skipped';
    output?: unknown;
    error?: string;
    durationMs: number;
  }>;
}

// ============================================================================
// Shared Types
// ============================================================================

export interface ListMemoriesResponse {
  memories: AgentMemory[];
  total: number;
  limit: number;
  offset: number;
}

export interface AgentMemory {
  id: string;
  tenant_id: string;
  agent_id: string;
  memory_type: string;
  content?: string;
  structured_data?: Record<string, unknown>;
  importance_score: number;
  access_count: number;
  last_accessed_at?: string;
  ttl_days: number;
  expires_at?: string;
  source_event_id?: string;
  r2_object_key?: string;
  r2_bucket?: string;
  is_offloaded: boolean;
  created_at: string;
  updated_at: string;
}

export interface TimelineEvent {
  id: string;
  timestamp: string;
  type: string;
  source: string;
  message: string;
  metadata?: Record<string, unknown>;
}

// ============================================================================
// Query Keys
// ============================================================================

// Helper to create readonly tuple keys
function studioKey(...parts: string[]) {
  return parts as readonly string[];
}

export const studioKeys = {
  all: ['studio'] as const,
  agents: () => ['studio', 'agents'] as const,
  agentList: (filters?: { limit?: number; offset?: number }) =>
    ['studio', 'agents', 'list', filters] as const,
  agent: (id: string) => ['studio', 'agents', 'detail', id] as const,
  simulation: () => ['studio', 'simulation'] as const,
  simulationStatus: () => ['studio', 'simulation', 'status'] as const,
  simulationResults: (id: string) => ['studio', 'simulation', 'results', id] as const,
  telemetry: (environment?: string) => ['studio', 'telemetry', environment ?? ''] as const,
  telemetryMetrics: (params?: { period?: string; environment?: string }) =>
    ['studio', 'telemetry', 'metrics', params] as const,
  workflow: (id?: string) => studioKey('studio', 'workflow', id ?? ''),
  workflowExecution: (id: string) => ['studio', 'workflow', 'execution', id] as const,
  timeline: (graphId: string) => ['studio', 'timeline', graphId] as const,
  ghost: () => ['studio', 'ghost'] as const,
  ghostBuilds: () => ['studio', 'ghost', 'builds'] as const,
  ghostTasks: () => ['studio', 'ghost', 'tasks'] as const,
  ghostTask: (id: string) => ['studio', 'ghost', 'task', id] as const,
  agentMemories: (agentId: string) => ['studio', 'agents', agentId, 'memories'] as const,
  runtimes: () => ['studio', 'runtimes'] as const,
};

// ============================================================================
// API Helpers
// ============================================================================

async function studioFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const environment = localStorage.getItem('ff-current-environment') || 'production';
  const token = localStorage.getItem('ff-access-token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Environment': environment,
    ...options?.headers,
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  return response.json();
}

// ============================================================================
// useStudioState - Main studio state management
// ============================================================================

export function useStudioState(initialState?: Partial<StudioState>) {
  const [selectedPanel, setSelectedPanel] = useState<string | null>(initialState?.selectedPanel ?? null);
  const [selectedTab, setSelectedTab] = useState<string | null>(initialState?.selectedTab ?? null);
  const [complexityLevel, setComplexityLevel] = useState<ComplexityLevel>(
    initialState?.complexityLevel ?? 'moderate'
  );
  const [zoomLevel, setZoomLevel] = useState(initialState?.zoomLevel ?? 1);
  const [viewMode, setViewMode] = useState<ViewMode>(initialState?.viewMode ?? 'design');
  const [isPanelCollapsed, setIsPanelCollapsed] = useState(initialState?.isPanelCollapsed ?? false);
  const [showMinimap, setShowMinimap] = useState(initialState?.showMinimap ?? true);
  const [gridEnabled, setGridEnabled] = useState(initialState?.gridEnabled ?? true);
  const [snapToGrid, setSnapToGrid] = useState(initialState?.snapToGrid ?? true);

  const state = useMemo<StudioState>(() => ({
    selectedPanel,
    selectedTab,
    complexityLevel,
    zoomLevel,
    viewMode,
    isPanelCollapsed,
    showMinimap,
    gridEnabled,
    snapToGrid,
  }), [
    selectedPanel,
    selectedTab,
    complexityLevel,
    zoomLevel,
    viewMode,
    isPanelCollapsed,
    showMinimap,
    gridEnabled,
    snapToGrid,
  ]);

  const zoomIn = useCallback(() => {
    setZoomLevel((z) => Math.min(z + 0.1, 3));
  }, []);

  const zoomOut = useCallback(() => {
    setZoomLevel((z) => Math.max(z - 0.1, 0.1));
  }, []);

  const resetZoom = useCallback(() => {
    setZoomLevel(1);
  }, []);

  const togglePanel = useCallback(() => {
    setIsPanelCollapsed((c) => !c);
  }, []);

  const toggleMinimap = useCallback(() => {
    setShowMinimap((m) => !m);
  }, []);

  const toggleGrid = useCallback(() => {
    setGridEnabled((g) => !g);
  }, []);

  const toggleSnapToGrid = useCallback(() => {
    setSnapToGrid((s) => !s);
  }, []);

  return {
    state,
    setSelectedPanel,
    setSelectedTab,
    setComplexityLevel,
    setZoomLevel,
    setViewMode,
    zoomIn,
    zoomOut,
    resetZoom,
    togglePanel,
    toggleMinimap,
    toggleGrid,
    toggleSnapToGrid,
  };
}

// ============================================================================
// useStudioAgents - Agent management hook
// ============================================================================

export function useStudioAgents(filters?: { limit?: number; offset?: number }) {
  const queryClient = useQueryClient();

  const agents = useQuery({
    queryKey: studioKeys.agentList(filters),
    queryFn: () =>
      studioFetch<AgentListResponse>('/v1/agent', {
        headers: { 'Content-Type': 'application/json' },
      }),
    staleTime: 1000 * 30,
    refetchInterval: 15000,
  });

  const spawnAgent = useMutation({
    mutationFn: (data: SpawnAgentRequest) =>
      studioFetch<SpawnAgentResponse>('/v1/agent/spawn', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.agents() });
      toast.success('Agent spawned successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to spawn agent: ${error.message}`);
    },
  });

  const pauseAgent = useMutation({
    mutationFn: (agentId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/agent/${agentId}/lifecycle/pause`, {
        method: 'PUT',
        body: JSON.stringify({}),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.agents() });
      toast.success('Agent paused');
    },
    onError: (error: Error) => {
      toast.error(`Failed to pause agent: ${error.message}`);
    },
  });

  const resumeAgent = useMutation({
    mutationFn: (agentId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/agent/${agentId}/lifecycle/resume`, {
        method: 'PUT',
        body: JSON.stringify({}),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.agents() });
      toast.success('Agent resumed');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resume agent: ${error.message}`);
    },
  });

  const terminateAgent = useMutation({
    mutationFn: (agentId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/agent/${agentId}/lifecycle/terminate`, {
        method: 'POST',
        body: JSON.stringify({ grace_period_seconds: 30, reason: 'user_terminated' }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.agents() });
      toast.success('Agent termination initiated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to terminate agent: ${error.message}`);
    },
  });

  const refreshAgents = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: studioKeys.agents() });
  }, [queryClient]);

  return {
    agents: agents.data?.agents ?? [],
    totalAgents: agents.data?.total ?? 0,
    isLoading: agents.isLoading,
    isError: agents.isError,
    error: agents.error,
    spawnAgent,
    pauseAgent,
    resumeAgent,
    terminateAgent,
    refreshAgents,
  };
}

export function useStudioAgent(agentId: string) {
  return useQuery({
    queryKey: studioKeys.agent(agentId),
    queryFn: () => studioFetch<{ ok: boolean; agent: AgentStatus }>(`/v1/agent/${agentId}`),
    enabled: !!agentId,
    staleTime: 1000 * 10,
    refetchInterval: 5000,
  });
}

// ============================================================================
// useStudioMemory - Agent memory management hook
// ============================================================================

export function useStudioMemory(agentId?: string) {
  const queryClient = useQueryClient();

  const memories = useQuery<ListMemoriesResponse>({
    queryKey: studioKeys.agentMemories(agentId || 'all'),
    queryFn: () => {
      const params = new URLSearchParams();
      if (agentId) params.set('agent_id', agentId);
      params.set('limit', '50');
      return studioFetch<ListMemoriesResponse>(`/v1/agent-memories?${params.toString()}`);
    },
    enabled: true,
    staleTime: 1000 * 30,
  });

  const addMemory = useMutation({
    mutationFn: (data: { content: string; memory_type: string; importance_score?: number }) =>
      studioFetch<{ memory: AgentMemory }>('/v1/agent-memories', {
        method: 'POST',
        body: JSON.stringify({ agent_id: agentId, ...data }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.agentMemories(agentId || 'all') });
      toast.success('Memory added');
    },
    onError: (error: Error) => {
      toast.error(`Failed to add memory: ${error.message}`);
    },
  });

  const deleteMemory = useMutation({
    mutationFn: (memoryId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/agent-memories/${memoryId}`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.agentMemories(agentId || 'all') });
      toast.success('Memory deleted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete memory: ${error.message}`);
    },
  });

  const searchMemories = useMutation({
    mutationFn: (query: string) =>
      studioFetch<ListMemoriesResponse>('/v1/agent-memories/search', {
        method: 'POST',
        body: JSON.stringify({ agent_id: agentId, query, limit: 20 }),
      }),
  });

  return {
    memories: memories.data?.memories ?? [],
    total: memories.data?.total ?? 0,
    isLoading: memories.isLoading,
    isError: memories.isError,
    error: memories.error,
    addMemory,
    deleteMemory,
    searchMemories,
    refreshMemories: () => queryClient.invalidateQueries({ queryKey: studioKeys.agentMemories(agentId || 'all') }),
  };
}

// ============================================================================
// useStudioSimulation - R-Sim simulation hook
// Bridges StudioPage canvas state to Go /v1/simulate/* backend.
// Phase 3.2: Wire frontend to Go simulation engine.
// ============================================================================

export interface WorkflowNode {
  id: string;
  type: string;
  name: string;
  config: Record<string, unknown>;
  position: { x: number; y: number };
}

export interface WorkflowEdge {
  id: string;
  source: string;
  target: string;
  condition?: string;
}

export interface WorkflowGraph {
  id?: string;
  name: string;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  metadata?: Record<string, unknown>;
}

/** Go SimulationResult (from /v1/simulate/workflow) */
export interface GoSimulationResult {
  id: string;
  status: string;
  success_rate: number;
  avg_latency_ms: number;
  avg_cost_usd: number;
  predicted_node_executions: Record<string, number>;
  failed_nodes: Record<string, number>;
  execution_count: number;
  started_at: string;
  completed_at?: string;
  iteration: number;
}

/** Go MonteCarloResult (from /v1/simulate/monte-carlo) */
export interface GoMonteCarloResult {
  id: string;
  iterations: number;
  success_rate: number;
  partial_failure_rate: number;
  total_failure_rate: number;
  avg_latency_ms: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  avg_cost_usd: number;
  outcomes: GoOutcomeSample[];
  bottleneck_nodes: string[];
  cost_breakdown: Record<string, number>;
}

export interface GoOutcomeSample {
  outcome: 'success' | 'partial' | 'failed';
  probability: number;
  latency_ms: number;
  cost_usd: number;
  failed_nodes: string[];
  risk_factors: string[];
}

/** Go ExecutionForecast (from /v1/forecast/execution) */
export interface GoExecutionForecast {
  workflow_id: string;
  time_horizon: string;
  predicted_executions: number;
  success_rate: number;
  avg_latency_ms: number;
  cost_usd: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  predictions: GoPredictionPoint[];
}

export interface GoPredictionPoint {
  timestamp: string;
  executions: number;
  success_rate: number;
}

/** Go CostForecast (from /v1/forecast/cost) */
export interface GoCostForecast {
  workflow_id: string;
  total_cost_usd: number;
  per_call_usd: number;
  lower_bound_usd: number;
  upper_bound_usd: number;
  confidence: number;
  by_node: Record<string, number>;
}

/** Go LatencyForecast (from /v1/forecast/latency) */
export interface GoLatencyForecast {
  workflow_id: string;
  load_level: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  avg_latency_ms: number;
}

/** Go StressTestResult (from /v1/stress-test/{id}) */
export interface GoStressTestResult {
  id: string;
  status: string;
  iterations: number;
  total_executions: number;
  success_rate: number;
  throughput: number;
  latency_p50: number;
  latency_p95: number;
  latency_p99: number;
  errors: GoErrorBreakdown[];
}

export interface GoErrorBreakdown {
  type: string;
  count: number;
  percentage: number;
}

/** Go CollisionResult (from /v1/detect/collisions) */
export interface GoCollisionResult {
  collisions: GoCollision[];
  resolutions: string[];
}

export interface GoCollision {
  resource_id: string;
  resource_name: string;
  severity: 'high' | 'medium' | 'low';
  conflicting_tasks: GoTaskSchedule[];
  resolution: string;
}

export interface GoTaskSchedule {
  task_id: string;
  start_ms: number;
  end_ms: number;
  usage: number;
  priority: number;
}

/** Go AgentBehaviorPrediction (from /v1/predict/agent-behavior) */
export interface GoAgentBehaviorPrediction {
  agent_id: string;
  confidence: number;
  based_on_samples: number;
  likely_actions: GoActionPrediction[];
  current_task_prediction: string;
}

export interface GoActionPrediction {
  action: string;
  probability: number;
  expected_outcome: string;
  risk_level: 'low' | 'medium' | 'high';
}

/** Go HallucinationRiskResult (from /v1/analyze/hallucination-risk) */
export interface GoHallucinationRiskResult {
  model_id: string;
  risk_score: number;
  risk_level: 'low' | 'medium' | 'high' | 'critical';
  contributing_factors: string[];
  recommendations: string[];
  confidence: number;
}

/** Convert Studio WorkflowGraph to Go WorkflowSpec */
function graphToWorkflowSpec(graph: WorkflowGraph): {
  nodes: Array<{ id: string; type: string; timeout_ms: number; cost_usd: number; metadata?: Record<string, unknown> }>;
  edges: Array<{ from: string; to: string; probability_success: number }>;
} {
  return {
    nodes: graph.nodes.map((n) => ({
      id: n.id,
      type: n.type || 'function',
      timeout_ms: (n.config?.timeout_ms as number) || 5000,
      cost_usd: (n.config?.cost_usd as number) || 0.001,
      metadata: n.config,
    })),
    edges: graph.edges.map((e) => ({
      from: e.source,
      to: e.target,
      probability_success: 0.95,
    })),
  };
}

export function useStudioSimulation() {
  const queryClient = useQueryClient();

  const [simulationState, setSimulationState] = useState<{
    activeSimulation: SimulationResult | null;
    recentSimulations: SimulationResult[];
  }>({ activeSimulation: null, recentSimulations: [] });

  /** Map Go result → local SimulationResult for StudioPanel */
  const mapGoResult = (go: GoSimulationResult): SimulationResult => ({
    id: go.id,
    status: go.status === 'completed' ? 'completed'
      : go.status === 'running' ? 'running'
      : go.status === 'aborted' ? 'aborted'
      : 'pending',
    config: {
      name: 'Simulation',
      iterations: go.iteration,
    },
    startedAt: go.started_at,
    completedAt: go.completed_at,
    metrics: {
      totalExecutions: go.execution_count,
      successfulExecutions: Math.round(go.success_rate * go.execution_count),
      failedExecutions: Math.round((1 - go.success_rate) * go.execution_count),
      averageLatencyMs: go.avg_latency_ms,
      p50LatencyMs: go.avg_latency_ms,
      p95LatencyMs: Math.round(go.avg_latency_ms * 1.5),
      p99LatencyMs: Math.round(go.avg_latency_ms * 2),
      throughput: 0,
      costUsd: go.avg_cost_usd,
    },
  });

  const startSimulation = useMutation({
    mutationFn: async (params: { graph: WorkflowGraph; iterations?: number }) => {
      const spec = graphToWorkflowSpec(params.graph);
      const iterations = params.iterations ?? 100;
      const res = await studioFetch<{ ok: boolean; result: GoSimulationResult }>('/v1/simulate/workflow', {
        method: 'POST',
        body: JSON.stringify({ workflow: spec, iterations }),
      });
      return res.result;
    },
    onSuccess: (go) => {
      const result = mapGoResult(go);
      setSimulationState((prev) => ({
        activeSimulation: result,
        recentSimulations: [result, ...prev.recentSimulations].slice(0, 20),
      }));
      queryClient.invalidateQueries({ queryKey: studioKeys.simulationStatus() });
      toast.success('Simulation started');
    },
    onError: (error: Error) => {
      toast.error(`Failed to start simulation: ${error.message}`);
    },
  });

  const stopSimulation = useMutation({
    mutationFn: (simulationId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/simulate/workflow/${simulationId}/abort`, {
        method: 'POST',
      }),
    onSuccess: () => {
      setSimulationState((prev) => ({
        ...prev,
        activeSimulation: prev.activeSimulation
          ? { ...prev.activeSimulation, status: 'aborted' }
          : null,
      }));
      queryClient.invalidateQueries({ queryKey: studioKeys.simulationStatus() });
      toast.success('Simulation stopped');
    },
    onError: (error: Error) => {
      toast.error(`Failed to stop simulation: ${error.message}`);
    },
  });

  const abortSimulation = useMutation({
    mutationFn: (simulationId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/simulate/workflow/${simulationId}/abort`, {
        method: 'POST',
      }),
    onSuccess: () => {
      setSimulationState((prev) => ({
        ...prev,
        activeSimulation: prev.activeSimulation
          ? { ...prev.activeSimulation, status: 'aborted' }
          : null,
      }));
      queryClient.invalidateQueries({ queryKey: studioKeys.simulationStatus() });
      toast.success('Simulation aborted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to abort simulation: ${error.message}`);
    },
  });

  const getSimulationResults = useMutation({
    mutationFn: (simulationId: string) =>
      studioFetch<{ ok: boolean; result: GoSimulationResult }>(`/v1/simulate/workflow/${simulationId}`),
    onSuccess: (data) => {
      const result = mapGoResult(data.result);
      setSimulationState((prev) => ({
        ...prev,
        activeSimulation: result,
      }));
    },
  });

  const runMonteCarlo = useMutation({
    mutationFn: async (params: { graph: WorkflowGraph; iterations?: number }) => {
      const spec = graphToWorkflowSpec(params.graph);
      const iterations = params.iterations ?? 1000;
      const res = await studioFetch<{ ok: boolean; result: GoMonteCarloResult }>('/v1/simulate/monte-carlo', {
        method: 'POST',
        body: JSON.stringify({ workflow: spec, iterations }),
      });
      return res.result;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.simulationStatus() });
      toast.success('Monte Carlo simulation completed');
    },
    onError: (error: Error) => {
      toast.error(`Failed to run Monte Carlo simulation: ${error.message}`);
    },
  });

  const injectFailure = useMutation({
    mutationFn: (params: { simulationId?: string; nodeId?: string; failureType: string }) =>
      studioFetch<{ ok: boolean }>('/v1/simulate/failure-inject', {
        method: 'POST',
        body: JSON.stringify({
          workflow_id: params.simulationId || '',
          nodes: params.nodeId
            ? [{ node_id: params.nodeId, failure_type: params.failureType, failure_rate: 0.5, recovery_action: 'retry' }]
            : [],
        }),
      }),
    onSuccess: () => {
      toast.success('Failure injected');
    },
    onError: (error: Error) => {
      toast.error(`Failed to inject failure: ${error.message}`);
    },
  });

  const runStressTest = useMutation({
    mutationFn: (params: { graph: WorkflowGraph; iterations?: number; parallelism?: number; stressLevel?: string }) =>
      studioFetch<{ ok: boolean; id: string }>('/v1/stress-test/start', {
        method: 'POST',
        body: JSON.stringify({
          iterations: params.iterations ?? 1000,
          parallelism: params.parallelism ?? 10,
          workflow_id: params.graph.id || 'default',
          load_profile: 'constant',
        }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.simulationStatus() });
      toast.success('Stress test started');
    },
    onError: (error: Error) => {
      toast.error(`Failed to run stress test: ${error.message}`);
    },
  });

  const getStressTest = useMutation({
    mutationFn: (id: string) =>
      studioFetch<{ ok: boolean; result: GoStressTestResult }>(`/v1/stress-test/${id}`),
  });

  const getExecutionForecast = useMutation({
    mutationFn: (params: { graph: WorkflowGraph; timeHorizon?: string; callVolume?: number }) =>
      studioFetch<{ ok: boolean; forecast: GoExecutionForecast }>('/v1/forecast/execution', {
        method: 'POST',
        body: JSON.stringify({
          workflow_id: params.graph.id || 'default',
          time_horizon: params.timeHorizon || '24h',
          call_volume: params.callVolume ?? 100,
          nodes: graphToWorkflowSpec(params.graph).nodes,
        }),
      }),
  });

  const getCostForecast = useMutation({
    mutationFn: (params: { graph: WorkflowGraph; callVolume?: number }) =>
      studioFetch<{ ok: boolean; forecast: GoCostForecast }>('/v1/forecast/cost', {
        method: 'POST',
        body: JSON.stringify({
          workflow_id: params.graph.id || 'default',
          nodes: graphToWorkflowSpec(params.graph).nodes,
          call_volume: params.callVolume ?? 100,
        }),
      }),
  });

  const getLatencyForecast = useMutation({
    mutationFn: (params: { graph: WorkflowGraph; loadLevel?: number }) =>
      studioFetch<{ ok: boolean; forecast: GoLatencyForecast }>('/v1/forecast/latency', {
        method: 'POST',
        body: JSON.stringify({
          workflow_id: params.graph.id || 'default',
          nodes: graphToWorkflowSpec(params.graph).nodes,
          load_level: params.loadLevel ?? 0.5,
        }),
      }),
  });

  const detectCollisions = useMutation({
    mutationFn: (params: { resources: Array<{ id: string; name: string; type: string; capacity: number; tasks: GoTaskSchedule[] }> }) =>
      studioFetch<{ ok: boolean; result: GoCollisionResult }>('/v1/detect/collisions', {
        method: 'POST',
        body: JSON.stringify({ resources: params.resources }),
      }),
  });

  const predictAgentBehavior = useMutation({
    mutationFn: (params: { agentId: string; historySize?: number; context?: string }) =>
      studioFetch<{ ok: boolean; prediction: GoAgentBehaviorPrediction }>('/v1/predict/agent-behavior', {
        method: 'POST',
        body: JSON.stringify({
          agent_id: params.agentId,
          history_size: params.historySize ?? 50,
          context: params.context || 'workflow execution',
        }),
      }),
  });

  const analyzeHallucinationRisk = useMutation({
    mutationFn: (params: { modelId: string; promptLength?: number; contextWindow?: number; taskType?: string; complexity?: string; previousErrors?: number }) =>
      studioFetch<{ ok: boolean; result: GoHallucinationRiskResult }>('/v1/analyze/hallucination-risk', {
        method: 'POST',
        body: JSON.stringify({
          model_id: params.modelId,
          prompt_length: params.promptLength ?? 1000,
          context_window: params.contextWindow ?? 128000,
          task_type: params.taskType || 'code_gen',
          complexity: params.complexity || 'medium',
          previous_errors: params.previousErrors ?? 0,
        }),
      }),
  });

  return {
    activeSimulation: simulationState.activeSimulation,
    recentSimulations: simulationState.recentSimulations,
    isLoading: false,
    isError: false,
    error: null,
    startSimulation,
    stopSimulation,
    abortSimulation,
    getSimulationResults,
    runMonteCarlo,
    injectFailure,
    runStressTest,
    getStressTest,
    getExecutionForecast,
    getCostForecast,
    getLatencyForecast,
    detectCollisions,
    predictAgentBehavior,
    analyzeHallucinationRisk,
  };
}

// ============================================================================
// useStudioGhost - Ghost Mode autonomous building hook
// ============================================================================

export function useStudioGhost() {
  const queryClient = useQueryClient();

  const builds = useQuery({
    queryKey: studioKeys.ghostBuilds(),
    queryFn: () => studioFetch<{ ok: boolean; builds: GhostBuild[] }>('/v1/ghost/builds'),
    staleTime: 1000 * 10,
    refetchInterval: 10000,
  });

  // Tasks are scoped to builds - query uses first build if available
  const tasks = useQuery({
    queryKey: studioKeys.ghostTasks(),
    queryFn: async () => {
      // Get tasks from all builds or specific build
      if (builds.data?.builds && builds.data.builds.length > 0) {
        const firstBuild = builds.data.builds[0];
        return studioFetch<{ ok: boolean; tasks: GhostTask[] }>(`/v1/ghost/builds/${firstBuild.id}/tasks`);
      }
      return { ok: true, tasks: [] as GhostTask[] };
    },
    staleTime: 1000 * 5,
    refetchInterval: 5000,
  });

  const createBuild = useMutation({
    mutationFn: (data: GhostCreateRequest) =>
      studioFetch<{ ok: boolean; build: GhostBuild }>('/v1/ghost/builds', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.ghostBuilds() });
      toast.success('Build created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create build: ${error.message}`);
    },
  });

  const cancelBuild = useMutation({
    mutationFn: (buildId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/ghost/builds/${buildId}/cancel`, {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.ghostBuilds() });
      toast.success('Build cancelled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to cancel build: ${error.message}`);
    },
  });

  const approveTask = useMutation({
    mutationFn: (params: { buildId: string; taskId: string }) =>
      studioFetch<{ ok: boolean; task: GhostTask }>(`/v1/ghost/builds/${params.buildId}/approve`, {
        method: 'POST',
        body: JSON.stringify({ task_id: params.taskId }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.ghostTasks() });
      toast.success('Task approved');
    },
    onError: (error: Error) => {
      toast.error(`Failed to approve task: ${error.message}`);
    },
  });

  const rejectTask = useMutation({
    mutationFn: (params: { buildId: string; taskId: string; reason?: string }) =>
      studioFetch<{ ok: boolean; task: GhostTask }>(`/v1/ghost/builds/${params.buildId}/approve`, {
        method: 'POST',
        body: JSON.stringify({ task_id: params.taskId, decision: 'reject', notes: params.reason }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.ghostTasks() });
      toast.success('Task rejected');
    },
    onError: (error: Error) => {
      toast.error(`Failed to reject task: ${error.message}`);
    },
  });

  // Task logs are fetched from the build details endpoint
  const getTaskLogs = useQuery({
    queryKey: studioKeys.ghostTasks(),
    queryFn: async () => {
      if (builds.data?.builds && builds.data.builds.length > 0) {
        const firstBuild = builds.data.builds[0];
        const response = await studioFetch<{ ok: boolean; logs: GhostLogEntry[] }>(
          `/v1/ghost/builds/${firstBuild.id}/logs`
        );
        return response.logs ?? [];
      }
      return [] as GhostLogEntry[];
    },
    staleTime: 1000 * 2,
    refetchInterval: 5000,
    enabled: !!builds.data?.builds?.length,
  });

  // Fetch task-specific logs for a given build and task
  const getTaskLogsForTask = useMutation({
    mutationFn: async ({ buildId, taskId }: { buildId: string; taskId: string }) => {
      const response = await studioFetch<{ ok: boolean; logs: GhostLogEntry[] }>(
        `/v1/ghost/builds/${buildId}/tasks/${taskId}/logs`
      );
      return response.logs ?? [];
    },
  });

  // Fetch build-level logs
  const getBuildLogs = useMutation({
    mutationFn: async (buildId: string) => {
      const response = await studioFetch<{ ok: boolean; logs: GhostLogEntry[] }>(
        `/v1/ghost/builds/${buildId}/logs`
      );
      return response.logs ?? [];
    },
  });

  return {
    builds: builds.data?.builds ?? [],
    tasks: tasks.data?.tasks ?? [],
    isLoadingBuilds: builds.isLoading,
    isLoadingTasks: tasks.isLoading,
    isError: builds.isError || tasks.isError,
    error: builds.error || tasks.error,
    createBuild,
    cancelBuild,
    approveTask,
    rejectTask,
    taskLogs: getTaskLogs.data ?? [],
    getTaskLogsForTask,
    getBuildLogs,
  };
}

// ============================================================================
// useStudioTelemetry - Live telemetry hook
// ============================================================================

export function useStudioTelemetry(params?: { period?: string; environment?: string }) {
  const environment = params?.environment || localStorage.getItem('ff-current-environment') || 'production';
  const metrics = useQuery<{ metrics: TelemetryMetrics[] }>({
    queryKey: studioKeys.telemetryMetrics({ ...params, environment }),
    queryFn: () =>
      studioFetch<{ metrics: TelemetryMetrics[] }>('/v1/studio/telemetry?hours=24', {
        headers: { 'Content-Type': 'application/json', 'X-Environment': environment },
      }),
    staleTime: 1000 * 10,
    refetchInterval: 10000,
  });

  const tokenUsage = useMemo(() => {
    const allMetrics = metrics.data?.metrics ?? [];
    return allMetrics.reduce(
      (acc, m) => {
        if (m.tokenUsage) {
          acc.promptTokens += m.tokenUsage.promptTokens;
          acc.completionTokens += m.tokenUsage.completionTokens;
          acc.totalTokens += m.tokenUsage.totalTokens;
          acc.costUsd += m.tokenUsage.costUsd;
        }
        return acc;
      },
      { promptTokens: 0, completionTokens: 0, totalTokens: 0, costUsd: 0 }
    );
  }, [metrics.data?.metrics]);

  const latencyStats = useMemo(() => {
    const allMetrics = metrics.data?.metrics ?? [];
    if (allMetrics.length === 0) {
      return { avg: 0, p50: 0, p95: 0, p99: 0 };
    }
    return {
      avg: allMetrics.reduce((sum, m) => sum + m.averageLatencyMs, 0) / allMetrics.length,
      p50: allMetrics[allMetrics.length - 1]?.p50LatencyMs ?? 0,
      p95: allMetrics[allMetrics.length - 1]?.p95LatencyMs ?? 0,
      p99: allMetrics[allMetrics.length - 1]?.p99LatencyMs ?? 0,
    };
  }, [metrics.data?.metrics]);

  const errorRateStats = useMemo(() => {
    const allMetrics = metrics.data?.metrics ?? [];
    if (allMetrics.length === 0) return { avg: 0, max: 0 };
    return {
      avg: allMetrics.reduce((sum, m) => sum + m.errorRate, 0) / allMetrics.length,
      max: Math.max(...allMetrics.map((m) => m.errorRate)),
    };
  }, [metrics.data?.metrics]);

  return {
    metrics: metrics.data?.metrics ?? [],
    summary: (metrics.data as any)?.summary,
    period: (metrics.data as any)?.period,
    tokenUsage,
    latencyStats,
    errorRateStats,
    isLoading: metrics.isLoading,
    isError: metrics.isError,
    error: metrics.error,
  };
}

// ============================================================================
// useStudioWorkflow - Workflow/graph management hook
// ============================================================================

export function useStudioWorkflow() {
  const queryClient = useQueryClient();

  const workflowGraph = useQuery({
    queryKey: studioKeys.workflow(),
    queryFn: () => studioFetch<{ ok: boolean; graph: WorkflowGraph }>('/v1/workflow/graph'),
    staleTime: 1000 * 30,
  });

  const executions = useQuery({
    queryKey: studioKeys.workflow(),
    queryFn: () => studioFetch<{ ok: boolean; executions: WorkflowExecution[] }>('/v1/workflow/executions'),
    staleTime: 1000 * 10,
    refetchInterval: 5000,
  });

  const createNode = useMutation({
    mutationFn: (node: Omit<WorkflowNode, 'id'>) =>
      studioFetch<{ ok: boolean; node: WorkflowNode }>('/v1/workflow/nodes', {
        method: 'POST',
        body: JSON.stringify(node),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.workflow() });
      toast.success('Node created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create node: ${error.message}`);
    },
  });

  const updateNode = useMutation({
    mutationFn: (params: { nodeId: string; updates: Partial<WorkflowNode> }) =>
      studioFetch<{ ok: boolean; node: WorkflowNode }>(`/v1/workflow/nodes/${params.nodeId}`, {
        method: 'PUT',
        body: JSON.stringify(params.updates),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.workflow() });
      toast.success('Node updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update node: ${error.message}`);
    },
  });

  const deleteNode = useMutation({
    mutationFn: (nodeId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/workflow/nodes/${nodeId}`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.workflow() });
      toast.success('Node deleted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete node: ${error.message}`);
    },
  });

  const createEdge = useMutation({
    mutationFn: (edge: Omit<WorkflowEdge, 'id'>) =>
      studioFetch<{ ok: boolean; edge: WorkflowEdge }>('/v1/workflow/edges', {
        method: 'POST',
        body: JSON.stringify(edge),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.workflow() });
      toast.success('Edge created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create edge: ${error.message}`);
    },
  });

  const updateEdge = useMutation({
    mutationFn: (params: { edgeId: string; updates: Partial<WorkflowEdge> }) =>
      studioFetch<{ ok: boolean; edge: WorkflowEdge }>(`/v1/workflow/edges/${params.edgeId}`, {
        method: 'PUT',
        body: JSON.stringify(params.updates),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.workflow() });
      toast.success('Edge updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update edge: ${error.message}`);
    },
  });

  const deleteEdge = useMutation({
    mutationFn: (edgeId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/workflow/edges/${edgeId}`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.workflow() });
      toast.success('Edge deleted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete edge: ${error.message}`);
    },
  });

  const executeWorkflow = useMutation({
    mutationFn: (graph: WorkflowGraph) =>
      studioFetch<WorkflowExecution>('/v1/workflow/execute', {
        method: 'POST',
        body: JSON.stringify({ graph }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.workflow() });
      toast.success('Workflow execution started');
    },
    onError: (error: Error) => {
      toast.error(`Failed to execute workflow: ${error.message}`);
    },
  });

  const cancelExecution = useMutation({
    mutationFn: (executionId: string) =>
      studioFetch<{ ok: boolean }>(`/v1/workflow/executions/${executionId}/cancel`, {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: studioKeys.workflow() });
      toast.success('Execution cancelled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to cancel execution: ${error.message}`);
    },
  });

  // Get execution by ID - use mutation for on-demand fetching
  const getExecution = useMutation({
    mutationFn: (executionId: string) =>
      studioFetch<{ ok: boolean; execution: WorkflowExecution }>(`/v1/workflow/executions/${executionId}`),
  });

  // Get timeline by graph ID - use mutation for on-demand fetching
  const timeline = useMutation({
    mutationFn: (graphId: string) =>
      studioFetch<{ ok: boolean; events: TimelineEvent[] }>(`/v1/workflow/${graphId}/timeline`),
  });

  // ── Studio Code Editor ─────────────────────────────────────────────────────
  const formatCode = useMutation({
    mutationFn: (params: { code: string; language: string; file_path: string; options?: Record<string, unknown> }) =>
      studioFetch<{ formatted: string; version: number; action: string }>('/v1/studio/code/format', {
        method: 'POST',
        body: JSON.stringify(params),
      }),
    onSuccess: () => {
      toast.success('Code formatted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to format code: ${error.message}`);
    },
  });

  const saveCode = useMutation({
    mutationFn: (params: { code: string; file_path: string; metadata?: Record<string, unknown> }) =>
      studioFetch<{ version: number; timestamp: string; action: string }>('/v1/studio/code/save', {
        method: 'POST',
        body: JSON.stringify(params),
      }),
    onSuccess: (data) => {
      toast.success(`Saved (version ${data.version})`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to save code: ${error.message}`);
    },
  });

  const undoCode = useMutation({
    mutationFn: (params: { file_path: string; current_version: number }) =>
      studioFetch<{ code: string; version: number; action: string; available: boolean; metadata?: Record<string, unknown> }>('/v1/studio/code/undo', {
        method: 'POST',
        body: JSON.stringify(params),
      }),
    onSuccess: (data) => {
      if (data.available) {
        toast.success(`Undone to version ${data.version}`);
      }
    },
    onError: (error: Error) => {
      toast.error(`Failed to undo: ${error.message}`);
    },
  });

  const redoCode = useMutation({
    mutationFn: (params: { file_path: string; current_version: number }) =>
      studioFetch<{ code: string; version: number; action: string; available: boolean; metadata?: Record<string, unknown> }>('/v1/studio/code/redo', {
        method: 'POST',
        body: JSON.stringify(params),
      }),
    onSuccess: (data) => {
      if (data.available) {
        toast.success(`Redone to version ${data.version}`);
      }
    },
    onError: (error: Error) => {
      toast.error(`Failed to redo: ${error.message}`);
    },
  });

  const getVersionHistory = useMutation({
    mutationFn: (filePath: string) =>
      studioFetch<{ versions: Array<{ version: number; action: string; created_at: string }>; total: number }>(
        `/v1/studio/code/history?file_path=${encodeURIComponent(filePath)}`
      ),
  });

  return {
    graph: workflowGraph.data?.graph,
    executions: executions.data?.executions ?? [],
    isLoadingGraph: workflowGraph.isLoading,
    isLoadingExecutions: executions.isLoading,
    isError: workflowGraph.isError || executions.isError,
    error: workflowGraph.error || executions.error,
    createNode,
    updateNode,
    deleteNode,
    createEdge,
    updateEdge,
    deleteEdge,
    executeWorkflow,
    cancelExecution,
    getExecution,
    timeline,
    // Studio code editor
    formatCode,
    saveCode,
    undoCode,
    redoCode,
    getVersionHistory,
  };
}

// ============================================================================
// useStudioRuntimes - Runtime selection hook
// ============================================================================

export interface StudioRuntimeInfo {
  id: string;
  name: string;
  version: string;
  status: string;
  features: string[];
  memoryLimit: number;
  timeout: number;
  type?: string;
  provider?: string;
  region?: string;
  supportedLanguages?: string[];
  maxMemory?: number;
  maxTimeout?: number;
  latency?: number;
  costPerExecution?: number;
  reliabilityScore?: number;
}

export interface RuntimeListResponse {
  runtimes: StudioRuntimeInfo[];
}

export function useStudioRuntimes() {
  const runtimes = useQuery<RuntimeListResponse>({
    queryKey: studioKeys.runtimes(),
    queryFn: () => studioFetch<RuntimeListResponse>('/v1/runtimes'),
    staleTime: 1000 * 60 * 5,
  });

  const transformedRuntimes = useMemo(() => {
    return (runtimes.data?.runtimes ?? []).map((runtime) => ({
      id: runtime.id,
      name: runtime.name,
      type: runtime.type ?? 'local' as const,
      provider: runtime.provider ?? 'functionfly',
      region: runtime.region ?? 'global',
      status: runtime.status === 'stable' ? 'online' as const : 'online' as const,
      supportedLanguages: runtime.supportedLanguages ?? extractLanguagesFromFeatures(runtime.features),
      maxMemory: runtime.maxMemory ?? runtime.memoryLimit,
      maxTimeout: runtime.maxTimeout ?? runtime.timeout,
      features: runtime.features,
      latency: runtime.latency ?? 5,
      costPerExecution: runtime.costPerExecution ?? 0.002,
      reliabilityScore: runtime.reliabilityScore ?? 95,
    }));
  }, [runtimes.data?.runtimes]);

  return {
    runtimes: transformedRuntimes,
    isLoading: runtimes.isLoading,
    isError: runtimes.isError,
    error: runtimes.error,
  };
}

function extractLanguagesFromFeatures(features: string[]): string[] {
  const languageMap: Record<string, string[]> = {
    'TypeScript': ['typescript'],
    'JavaScript': ['javascript'],
    'Python': ['python'],
    'Rust': ['rust'],
    'Go': ['go'],
    'C': ['c'],
    'C++': ['cpp'],
    'Ruby': ['ruby'],
    'Kotlin': ['kotlin'],
    'Swift': ['swift'],
  };
  const languages: string[] = [];
  for (const feature of features) {
    for (const [key, value] of Object.entries(languageMap)) {
      if (feature.includes(key) && !languages.includes(value[0])) {
        languages.push(value[0]);
      }
    }
  }
  return languages.length > 0 ? languages : ['javascript', 'typescript'];
}
