/**
 * Observability Store
 * Global state management for live telemetry and monitoring components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// ============================================================================
// Types
// ============================================================================

export interface TelemetryMetric {
  id: string
  label: string
  value: number
  unit: string
  timestamp: number
  trend?: "up" | "down" | "stable"
  delta?: number
  status?: "ok" | "warning" | "error"
  min?: number
  max?: number
  avg?: number
  p50?: number
  p95?: number
  p99?: number
}

export interface TokenUsage {
  inputTokens: number
  outputTokens: number
  totalTokens: number
  costUSD: number
  models: Record<string, { calls: number; tokens: number; cost: number }>
  timeRange: { start: string; end: string }
}

export interface InferenceSpan {
  id: string
  parentId?: string
  name: string
  startTime: number
  endTime?: number
  duration?: number
  status: "running" | "completed" | "failed"
  model?: string
  inputTokens?: number
  outputTokens?: number
  error?: string
  children?: InferenceSpan[]
}

export interface ErrorNode {
  id: string
  name: string
  type: "function" | "agent" | "runtime" | "external"
  error?: string
  children?: ErrorNode[]
}

export interface TraceSpan {
  id: string
  traceId: string
  parentId?: string
  serviceName: string
  operationName: string
  startTime: number
  duration: number
  status: "ok" | "error"
  tags?: Record<string, string>
}

export interface ReplayRequest {
  id: string
  timestamp: string
  method: string
  path: string
  status: number
  duration: number
  requestSize: number
  responseSize: number
}

export interface LogEntry {
  id: string
  timestamp: number
  level: "debug" | "info" | "warn" | "error"
  message: string
  service?: string
  fields?: Record<string, string>
}

export interface Anomaly {
  id: string
  metric: string
  description: string
  detectedAt: number
  severity: "critical" | "high" | "medium" | "low"
  value: number
  expectedRange: [number, number]
  confidence: number
}

export interface Incident {
  id: string
  severity: "critical" | "high" | "medium" | "low"
  title: string
  description?: string
  startTime: number
  endTime?: number
  status: "active" | "resolved"
  affectedServices?: string[]
}

export interface CostProjection {
  period: string
  cost: number
  confidence: number
}

export interface ResourceForecastPoint {
  timestamp: number
  actual?: number
  predicted?: number
  lower?: number
  upper?: number
}

// ============================================================================
// Store Interface
// ============================================================================

interface ObservabilityState {
  // Time range
  timeRange: "1h" | "6h" | "24h" | "7d" | "30d"
  setTimeRange: (range: "1h" | "6h" | "24h" | "7d" | "30d") => void

  // Metrics
  metrics: TelemetryMetric[]
  tokenUsage: TokenUsage | null
  setMetrics: (metrics: TelemetryMetric[]) => void
  setTokenUsage: (usage: TokenUsage) => void
  updateMetric: (id: string, updates: Partial<TelemetryMetric>) => void

  // Latency data
  latencyData: Array<{ timestamp: number; p50: number; p95: number; p99: number }>
  setLatencyData: (data: Array<{ timestamp: number; p50: number; p95: number; p99: number }>) => void

  // Cost heatmap
  costHeatmapData: Array<{ hour: number; day: string; cost: number; requests: number }>
  setCostHeatmapData: (data: Array<{ hour: number; day: string; cost: number; requests: number }>) => void

  // Execution profiler
  executions: Array<{
    id: string
    name: string
    duration: number
    tokens: number
    cost: number
    timestamp: string
    status: "success" | "error" | "timeout"
    retries: number
  }>
  setExecutions: (executions: ObservabilityState['executions']) => void

  // Inference traces
  currentTraceId: string | null
  spans: InferenceSpan[]
  selectedSpanId: string | null
  setCurrentTrace: (traceId: string, spans: InferenceSpan[]) => void
  setSelectedSpanId: (id: string | null) => void

  // Error cascade
  rootError: ErrorNode | null
  setRootError: (error: ErrorNode) => void

  // Memory pressure
  memoryData: Array<{
    runtime: string
    usedMB: number
    limitMB: number
    pressure: "low" | "medium" | "high" | "critical"
  }>
  setMemoryData: (data: ObservabilityState['memoryData']) => void

  // GPU data
  gpuData: Array<{
    gpuId: string
    name: string
    utilizationPct: number
    memoryUsedMB: number
    memoryTotalMB: number
    temperatureC?: number
    powerW?: number
  }>
  setGpuData: (data: ObservabilityState['gpuData']) => void

  // Bandwidth data
  bandwidthData: Array<{
    interface: string
    rxMB: number
    txMB: number
    rxPackets: number
    txPackets: number
  }>
  setBandwidthData: (data: ObservabilityState['bandwidthData']) => void

  // Distributed traces
  distributedTraceId: string | null
  distributedSpans: TraceSpan[]
  setDistributedTrace: (traceId: string, spans: TraceSpan[]) => void

  // Request replay
  replayRequests: ReplayRequest[]
  setReplayRequests: (requests: ReplayRequest[]) => void

  // Live logs
  logs: LogEntry[]
  selectedLogId: string | null
  addLog: (log: LogEntry) => void
  clearLogs: () => void
  setSelectedLogId: (id: string | null) => void

  // Runtime metrics
  runtimeMetrics: Array<{
    id: string
    name: string
    status: "healthy" | "degraded" | "down"
    metrics: {
      requests: number
      latencyP50: number
      latencyP99: number
      errorRate: number
      cpu: number
      memory: number
    }
  }>
  setRuntimeMetrics: (runtimes: ObservabilityState['runtimeMetrics']) => void

  // Radar metrics
  radarMetrics: Array<{ label: string; value: number; max: number }>
  setRadarMetrics: (metrics: Array<{ label: string; value: number; max: number }>) => void

  // Incidents
  incidents: Incident[]
  selectedIncidentId: string | null
  setIncidents: (incidents: Incident[]) => void
  setSelectedIncidentId: (id: string | null) => void

  // Anomalies
  anomalies: Anomaly[]
  selectedAnomalyId: string | null
  setAnomalies: (anomalies: Anomaly[]) => void
  setSelectedAnomalyId: (id: string | null) => void

  // Cost prediction
  currentCost: number
  predictedCost: number
  costTrend: "increasing" | "stable" | "decreasing"
  costBasis: string
  costProjections: CostProjection[]
  setCostPrediction: (data: {
    current: number
    predicted: number
    trend: "increasing" | "stable" | "decreasing"
    basis: string
    projections: CostProjection[]
  }) => void

  // Resource forecast
  resourceForecastData: ResourceForecastPoint[]
  setResourceForecastData: (data: ResourceForecastPoint[]) => void

  // Show/hide panels
  visiblePanels: Set<string>
  togglePanel: (panel: string) => void

  // Reset
  reset: () => void
}

// ============================================================================
// Initial State
// ============================================================================

const initialState = {
  timeRange: "24h" as const,
  metrics: [],
  tokenUsage: null,
  latencyData: [],
  costHeatmapData: [],
  executions: [],
  currentTraceId: null,
  spans: [],
  selectedSpanId: null,
  rootError: null,
  memoryData: [],
  gpuData: [],
  bandwidthData: [],
  distributedTraceId: null,
  distributedSpans: [],
  replayRequests: [],
  logs: [],
  selectedLogId: null,
  runtimeMetrics: [],
  radarMetrics: [],
  incidents: [],
  selectedIncidentId: null,
  anomalies: [],
  selectedAnomalyId: null,
  currentCost: 0,
  predictedCost: 0,
  costTrend: "stable" as const,
  costBasis: "",
  costProjections: [],
  resourceForecastData: [],
  visiblePanels: new Set(["telemetry", "latency", "errors"]),
}

// ============================================================================
// Store
// ============================================================================

export const useObservabilityStore = create<ObservabilityState>()(
  immer((set) => ({
    ...initialState,

    setTimeRange: (range) => set((state) => { state.timeRange = range }),

    setMetrics: (metrics) => set((state) => { state.metrics = metrics }),
    setTokenUsage: (usage) => set((state) => { state.tokenUsage = usage }),
    updateMetric: (id, updates) => set((state) => {
      const metric = state.metrics.find((m) => m.id === id)
      if (metric) Object.assign(metric, updates)
    }),

    setLatencyData: (data) => set((state) => { state.latencyData = data }),
    setCostHeatmapData: (data) => set((state) => { state.costHeatmapData = data }),
    setExecutions: (executions) => set((state) => { state.executions = executions }),

    setCurrentTrace: (traceId, spans) => set((state) => {
      state.currentTraceId = traceId
      state.spans = spans
    }),
    setSelectedSpanId: (id) => set((state) => { state.selectedSpanId = id }),

    setRootError: (error) => set((state) => { state.rootError = error }),

    setMemoryData: (data) => set((state) => { state.memoryData = data }),
    setGpuData: (data) => set((state) => { state.gpuData = data }),
    setBandwidthData: (data) => set((state) => { state.bandwidthData = data }),

    setDistributedTrace: (traceId, spans) => set((state) => {
      state.distributedTraceId = traceId
      state.distributedSpans = spans
    }),

    setReplayRequests: (requests) => set((state) => { state.replayRequests = requests }),

    addLog: (log) => set((state) => {
      state.logs.push(log)
      if (state.logs.length > 1000) {
        state.logs = state.logs.slice(-1000)
      }
    }),
    clearLogs: () => set((state) => { state.logs = [] }),
    setSelectedLogId: (id) => set((state) => { state.selectedLogId = id }),

    setRuntimeMetrics: (runtimes) => set((state) => { state.runtimeMetrics = runtimes }),
    setRadarMetrics: (metrics) => set((state) => { state.radarMetrics = metrics }),

    setIncidents: (incidents) => set((state) => { state.incidents = incidents }),
    setSelectedIncidentId: (id) => set((state) => { state.selectedIncidentId = id }),

    setAnomalies: (anomalies) => set((state) => { state.anomalies = anomalies }),
    setSelectedAnomalyId: (id) => set((state) => { state.selectedAnomalyId = id }),

    setCostPrediction: ({ current, predicted, trend, basis, projections }) => set((state) => {
      state.currentCost = current
      state.predictedCost = predicted
      state.costTrend = trend
      state.costBasis = basis
      state.costProjections = projections
    }),

    setResourceForecastData: (data) => set((state) => { state.resourceForecastData = data }),

    togglePanel: (panel) => set((state) => {
      if (state.visiblePanels.has(panel)) {
        state.visiblePanels.delete(panel)
      } else {
        state.visiblePanels.add(panel)
      }
    }),

    reset: () => set(() => ({ ...initialState, visiblePanels: new Set(["telemetry", "latency", "errors"]) })),
  }))
)

// ============================================================================
// Selectors
// ============================================================================

export const selectCriticalMetrics = (metrics: TelemetryMetric[]) =>
  metrics.filter((m) => m.status === "error" || m.status === "warning")

export const selectRecentLogs = (logs: LogEntry[], count = 100) =>
  logs.slice(-count)

export const selectActiveIncidents = (incidents: Incident[]) =>
  incidents.filter((i) => i.status === "active")
