/**
 * R-Sim Simulation API Service
 * Connects StudioPage simulation components to the backend simulation engine
 * Phase 3.2 Backend — Wire frontend to Go simulation engine
 */

import { useState, useCallback } from 'react';

// Types matching the Go backend structs
export interface NodeSpec {
  id: string;
  type: string;
  timeout: number;
  cost_usd: number;
  metadata?: Record<string, unknown>;
}

export interface EdgeSpec {
  from: string;
  to: string;
  probability_success: number;
}

export interface WorkflowSpec {
  nodes: NodeSpec[];
  edges: EdgeSpec[];
}

export interface SimulationResult {
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

export interface MonteCarloResult {
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
  outcomes: OutcomeSample[];
  bottleneck_nodes: string[];
  cost_breakdown: Record<string, number>;
}

export interface OutcomeSample {
  outcome: 'success' | 'partial' | 'failed';
  probability: number;
  latency_ms: number;
  cost_usd: number;
  failed_nodes: string[];
  risk_factors: string[];
}

export interface ExecutionForecast {
  workflow_id: string;
  time_horizon: string;
  predicted_executions: number;
  success_rate: number;
  avg_latency_ms: number;
  cost_usd: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  predictions: PredictionPoint[];
}

export interface PredictionPoint {
  timestamp: string;
  executions: number;
  success_rate: number;
}

export interface CostForecast {
  workflow_id: string;
  total_cost_usd: number;
  per_call_usd: number;
  lower_bound_usd: number;
  upper_bound_usd: number;
  confidence: number;
  by_node: Record<string, number>;
}

export interface LatencyForecast {
  workflow_id: string;
  load_level: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  avg_latency_ms: number;
}

export interface StressTestConfig {
  iterations: number;
  parallelism: number;
  workflow_id: string;
  load_profile: string;
}

export interface StressTestResult {
  id: string;
  status: string;
  iterations: number;
  total_executions: number;
  success_rate: number;
  throughput: number;
  latency_p50: number;
  latency_p95: number;
  latency_p99: number;
  errors: ErrorBreakdown[];
}

export interface ErrorBreakdown {
  type: string;
  count: number;
  percentage: number;
}

export interface ResourceCollisionSpec {
  resources: ResourceSpec[];
}

export interface ResourceSpec {
  id: string;
  name: string;
  type: string;
  capacity: number;
  tasks: TaskSchedule[];
}

export interface TaskSchedule {
  task_id: string;
  start_ms: number;
  end_ms: number;
  usage: number;
  priority: number;
}

export interface CollisionResult {
  collisions: Collision[];
  resolutions: string[];
}

export interface Collision {
  resource_id: string;
  resource_name: string;
  severity: 'high' | 'medium' | 'low';
  conflicting_tasks: TaskSchedule[];
  resolution: string;
}

export interface AgentBehaviorSpec {
  agent_id: string;
  history_size: number;
  context: string;
}

export interface AgentBehaviorPrediction {
  agent_id: string;
  confidence: number;
  based_on_samples: number;
  likely_actions: ActionPrediction[];
  current_task_prediction: string;
}

export interface ActionPrediction {
  action: string;
  probability: number;
  expected_outcome: string;
  risk_level: 'low' | 'medium' | 'high';
}

export interface HallucinationRiskSpec {
  model_id: string;
  prompt_length: number;
  context_window: number;
  task_type: 'code_gen' | 'reasoning' | 'summarization' | 'factual';
  complexity: 'low' | 'medium' | 'high' | 'extreme';
  previous_errors: number;
}

export interface HallucinationRiskResult {
  model_id: string;
  risk_score: number;
  risk_level: 'low' | 'medium' | 'high' | 'critical';
  contributing_factors: string[];
  recommendations: string[];
  confidence: number;
}

// API client
const API_BASE = '/api';

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    const msg = (body as any)?.error?.message || `HTTP ${res.status}`;
    throw new Error(msg);
  }

  return res.json() as Promise<T>;
}

// Hook for simulation data
export function useSimulation() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Simulate workflow (dry-run)
  const simulateWorkflow = useCallback(async (workflow: WorkflowSpec, iterations = 100): Promise<SimulationResult> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/simulate/workflow', {
        method: 'POST',
        body: JSON.stringify({ workflow, iterations }),
      });
      return res.result;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Monte Carlo simulation
  const runMonteCarlo = useCallback(async (workflow: WorkflowSpec, iterations = 1000): Promise<MonteCarloResult> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/simulate/monte-carlo', {
        method: 'POST',
        body: JSON.stringify({ workflow, iterations }),
      });
      return res.result;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Execution forecast
  const getExecutionForecast = useCallback(async (workflowId: string, timeHorizon: string, callVolume: number, nodes: NodeSpec[]): Promise<ExecutionForecast> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/forecast/execution', {
        method: 'POST',
        body: JSON.stringify({ workflow_id: workflowId, time_horizon: timeHorizon, call_volume: callVolume, nodes }),
      });
      return res.forecast;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Cost forecast
  const getCostForecast = useCallback(async (workflowId: string, nodes: NodeSpec[], callVolume: number): Promise<CostForecast> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/forecast/cost', {
        method: 'POST',
        body: JSON.stringify({ workflow_id: workflowId, nodes, call_volume: callVolume }),
      });
      return res.forecast;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Latency forecast
  const getLatencyForecast = useCallback(async (workflowId: string, nodes: NodeSpec[], loadLevel: number): Promise<LatencyForecast> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/forecast/latency', {
        method: 'POST',
        body: JSON.stringify({ workflow_id: workflowId, nodes, load_level: loadLevel }),
      });
      return res.forecast;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Stress test
  const startStressTest = useCallback(async (config: StressTestConfig): Promise<string> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/stress-test/start', {
        method: 'POST',
        body: JSON.stringify(config),
      });
      return res.id;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  const getStressTest = useCallback(async (id: string): Promise<StressTestResult> => {
    const res = await apiFetch<any>(`/v1/stress-test/${id}`);
    return res.result;
  }, []);

  const abortStressTest = useCallback(async (id: string): Promise<void> => {
    await apiFetch(`/v1/stress-test/${id}/abort`, { method: 'POST' });
  }, []);

  // Resource collision detection
  const detectCollisions = useCallback(async (spec: ResourceCollisionSpec): Promise<CollisionResult> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/detect/collisions', {
        method: 'POST',
        body: JSON.stringify(spec),
      });
      return res.result;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Agent behavior prediction
  const predictAgentBehavior = useCallback(async (spec: AgentBehaviorSpec): Promise<AgentBehaviorPrediction> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/predict/agent-behavior', {
        method: 'POST',
        body: JSON.stringify(spec),
      });
      return res.prediction;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Hallucination risk analysis
  const analyzeHallucinationRisk = useCallback(async (spec: HallucinationRiskSpec): Promise<HallucinationRiskResult> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/analyze/hallucination-risk', {
        method: 'POST',
        body: JSON.stringify(spec),
      });
      return res.result;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    loading,
    error,
    simulateWorkflow,
    runMonteCarlo,
    getExecutionForecast,
    getCostForecast,
    getLatencyForecast,
    startStressTest,
    getStressTest,
    abortStressTest,
    detectCollisions,
    predictAgentBehavior,
    analyzeHallucinationRisk,
  };
}