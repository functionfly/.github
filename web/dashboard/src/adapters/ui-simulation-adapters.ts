/**
 * UI Package Adapters - Simulation
 * Transform dashboard store types to match UI package component expectations
 */

import type { SimulationConfig as StudioSimulationConfig, SimulationResult as StudioSimulationResult } from '../hooks/useStudio';

// Target types for UI package compatibility (derived from ui-simulation types)
// Note: These types are defined locally because the UI package doesn't export them as standalone types

export interface UISimulationConfig {
  id: string;
  name: string;
  type: 'load' | 'stress' | 'chaos' | 'regression' | 'capacity';
  duration: number;
  warmupDuration?: number;
  cooldownDuration?: number;
  parallelism?: number;
  rampUpTime?: number;
}

export interface UISimulationMetrics {
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

export type UISimulationStatus = 'idle' | 'preparing' | 'running' | 'paused' | 'completed' | 'failed';

export interface UIForecastDataPoint {
  timestamp: number;
  expectedLatency: number;
  confidence: number;
  upperBound: number;
  lowerBound: number;
}

export interface UIFailureNode {
  id: string;
  name: string;
  type: 'service' | 'endpoint' | 'database' | 'cache' | 'queue' | 'worker';
  failureProbability: number;
  historicalRate?: number;
  affectedRequests?: number;
  correlationId?: string;
}

export interface UILatencyDataPoint {
  timestamp: number;
  predicted: number;
  actual?: number;
  p50: number;
  p95: number;
  p99: number;
}

export interface UICostProjection {
  timestamp: number;
  computeCost: number;
  memoryCost: number;
  networkCost: number;
  storageCost: number;
  totalCost: number;
  cumulativeCost: number;
}

export interface UIScalingProjection {
  timestamp: number;
  currentReplicas: number;
  predictedReplicas: number;
  confidence: number;
  trigger: 'cpu' | 'memory' | 'requests' | 'queue_depth' | 'custom';
  estimatedCostPerHour: number;
}

export interface UIBehaviorPrediction {
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

export interface UIStressTestResult {
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

export interface UIHallucinationRisk {
  id: string;
  source: string;
  type: 'context' | 'training' | 'reasoning' | 'retrieval';
  severity: 'critical' | 'high' | 'medium' | 'low';
  confidence: number;
  description: string;
  indicators?: string[];
  mitigationSuggestion?: string;
}

export function adaptSimulationConfig(config: StudioSimulationConfig | null): UISimulationConfig | null {
  if (!config) return null;
  return {
    id: config.id || crypto.randomUUID(),
    name: config.name,
    type: mapStressLevelToSimulationType(config.stressLevel),
    duration: config.duration || 60,
    warmupDuration: config.latencyMs,
    parallelism: config.iterations,
  };
}

function mapStressLevelToSimulationType(stressLevel?: 'low' | 'medium' | 'high' | 'extreme'): 'load' | 'stress' | 'chaos' | 'regression' | 'capacity' {
  switch (stressLevel) {
    case 'low':
      return 'load';
    case 'medium':
      return 'regression';
    case 'high':
      return 'capacity';
    case 'extreme':
      return 'stress';
    default:
      return 'regression';
  }
}

export function adaptSimulationStatus(result: StudioSimulationResult | null | undefined): UISimulationStatus {
  if (!result) return 'idle';
  switch (result.status) {
    case 'pending':
      return 'preparing';
    case 'running':
      return 'running';
    case 'completed':
      return 'completed';
    case 'failed':
      return 'failed';
    case 'aborted':
      return 'failed';
    default:
      return 'idle';
  }
}

export function adaptSimulationMetrics(result: StudioSimulationResult | null | undefined): UISimulationMetrics | null {
  if (!result || !result.metrics) return null;
  return {
    requestsTotal: result.metrics.totalExecutions,
    requestsSuccess: result.metrics.successfulExecutions,
    requestsFailed: result.metrics.failedExecutions,
    avgLatency: result.metrics.averageLatencyMs,
    p50Latency: result.metrics.p50LatencyMs,
    p95Latency: result.metrics.p95LatencyMs,
    p99Latency: result.metrics.p99LatencyMs,
    throughput: result.metrics.throughput,
    errorRate: result.metrics.totalExecutions > 0
      ? result.metrics.failedExecutions / result.metrics.totalExecutions
      : 0,
  };
}

export function adaptStudioConfigFromUI(uiConfig: UISimulationConfig): StudioSimulationConfig {
  return {
    id: uiConfig.id,
    name: uiConfig.name,
    iterations: uiConfig.parallelism || 10,
    duration: uiConfig.duration,
    stressLevel: mapSimulationTypeToStressLevel(uiConfig.type),
  };
}

function mapSimulationTypeToStressLevel(type: 'load' | 'stress' | 'chaos' | 'regression' | 'capacity'): 'low' | 'medium' | 'high' | 'extreme' {
  switch (type) {
    case 'load':
      return 'low';
    case 'regression':
      return 'medium';
    case 'capacity':
      return 'high';
    case 'stress':
      return 'extreme';
    default:
      return 'medium';
  }
}

export interface ForecastDataInput {
  timestamp: number;
  value: number;
  lower: number;
  upper: number;
  confidence: number;
}

export function adaptForecastData(data: ForecastDataInput[]): UIForecastDataPoint[] {
  return data.map(d => ({
    timestamp: d.timestamp,
    expectedLatency: d.value,
    confidence: d.confidence,
    upperBound: d.upper,
    lowerBound: d.lower,
  }));
}

export interface FailureNodeInput {
  nodeId: string;
  nodeName: string;
  probability: number;
  factors: Array<{ name: string; contribution: number }>;
  trend: string;
}

export function adaptFailureNodes(data: FailureNodeInput[]): UIFailureNode[] {
  return data.map((n, i) => ({
    id: n.nodeId,
    name: n.nodeName,
    type: 'service' as const,
    failureProbability: n.probability,
    historicalRate: n.factors[0]?.contribution,
    affectedRequests: Math.floor(n.probability * 1000),
  }));
}

export interface LatencyDataPointInput {
  timestamp: number;
  p50: number;
  p95: number;
  p99: number;
  samples?: number;
}

export function adaptLatencyData(data: LatencyDataPointInput[]): UILatencyDataPoint[] {
  return data.map(d => ({
    timestamp: d.timestamp,
    predicted: d.p50,
    actual: d.p95,
    p50: d.p50,
    p95: d.p95,
    p99: d.p99,
  }));
}

export interface CostEstimateInput {
  totalTokens: number;
  inputTokens: number;
  outputTokens: number;
  computeCost: number;
  apiCost: number;
  totalCost: number;
  hourlyBreakdown: Array<{ timestamp: number; cost: number }>;
}

export function adaptCostProjections(estimate: CostEstimateInput): UICostProjection[] {
  const projections: UICostProjection[] = [];
  const now = Date.now();
  for (let i = 0; i < 24; i++) {
    const hourCost = estimate.totalCost / 24;
    projections.push({
      timestamp: now + i * 3600000,
      computeCost: hourCost * 0.4,
      memoryCost: hourCost * 0.2,
      networkCost: hourCost * 0.15,
      storageCost: hourCost * 0.1,
      totalCost: hourCost,
      cumulativeCost: hourCost * (i + 1),
    });
  }
  return projections;
}

export interface ScalingProjectionInput {
  component: string;
  currentReplicas: number;
  predictedReplicas: number;
  timestamp: number;
  trigger: 'requests' | 'memory' | 'cpu' | string;
}

export function adaptScalingProjections(data: ScalingProjectionInput[]): UIScalingProjection[] {
  return data.map(d => ({
    timestamp: d.timestamp,
    currentReplicas: d.currentReplicas,
    predictedReplicas: d.predictedReplicas,
    confidence: 0.85,
    trigger: (d.trigger === 'requests' ? 'requests' : d.trigger === 'memory' ? 'memory' : 'cpu') as UIScalingProjection['trigger'],
    estimatedCostPerHour: d.predictedReplicas * 0.05,
  }));
}

export interface AgentPredictionInput {
  agentId: string;
  likelyActions: Array<{ action: string; probability: number; expectedOutcome: string }>;
  confidence: number;
  basedOnSamples: number;
}

export function adaptBehaviorPredictions(data: AgentPredictionInput[], agentNames: Record<string, string>): UIBehaviorPrediction[] {
  return data.map(d => ({
    agentId: d.agentId,
    agentName: agentNames[d.agentId] || `Agent ${d.agentId.slice(0, 8)}`,
    predictedActions: d.likelyActions.map(a => ({
      action: a.action,
      probability: a.probability,
      expectedOutcome: a.expectedOutcome,
      confidence: d.confidence,
      timestamp: Date.now(),
    })),
    riskScore: 1 - d.confidence,
    recommendedInterventions: d.confidence < 0.8 ? ['Review agent output', 'Add validation'] : undefined,
  }));
}

export function adaptStressTestResults(results: UIStressTestResult[]): UIStressTestResult[] {
  return results;
}

export function createDefaultStressTestConfig(config: StudioSimulationConfig): UISimulationConfig {
  return {
    id: crypto.randomUUID(),
    name: `${config.name || 'Stress Test'}`,
    type: 'stress',
    duration: config.duration || 60,
    parallelism: config.iterations || 100,
  };
}
