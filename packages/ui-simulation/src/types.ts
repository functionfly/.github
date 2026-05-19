/**
 * @functionfly/ui-simulation
 * Runtime Simulation Components - R-Sim system for predictive analysis
 */

// ============================================================================
// Simulation Control Center
// ============================================================================

export interface SimulationConfig {
  id: string;
  name: string;
  type: 'load' | 'stress' | 'chaos' | 'regression' | 'capacity';
  duration: number;
  warmupDuration?: number;
  cooldownDuration?: number;
  parallelism?: number;
  rampUpTime?: number;
}

export interface SimulationMetrics {
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

export interface SimulationControlCenterProps {
  config: SimulationConfig | null;
  status: 'idle' | 'preparing' | 'running' | 'paused' | 'completed' | 'failed';
  metrics: SimulationMetrics | null;
  onStart?: () => void;
  onPause?: () => void;
  onStop?: () => void;
  onReset?: () => void;
  onConfigChange?: (config: SimulationConfig) => void;
  className?: string;
}

// ============================================================================
// Execution Forecast Panel
// ============================================================================

export interface ForecastDataPoint {
  timestamp: number;
  expectedLatency: number;
  confidence: number;
  upperBound: number;
  lowerBound: number;
}

export interface ExecutionForecastPanelProps {
  forecasts: ForecastDataPoint[];
  horizon?: number;
  selectedPointId?: string | null;
  onPointSelect?: (point: ForecastDataPoint) => void;
  onRefresh?: () => void;
  className?: string;
}

// ============================================================================
// Failure Probability Map
// ============================================================================

export interface FailureNode {
  id: string;
  name: string;
  type: 'service' | 'endpoint' | 'database' | 'cache' | 'queue' | 'worker';
  failureProbability: number;
  historicalRate?: number;
  affectedRequests?: number;
  correlationId?: string;
}

export interface FailureProbabilityMapProps {
  nodes: FailureNode[];
  selectedNodeId?: string | null;
  threshold?: number;
  onNodeSelect?: (node: FailureNode) => void;
  onNodeHover?: (node: FailureNode | null) => void;
  className?: string;
}

// ============================================================================
// Latency Prediction Graph
// ============================================================================

export interface LatencyDataPoint {
  timestamp: number;
  predicted: number;
  actual?: number;
  p50: number;
  p95: number;
  p99: number;
}

export interface LatencyPredictionGraphProps {
  data: LatencyDataPoint[];
  selectedTimestamp?: number | null;
  showActual?: boolean;
  showConfidence?: boolean;
  onDataPointSelect?: (point: LatencyDataPoint) => void;
  className?: string;
}

// ============================================================================
// Hallucination Risk Analyzer
// ============================================================================

export interface HallucinationRisk {
  id: string;
  source: string;
  type: 'context' | 'training' | 'reasoning' | 'retrieval';
  severity: 'critical' | 'high' | 'medium' | 'low';
  confidence: number;
  description: string;
  indicators?: string[];
  mitigationSuggestion?: string;
}

export interface HallucinationRiskAnalyzerProps {
  risks: HallucinationRisk[];
  selectedRiskId?: string | null;
  threshold?: number;
  onRiskSelect?: (risk: HallucinationRisk) => void;
  onRiskDismiss?: (riskId: string) => void;
  className?: string;
}

// ============================================================================
// Cost Simulation Chart
// ============================================================================

export interface CostProjection {
  timestamp: number;
  computeCost: number;
  memoryCost: number;
  networkCost: number;
  storageCost: number;
  totalCost: number;
  cumulativeCost: number;
}

export interface CostSimulationChartProps {
  projections: CostProjection[];
  selectedPointId?: string | null;
  showBreakdown?: boolean;
  comparisonBaseline?: CostProjection[];
  onPointSelect?: (point: CostProjection) => void;
  className?: string;
}

// ============================================================================
// Stress Test Runner
// ============================================================================

export interface StressTestResult {
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

export interface StressTestRunnerProps {
  results: StressTestResult[];
  currentTestId?: string | null;
  isRunning?: boolean;
  onTestStart?: (config: SimulationConfig) => void;
  onTestStop?: () => void;
  onTestSelect?: (result: StressTestResult) => void;
  className?: string;
}

// ============================================================================
// Scaling Forecast Map
// ============================================================================

export interface ScalingProjection {
  timestamp: number;
  currentReplicas: number;
  predictedReplicas: number;
  confidence: number;
  trigger: 'cpu' | 'memory' | 'requests' | 'queue_depth' | 'custom';
  estimatedCostPerHour: number;
}

export interface ScalingForecastMapProps {
  projections: ScalingProjection[];
  selectedTimestamp?: number | null;
  onProjectionSelect?: (projection: ScalingProjection) => void;
  onProjectionHover?: (projection: ScalingProjection | null) => void;
  className?: string;
}

// ============================================================================
// Agent Behavior Predictor
// ============================================================================

export interface BehaviorPrediction {
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

export interface AgentBehaviorPredictorProps {
  predictions: BehaviorPrediction[];
  selectedAgentId?: string | null;
  timeHorizon?: number;
  onAgentSelect?: (prediction: BehaviorPrediction) => void;
  onInterventionApply?: (agentId: string, intervention: string) => void;
  className?: string;
}

// ============================================================================
// Workflow Outcome Simulator
// ============================================================================

export interface WorkflowPath {
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

export interface WorkflowOutcomeSimulatorProps {
  workflowId: string;
  paths: WorkflowPath[];
  selectedPathId?: string | null;
  onPathSelect?: (path: WorkflowPath) => void;
  onSimulationRun?: () => void;
  className?: string;
}

// ============================================================================
// Resource Collision Detector
// ============================================================================

export interface ResourceCollision {
  id: string;
  resourceA: string;
  resourceB: string;
  type: 'cpu' | 'memory' | 'io' | 'network' | 'disk';
  severity: 'critical' | 'high' | 'medium' | 'low';
  probability: number;
  impact: string;
  mitigation?: string;
}

export interface ResourceCollisionDetectorProps {
  collisions: ResourceCollision[];
  selectedCollisionId?: string | null;
  threshold?: number;
  onCollisionSelect?: (collision: ResourceCollision) => void;
  onCollisionResolve?: (collisionId: string) => void;
  className?: string;
}
