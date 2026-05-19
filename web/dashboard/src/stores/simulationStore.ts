/**
 * Simulation Store
 * Global state management for R-Sim runtime simulation components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// ============================================================================
// Types
// ============================================================================

export interface SimulationConfig {
  id: string
  name: string
  type: 'load' | 'stress' | 'chaos' | 'regression' | 'capacity'
  duration: number
  warmupDuration?: number
  cooldownDuration?: number
  parallelism?: number
  rampUpTime?: number
}

export interface SimulationMetrics {
  requestsTotal?: number
  requestsSuccess?: number
  requestsFailed?: number
  avgLatency?: number
  p50Latency?: number
  p95Latency?: number
  p99Latency?: number
  maxLatency?: number
  throughput?: number
  errorRate?: number
  timestamp?: number
}

export interface ForecastDataPoint {
  timestamp: number
  expectedLatency: number
  confidence: number
  upperBound: number
  lowerBound: number
}

export interface FailureNode {
  id: string
  name: string
  type: 'service' | 'endpoint' | 'database' | 'cache' | 'queue' | 'worker'
  failureProbability: number
  historicalRate?: number
  affectedRequests?: number
  correlationId?: string
}

export interface LatencyDataPoint {
  timestamp: number
  predicted: number
  actual?: number
  p50: number
  p95: number
  p99: number
}

export interface HallucinationRisk {
  id: string
  source: string
  type: 'context' | 'training' | 'reasoning' | 'retrieval'
  severity: 'critical' | 'high' | 'medium' | 'low'
  confidence: number
  description: string
  indicators?: string[]
  mitigationSuggestion?: string
}

export interface CostProjection {
  timestamp: number
  computeCost: number
  memoryCost: number
  networkCost: number
  storageCost: number
  totalCost: number
  cumulativeCost: number
}

export interface StressTestResult {
  id: string
  timestamp: number
  duration: number
  peakLoad: number
  steadyStateLoad?: number
  successRate: number
  avgResponseTime: number
  maxResponseTime: number
  errors: Array<{ type: string; count: number }>
  bottlenecks?: string[]
}

export interface ScalingProjection {
  timestamp: number
  currentReplicas: number
  predictedReplicas: number
  confidence: number
  trigger: 'cpu' | 'memory' | 'requests' | 'queue_depth' | 'custom'
  estimatedCostPerHour: number
}

export interface BehaviorPrediction {
  agentId: string
  agentName: string
  predictedActions: Array<{
    action: string
    probability: number
    expectedOutcome: string
    confidence: number
    timestamp: number
  }>
  riskScore?: number
  recommendedInterventions?: string[]
}

export interface WorkflowPath {
  id: string
  name: string
  probability: number
  steps: Array<{
    name: string
    duration: number
    successProbability: number
    alternativePath?: string
  }>
  totalDuration: number
  expectedCost: number
}

export interface ResourceCollision {
  id: string
  resourceA: string
  resourceB: string
  type: 'cpu' | 'memory' | 'io' | 'network' | 'disk'
  severity: 'critical' | 'high' | 'medium' | 'low'
  probability: number
  impact: string
  mitigation?: string
}

// ============================================================================
// State Interface
// ============================================================================

export interface SimulationState {
  // Simulation Control
  simulationConfig: SimulationConfig | null
  simulationStatus: 'idle' | 'preparing' | 'running' | 'paused' | 'completed' | 'failed'
  simulationMetrics: SimulationMetrics | null

  // Execution Forecast
  forecasts: ForecastDataPoint[]
  selectedForecastId: string | null

  // Failure Probability
  failureNodes: FailureNode[]
  selectedFailureNodeId: string | null
  failureThreshold: number

  // Latency Prediction
  latencyData: LatencyDataPoint[]
  selectedLatencyTimestamp: number | null

  // Hallucination Risk
  hallucinationRisks: HallucinationRisk[]
  selectedRiskId: string | null

  // Cost Simulation
  costProjections: CostProjection[]
  selectedCostPointId: string | null
  showCostBreakdown: boolean

  // Stress Test
  stressTestResults: StressTestResult[]
  currentTestId: string | null
  isStressTestRunning: boolean

  // Scaling Forecast
  scalingProjections: ScalingProjection[]
  selectedScalingTimestamp: number | null

  // Agent Behavior
  agentPredictions: BehaviorPrediction[]
  selectedAgentId: string | null

  // Workflow Outcome
  workflowPaths: WorkflowPath[]
  selectedPathId: string | null

  // Resource Collision
  resourceCollisions: ResourceCollision[]
  selectedCollisionId: string | null

  // UI State
  activePanel: string
  sidebarCollapsed: boolean
}

// ============================================================================
// Store
// ============================================================================

interface SimulationActions {
  // Simulation Control
  setSimulationConfig: (config: SimulationConfig | null) => void
  setSimulationStatus: (status: 'idle' | 'preparing' | 'running' | 'paused' | 'completed' | 'failed') => void
  setSimulationMetrics: (metrics: SimulationMetrics | null) => void
  startSimulation: () => void
  pauseSimulation: () => void
  stopSimulation: () => void
  resetSimulation: () => void

  // Execution Forecast
  setForecasts: (forecasts: ForecastDataPoint[]) => void
  selectForecast: (id: string | null) => void

  // Failure Probability
  setFailureNodes: (nodes: FailureNode[]) => void
  selectFailureNode: (id: string | null) => void
  setFailureThreshold: (threshold: number) => void

  // Latency Prediction
  setLatencyData: (data: LatencyDataPoint[]) => void
  selectLatencyTimestamp: (timestamp: number | null) => void

  // Hallucination Risk
  setHallucinationRisks: (risks: HallucinationRisk[]) => void
  selectHallucinationRisk: (id: string | null) => void
  dismissHallucinationRisk: (id: string) => void

  // Cost Simulation
  setCostProjections: (projections: CostProjection[]) => void
  selectCostPoint: (id: string | null) => void
  setShowCostBreakdown: (show: boolean) => void

  // Stress Test
  setStressTestResults: (results: StressTestResult[]) => void
  selectStressTest: (id: string | null) => void
  setIsStressTestRunning: (running: boolean) => void
  startStressTest: (config: SimulationConfig) => void
  stopStressTest: () => void

  // Scaling Forecast
  setScalingProjections: (projections: ScalingProjection[]) => void
  selectScalingProjection: (timestamp: number | null) => void

  // Agent Behavior
  setAgentPredictions: (predictions: BehaviorPrediction[]) => void
  selectAgent: (agentId: string | null) => void
  applyIntervention: (agentId: string, intervention: string) => void

  // Workflow Outcome
  setWorkflowPaths: (paths: WorkflowPath[]) => void
  selectWorkflowPath: (pathId: string | null) => void
  runWorkflowSimulation: () => void

  // Resource Collision
  setResourceCollisions: (collisions: ResourceCollision[]) => void
  selectResourceCollision: (id: string | null) => void
  resolveResourceCollision: (id: string) => void

  // UI Actions
  setActivePanel: (panel: string) => void
  toggleSidebar: () => void
}

export const useSimulationStore = create<SimulationState & SimulationActions>()(
  immer((set) => ({
    // ============================================================================
    // Initial State
    // ============================================================================

    simulationConfig: null,
    simulationStatus: 'idle',
    simulationMetrics: null,

    forecasts: [],
    selectedForecastId: null,

    failureNodes: [],
    selectedFailureNodeId: null,
    failureThreshold: 0.3,

    latencyData: [],
    selectedLatencyTimestamp: null,

    hallucinationRisks: [],
    selectedRiskId: null,

    costProjections: [],
    selectedCostPointId: null,
    showCostBreakdown: false,

    stressTestResults: [],
    currentTestId: null,
    isStressTestRunning: false,

    scalingProjections: [],
    selectedScalingTimestamp: null,

    agentPredictions: [],
    selectedAgentId: null,

    workflowPaths: [],
    selectedPathId: null,

    resourceCollisions: [],
    selectedCollisionId: null,

    activePanel: 'control',
    sidebarCollapsed: false,

    // ============================================================================
    // Simulation Control Actions
    // ============================================================================

    setSimulationConfig: (config) =>
      set((state) => {
        state.simulationConfig = config
      }),

    setSimulationStatus: (status) =>
      set((state) => {
        state.simulationStatus = status
      }),

    setSimulationMetrics: (metrics) =>
      set((state) => {
        state.simulationMetrics = metrics
      }),

    startSimulation: () =>
      set((state) => {
        state.simulationStatus = 'running'
      }),

    pauseSimulation: () =>
      set((state) => {
        state.simulationStatus = 'paused'
      }),

    stopSimulation: () =>
      set((state) => {
        state.simulationStatus = 'idle'
      }),

    resetSimulation: () =>
      set((state) => {
        state.simulationStatus = 'idle'
        state.simulationMetrics = null
      }),

    // ============================================================================
    // Execution Forecast Actions
    // ============================================================================

    setForecasts: (forecasts) =>
      set((state) => {
        state.forecasts = forecasts
      }),

    selectForecast: (id) =>
      set((state) => {
        state.selectedForecastId = id
      }),

    // ============================================================================
    // Failure Probability Actions
    // ============================================================================

    setFailureNodes: (nodes) =>
      set((state) => {
        state.failureNodes = nodes
      }),

    selectFailureNode: (id) =>
      set((state) => {
        state.selectedFailureNodeId = id
      }),

    setFailureThreshold: (threshold) =>
      set((state) => {
        state.failureThreshold = threshold
      }),

    // ============================================================================
    // Latency Prediction Actions
    // ============================================================================

    setLatencyData: (data) =>
      set((state) => {
        state.latencyData = data
      }),

    selectLatencyTimestamp: (timestamp) =>
      set((state) => {
        state.selectedLatencyTimestamp = timestamp
      }),

    // ============================================================================
    // Hallucination Risk Actions
    // ============================================================================

    setHallucinationRisks: (risks) =>
      set((state) => {
        state.hallucinationRisks = risks
      }),

    selectHallucinationRisk: (id) =>
      set((state) => {
        state.selectedRiskId = id
      }),

    dismissHallucinationRisk: (id) =>
      set((state) => {
        state.hallucinationRisks = state.hallucinationRisks.filter((r) => r.id !== id)
        if (state.selectedRiskId === id) {
          state.selectedRiskId = null
        }
      }),

    // ============================================================================
    // Cost Simulation Actions
    // ============================================================================

    setCostProjections: (projections) =>
      set((state) => {
        state.costProjections = projections
      }),

    selectCostPoint: (id) =>
      set((state) => {
        state.selectedCostPointId = id
      }),

    setShowCostBreakdown: (show) =>
      set((state) => {
        state.showCostBreakdown = show
      }),

    // ============================================================================
    // Stress Test Actions
    // ============================================================================

    setStressTestResults: (results) =>
      set((state) => {
        state.stressTestResults = results
      }),

    selectStressTest: (id) =>
      set((state) => {
        state.currentTestId = id
      }),

    setIsStressTestRunning: (running) =>
      set((state) => {
        state.isStressTestRunning = running
      }),

    startStressTest: (config) =>
      set((state) => {
        state.isStressTestRunning = true
        state.simulationConfig = config
      }),

    stopStressTest: () =>
      set((state) => {
        state.isStressTestRunning = false
      }),

    // ============================================================================
    // Scaling Forecast Actions
    // ============================================================================

    setScalingProjections: (projections) =>
      set((state) => {
        state.scalingProjections = projections
      }),

    selectScalingProjection: (timestamp) =>
      set((state) => {
        state.selectedScalingTimestamp = timestamp
      }),

    // ============================================================================
    // Agent Behavior Actions
    // ============================================================================

    setAgentPredictions: (predictions) =>
      set((state) => {
        state.agentPredictions = predictions
      }),

    selectAgent: (agentId) =>
      set((state) => {
        state.selectedAgentId = agentId
      }),

    applyIntervention: (agentId, intervention) =>
      set((state) => {
        console.log('Applying intervention:', agentId, intervention)
      }),

    // ============================================================================
    // Workflow Outcome Actions
    // ============================================================================

    setWorkflowPaths: (paths) =>
      set((state) => {
        state.workflowPaths = paths
      }),

    selectWorkflowPath: (pathId) =>
      set((state) => {
        state.selectedPathId = pathId
      }),

    runWorkflowSimulation: () =>
      set((state) => {
        console.log('Running workflow simulation')
      }),

    // ============================================================================
    // Resource Collision Actions
    // ============================================================================

    setResourceCollisions: (collisions) =>
      set((state) => {
        state.resourceCollisions = collisions
      }),

    selectResourceCollision: (id) =>
      set((state) => {
        state.selectedCollisionId = id
      }),

    resolveResourceCollision: (id) =>
      set((state) => {
        state.resourceCollisions = state.resourceCollisions.filter((c) => c.id !== id)
        if (state.selectedCollisionId === id) {
          state.selectedCollisionId = null
        }
      }),

    // ============================================================================
    // UI Actions
    // ============================================================================

    setActivePanel: (panel) =>
      set((state) => {
        state.activePanel = panel
      }),

    toggleSidebar: () =>
      set((state) => {
        state.sidebarCollapsed = !state.sidebarCollapsed
      }),
  }))
)

// ============================================================================
// Selectors
// ============================================================================

export const useSimulationControl = () =>
  useSimulationStore((state) => ({
    config: state.simulationConfig,
    status: state.simulationStatus,
    metrics: state.simulationMetrics,
  }))

export const useExecutionForecast = () =>
  useSimulationStore((state) => ({
    forecasts: state.forecasts,
    selectedId: state.selectedForecastId,
  }))

export const useFailureProbability = () =>
  useSimulationStore((state) => ({
    nodes: state.failureNodes,
    selectedNodeId: state.selectedFailureNodeId,
    threshold: state.failureThreshold,
  }))

export const useLatencyPrediction = () =>
  useSimulationStore((state) => ({
    data: state.latencyData,
    selectedTimestamp: state.selectedLatencyTimestamp,
  }))

export const useHallucinationRisk = () =>
  useSimulationStore((state) => ({
    risks: state.hallucinationRisks,
    selectedId: state.selectedRiskId,
  }))

export const useCostSimulation = () =>
  useSimulationStore((state) => ({
    projections: state.costProjections,
    selectedPointId: state.selectedCostPointId,
    showBreakdown: state.showCostBreakdown,
  }))

export const useStressTest = () =>
  useSimulationStore((state) => ({
    results: state.stressTestResults,
    currentTestId: state.currentTestId,
    isRunning: state.isStressTestRunning,
  }))

export const useScalingForecast = () =>
  useSimulationStore((state) => ({
    projections: state.scalingProjections,
    selectedTimestamp: state.selectedScalingTimestamp,
  }))

export const useAgentBehavior = () =>
  useSimulationStore((state) => ({
    predictions: state.agentPredictions,
    selectedAgentId: state.selectedAgentId,
  }))

export const useWorkflowOutcome = () =>
  useSimulationStore((state) => ({
    paths: state.workflowPaths,
    selectedPathId: state.selectedPathId,
  }))

export const useResourceCollision = () =>
  useSimulationStore((state) => ({
    collisions: state.resourceCollisions,
    selectedId: state.selectedCollisionId,
  }))

export const useSimulationUI = () =>
  useSimulationStore((state) => ({
    activePanel: state.activePanel,
    sidebarCollapsed: state.sidebarCollapsed,
  }))
