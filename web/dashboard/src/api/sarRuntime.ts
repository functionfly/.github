/**
 * SAR Runtime API Client
 * Direct HTTP client for the Rust-based Semantic Action Runtime (port 8082)
 */

import axios from 'axios';

const SAR_RUNTIME_URL = import.meta.env.VITE_SAR_RUNTIME_URL;
if (!SAR_RUNTIME_URL) {
  throw new Error('VITE_SAR_RUNTIME_URL environment variable is required');
}

// Create a separate axios instance for SAR runtime (no auth required, direct connection)
const sarClient = axios.create({
  baseURL: SAR_RUNTIME_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000, // 30s timeout for graph execution
});

// ==================== Types ====================

export interface GraphDefinition {
  id: string;
  nodes: GraphNode[];
  edges: GraphEdge[];
  metadata?: GraphMetadata;
}

export interface GraphNode {
  id: string;
  type: 'llm' | 'tool' | 'memory' | 'action' | 'input' | 'output' | 'condition' | 'loop';
  name: string;
  config: Record<string, unknown>;
  position?: { x: number; y: number };
}

export interface GraphEdge {
  id: string;
  source: string;
  target: string;
  condition?: string;
}

export interface GraphMetadata {
  name: string;
  description?: string;
  version: string;
  createdAt: string;
  updatedAt: string;
}

export interface HealthStatus {
  status: 'healthy' | 'degraded' | 'unhealthy';
  version: string;
  uptime: number;
  features: string[];
}

export interface SchedulerStatus {
  queues: QueueStatus[];
  total_depth: number;
  backpressure: boolean;
}

export interface QueueStatus {
  name: string;
  priority: number;
  depth: number;
  workers: number;
}

export interface MemoryTierStatus {
  hot: HotMemoryStatus;
  warm: WarmMemoryStatus;
  cold: ColdMemoryStatus;
}

export interface HotMemoryStatus {
  enabled: boolean;
  entries: number;
  capacity: number;
  hit_rate: number;
}

export interface WarmMemoryStatus {
  enabled: boolean;
  connected: boolean;
  entries?: number;
}

export interface ColdMemoryStatus {
  enabled: boolean;
  connected: boolean;
}

export interface ExecuteGraphRequest {
  graph: GraphDefinition;
  initial_input: unknown;
  trace_id?: string;
}

export interface ExecuteGraphResponse {
  execution_id: string;
  status: 'completed' | 'failed' | 'partial';
  result: unknown;
  duration_ms: number;
  cost_usd: number;
  node_results: NodeResult[];
}

export interface NodeResult {
  node_id: string;
  status: 'success' | 'error' | 'skipped';
  output?: unknown;
  error?: string;
  duration_ms: number;
  cost_usd: number;
}

export interface ScheduleGraphRequest {
  graph: GraphDefinition;
  priority?: 'critical' | 'high' | 'normal' | 'low' | 'background';
  scheduled_at?: string; // ISO timestamp for delayed execution
}

export interface ScheduleGraphResponse {
  job_id: string;
  status: 'queued' | 'scheduled';
  estimated_start?: string;
  queue_position?: number;
}

export interface JobStatus {
  job_id: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  graph_id: string;
  priority: string;
  queue_position?: number;
  started_at?: string;
  completed_at?: string;
  progress?: number;
  result?: unknown;
  error?: string;
}

export interface JobListResponse {
  jobs: JobStatus[];
  total: number;
  running: number;
  queued: number;
}

export interface CancelJobResponse {
  job_id: string;
  status: 'cancelled' | 'not_found' | 'already_completed';
}

// ==================== API Functions ====================

/**
 * Get SAR runtime health status
 * GET /health
 */
export async function getHealth(): Promise<HealthStatus> {
  const response = await sarClient.get<HealthStatus>('/health');
  return response.data;
}

/**
 * Execute a graph synchronously
 * POST /execute/graph
 */
export async function executeGraph(
  graph: GraphDefinition,
  initialInput: unknown,
  traceId?: string
): Promise<ExecuteGraphResponse> {
  const body: ExecuteGraphRequest = {
    graph,
    initial_input: initialInput,
    trace_id: traceId,
  };
  const response = await sarClient.post<ExecuteGraphResponse>('/execute/graph', body);
  return response.data;
}

/**
 * Schedule a graph for async execution
 * POST /execute/graph/scheduled
 */
export async function scheduleGraph(
  graph: GraphDefinition,
  priority?: ScheduleGraphRequest['priority'],
  scheduledAt?: string
): Promise<ScheduleGraphResponse> {
  const body: ScheduleGraphRequest = {
    graph,
    priority,
    scheduled_at: scheduledAt,
  };
  const response = await sarClient.post<ScheduleGraphResponse>('/execute/graph/scheduled', body);
  return response.data;
}

/**
 * Get job execution status
 * GET /execute/graph/status/:jobId
 */
export async function getJobStatus(jobId: string): Promise<JobStatus> {
  const response = await sarClient.get<JobStatus>(`/execute/graph/status/${jobId}`);
  return response.data;
}

/**
 * Get scheduler status and queue depths
 * GET /scheduler
 */
export async function getSchedulerStatus(): Promise<SchedulerStatus> {
  const response = await sarClient.get<SchedulerStatus>('/scheduler');
  return response.data;
}

/**
 * Get memory tier status
 * GET /memory/status
 */
export async function getMemoryStatus(): Promise<MemoryTierStatus> {
  const response = await sarClient.get<MemoryTierStatus>('/memory/status');
  return response.data;
}

/**
 * List recent jobs
 * GET /jobs?limit=&offset=&status=
 */
export async function listJobs(
  limit = 20,
  offset = 0,
  status?: string
): Promise<JobListResponse> {
  const params = new URLSearchParams({
    limit: limit.toString(),
    offset: offset.toString(),
  });
  if (status) params.set('status', status);

  const response = await sarClient.get<JobListResponse>(`/jobs?${params}`);
  return response.data;
}

/**
 * Cancel a pending or running job
 * POST /jobs/:jobId/cancel
 */
export async function cancelJob(jobId: string): Promise<CancelJobResponse> {
  const response = await sarClient.post<CancelJobResponse>(`/jobs/${jobId}/cancel`);
  return response.data;
}

/**
 * Get recent executions for a graph
 * GET /graphs/:graphId/executions
 */
export async function getGraphExecutions(
  graphId: string,
  limit = 10
): Promise<ExecuteGraphResponse[]> {
  const params = new URLSearchParams({
    limit: limit.toString(),
  });
  const response = await sarClient.get<ExecuteGraphResponse[]>(
    `/graphs/${graphId}/executions?${params}`
  );
  return response.data;
}

// ==================== Error Helpers ====================

export function getSarRuntimeErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    if (!error.response) {
      return 'Cannot connect to SAR Runtime. Is port 8082 running?';
    }
    if (error.response.status === 503) {
      return 'SAR Runtime is starting up, please wait...';
    }
    return error.response.data?.error || `SAR Runtime error: ${error.response.status}`;
  }
  return 'Unexpected error connecting to SAR Runtime';
}

// Export the client for direct use if needed
export { sarClient };
